package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("=== STARTING TEST HTTP SERVER ===")
	
	router := mux.NewRouter()
	
	// Test endpoint
	router.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Test endpoint called: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message":"test server works"}`)
	}).Methods("GET")
	
	// Chat endpoint
	router.HandleFunc("/api/v1/chats/{chat_id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		chatID := vars["chat_id"]
		log.Printf("Chat endpoint called: %s %s, chat_id: %s", r.Method, r.URL.Path, chatID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"chat endpoint works", "chat_id":"%s"}`, chatID)
	}).Methods("GET")
	
	log.Println("Starting test server on :8096")
	log.Fatal(http.ListenAndServe(":8096", router))
}