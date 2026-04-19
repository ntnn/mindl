package mindl

// CommonTool records the templates used for a commonly used tool.
type CommonTool struct {
	URL       string
	InArchive string
}

// CommonTools keys commonly used tools to their templates.
var CommonTools = map[string]CommonTool{
	"golangci-lint": {
		URL:       "https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}.{{.OSArchive}}",
		InArchive: "golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}/golangci-lint{{.Exe}}",
	},
	"mkcert": {
		URL:       "https://github.com/FiloSottile/mkcert/releases/download/v{{.Version}}/mkcert-v{{.Version}}-{{.OS}}-{{.Arch}}{{.Exe}}",
		InArchive: "",
	},
}
