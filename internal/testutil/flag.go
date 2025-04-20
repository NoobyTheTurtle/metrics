package testutil

import (
	"flag"
	"os"
)

func ResetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}
