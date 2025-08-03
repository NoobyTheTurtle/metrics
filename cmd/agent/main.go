package main

import (
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

	if err := app.StartAgent(); err != nil {
		log.Fatal(err)
	}
}
