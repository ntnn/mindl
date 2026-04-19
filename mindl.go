// Package main is the entry point for the mindl CLI.
package main

import (
	"github.com/ntnn/mindl/cmd"
	"github.com/ntnn/mindl/pkg/simplcli"
)

func main() {
	simplcli.Entrypoint(cmd.Main)
}
