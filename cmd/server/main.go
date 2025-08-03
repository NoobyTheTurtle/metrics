package main

import (
	"context"
	"log"

	"github.com/NoobyTheTurtle/metrics/internal/app"
	"github.com/NoobyTheTurtle/metrics/internal/util"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	util.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	ctx := context.Background()

	if err := app.StartServer(ctx); err != nil {
		log.Fatal(err)
	}
}
