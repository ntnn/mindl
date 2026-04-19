package cmd

import (
	"github.com/ntnn/mindl/pkg/simplcli"
)

const sumDBPath = "mindl.sum"

// Main dispatches CLI commands.
var Main = simplcli.SimplCLI{
	SubCmds: map[string]simplcli.SubCmd{
		"debug":    {Debug.Run, "Debugging commands"},
		"download": {Download, "Download a tool"},
		"common":   {Common.Run, "Show available common tools"},
	},
}
