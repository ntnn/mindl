// Package main is the entry point for the mindl CLI.
package main

import (
	"context"
	"log"
	"os"

	"github.com/ntnn/mindl/cmd"
)

func main() {
	if err := cmd.Main.Run(context.Background(), os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
