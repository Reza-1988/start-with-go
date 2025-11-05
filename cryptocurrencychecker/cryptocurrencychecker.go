package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/sayHelloWorld", handleHelloWorld)
	http.ListenAndServe(":8080", nil)
}

func handleHelloWorld(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Hello, World",
		"status":  "success",
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Contant-Type", "application/json")
	w.Write(jsonResponse)
}
