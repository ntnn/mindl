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
	"goreleaser": {
		URL:       "https://github.com/goreleaser/goreleaser/releases/download/v{{.Version}}/goreleaser_{{.OS | title}}_{{.Arch | x86_64}}.{{.OSArchive}}",
		InArchive: "goreleaser{{.Exe}}",
	},
	"kcp": {
		URL:       "https://github.com/kcp-dev/kcp/releases/download/v{{.Version}}/kcp_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
		InArchive: "bin/kcp",
	},
	"kubectl": {
		URL:       "https://dl.k8s.io/v{{.Version}}/kubernetes-client-{{.OS}}-{{.Arch}}.tar.gz",
		InArchive: "kubernetes/client/bin/kubectl{{.Exe}}",
	},
	"setup-envtest": {
		URL:       "https://github.com/kubernetes-sigs/controller-runtime/releases/download/v{{.Version}}/setup-envtest-{{.OS}}-{{.Arch}}{{.Exe}}",
		InArchive: "",
	},
}
