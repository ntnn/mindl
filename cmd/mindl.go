package cmd

import (
	"github.com/ntnn/mindl/pkg/simplcli"
)

const sumDBPath = "mindl.sum"

// Main dispatches CLI commands.
var Main = simplcli.SimplCLI{
	SubCmds: map[string]simplcli.SubCmd{
		"debug":    {Debug.Run, "debugging commands"},
		"download": {Download, "download a tool"},
	},
}
