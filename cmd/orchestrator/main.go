package main

import (
	"kiskislaya/ConcurrencyCalc/internal/orchestrator"
	"log"
	"net/http"
)

func main() {
	orchestrator.RegisterHandlers()
	log.Println("Orchestrator running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
