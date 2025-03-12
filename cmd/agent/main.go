package main

import (
	"kiskislaya/ConcurrencyCalc/internal/agent"
	"log"
	"os"
	"strconv"
)

func main() {
	computingPower, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if computingPower == 0 {
		computingPower = 2
	}

	for i := 0; i < computingPower; i++ {
		go agent.Worker()
	}

	log.Println("Agent running...")
	select {}
}
