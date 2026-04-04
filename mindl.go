package main

import (
	"context"
	"log"
	"os"

	"github.com/ntnn/mindl/cmd"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal(cmd.ErrNoArgs)
	}
	if err := cmd.Main(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
