package main

import (
	"log"

	"github.com/NoobyTheTurtle/metrics/internal/apps"
)

func main() {
	if err := apps.StartServer(); err != nil {
		log.Fatal(err)
	}
}
