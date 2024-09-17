package main

import (
	"os"

	"github.com/mikemykhaylov/asu-dining-bot/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
