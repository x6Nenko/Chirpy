package main

import (
	"log"
	"net/http"
)

func main() {
	// Creating a new ServeMux
	ServeMux := http.NewServeMux()

	// Creating a server struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: ServeMux,
	}

	fs := http.FileServer(http.Dir("."))
	ServeMux.Handle("/", fs)

	// Start the server
	log.Printf("Serving on port: 8080\n")
	log.Fatal(server.ListenAndServe())
}