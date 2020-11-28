package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// MessageDTO ...
type MessageDTO struct {
	Description string  `json:"description"`
	Headline    string  `json:"headline"`
	Rate        float64 `json:"rate"`
}

// MessageDTO ...
type ResultDTO struct {
	Class       string  `json:"class"`
	Probability float64 `json:"probability"`
}

type Configuration struct {
	Category string `json:"category"`
	API      string `json:"api"`
}

func getConfiguration() []Configuration {
	var configuration []Configuration
	file, _ := ioutil.ReadFile("./configuration.json")
	_ = json.Unmarshal([]byte(file), &configuration)
	return configuration
}

func sendData(data MessageDTO, url, category string, results *map[string]float64, mutex *sync.Mutex, wg *sync.WaitGroup) {
	body, _ := json.Marshal(data)

	buff := bytes.NewBuffer(body)
	resp, err := http.Post(url, "application/json; charset=UTF-8", buff)

	if err != nil {
		log.Printf("An Error Occured %v", err)
		wg.Done()
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		wg.Done()
		return
	}

	var result ResultDTO
	if json.Unmarshal(response, &result); err == nil {
		mutex.Lock()
		fmt.Println(result)
		(*results)[category] = result.Probability
		mutex.Unlock()
	}
	wg.Done()
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	body, err := json.Marshal(data)
	buff := bytes.NewBuffer(body)

	if err != nil {
		log.Printf("Failed to encode a JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buff.Bytes())
	if err != nil {
		log.Printf("Failed to write the response body: %v", err)
		return
	}
}

// Classify ...
func Classify(w http.ResponseWriter, req *http.Request) {
	configuration := getConfiguration()

	var data MessageDTO
	body, _ := ioutil.ReadAll(req.Body)
	json.Unmarshal(body, &data)

	results := make(map[string]float64)
	for _, configuration := range configuration {
		results[configuration.Category] = 0.0
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, configuration := range configuration {
		wg.Add(1)
		go sendData(data, configuration.API, configuration.Category, &results, &mutex, &wg)
	}
	wg.Wait()

	maxResult := configuration[0].Category
	for key, value := range results {
		if results[maxResult] < value {
			maxResult = key
		}
	}
	sendJSONResponse(w, map[string]string{
		"category": maxResult,
	})
	return
}

func HeartBeat(w http.ResponseWriter, req *http.Request) {
	sendJSONResponse(w, map[string]string{
		"message": "ok",
	})
	return
}
