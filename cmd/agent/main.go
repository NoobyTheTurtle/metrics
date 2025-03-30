package main

import (
	"log"

	"github.com/NoobyTheTurtle/metrics/internal/apps"
)

func main() {
	if err := apps.StartAgent(); err != nil {
		log.Fatal(err)
	}
}
