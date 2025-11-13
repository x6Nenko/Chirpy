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
	ServeMux.Handle("/app/", http.StripPrefix("/app/", fs))

	ServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	// Start the server
	log.Printf("Serving on port: 8080\n")
	log.Fatal(server.ListenAndServe())
}