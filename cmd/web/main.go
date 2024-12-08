package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/view", view)
	mux.HandleFunc("/create", create)
	mux.HandleFunc("/snippetView/", snippetView)
	log.Printf("Listening on ::%v", 80)
	err := http.ListenAndServe(":http", mux)
	if err != nil {
		log.Fatal(err)
	}
}
