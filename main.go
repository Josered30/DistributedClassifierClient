package main

import (
	"log"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)



type ReceiveMessage struct {
	Message string  `json: "message"`
	Rate    float64 `json: "rate"`
}

func makeRouter() *mux.Router {
    router := mux.NewRouter()

	//endpoints
	router.HandleFunc("/heartbeat", HeartBeat).Methods("GET")
	router.HandleFunc("/classify", Classify).Methods("POST")
	return router
}

func main() {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}

	router := makeRouter()
	log.Fatal(http.ListenAndServe(":"+os.Getenv("ADDRESS"), router))
}
