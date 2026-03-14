package main

import (
	"fmt"
	"net/http"
	"os"

	"sherlock/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("http://127.0.0.1:%s\n", port)

	if err := http.ListenAndServe(":"+port, server.New()); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
