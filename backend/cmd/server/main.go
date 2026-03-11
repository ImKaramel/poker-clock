package main

import (
	"log"
	"net/http"

	"backend/internal/api"
)

func main() {
	r := api.NewRouter()

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
