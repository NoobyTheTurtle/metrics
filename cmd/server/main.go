package main

import (
	"github.com/NoobyTheTurtle/metrics/internal/apps"
	"log"
)

func main() {
	if err := apps.StartServer(); err != nil {
		log.Fatal(err)
	}
}
