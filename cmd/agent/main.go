package main

import (
	"log"

	"github.com/NoobyTheTurtle/metrics/internal/app"
)

func main() {
	if err := app.StartAgent(); err != nil {
		log.Fatal(err)
	}
}
