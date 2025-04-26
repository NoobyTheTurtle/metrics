package main

import (
	"context"
	"log"

	"github.com/NoobyTheTurtle/metrics/internal/app"
)

func main() {
	ctx := context.Background()
	if err := app.StartServer(ctx); err != nil {
		log.Fatal(err)
	}
}
