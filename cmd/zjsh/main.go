package main

import (
	"os"

	"github.com/tassis/zjsh/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
