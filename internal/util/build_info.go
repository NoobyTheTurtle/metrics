package util

import (
	"fmt"
	"os"
)

type BuildInfo struct {
	Version string
	Date    string
	Commit  string
}

func PrintBuildInfo(version, date, commit string) {
	info := BuildInfo{
		Version: version,
		Date:    date,
		Commit:  commit,
	}

	if info.Version == "" {
		info.Version = "N/A"
	}
	if info.Date == "" {
		info.Date = "N/A"
	}
	if info.Commit == "" {
		info.Commit = "N/A"
	}

	fmt.Fprintf(os.Stdout, "Build version: %s\n", info.Version)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", info.Date)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", info.Commit)
}
