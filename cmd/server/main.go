package main

import (
	"github.com/NoobyTheTurtle/metrics/internal/server"
	"log"
)

func main() {
	if err := server.StartServer(); err != nil {
		log.Fatal(err)
	}
}
