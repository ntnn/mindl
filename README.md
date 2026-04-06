# mindl

mindl is a downloader that only uses Go's stdlib to have a no-dependency
tool to download pre-built binaries from releases.

## Usage

mindl is intended to be used as a [Go tool dependency](https://go.dev/doc/modules/managing-dependencies#tools)
in Makefiles to bootstrap project-local development tools reproducibly.

Add it as a tool dependency:

```sh
go get -tool github.com/ntnn/mindl@latest
```

Then use it in a Makefile:

```makefile
GOLANGCI_LINT_VER := 2.10.0
GOLANGCI_LINT := hack/tools/golangci-lint-$(GOLANGCI_LINT_VER)

$(GOLANGCI_LINT):
	mkdir -p hack/tools
	go tool github.com/ntnn/mindl download -common -out $@ \
		-url 'https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}.{{.OSArchive}}' \
		-inarchive 'golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}/golangci-lint{{.Exe}}' \
		-version $(GOLANGCI_LINT_VER)

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...
```

mindl uses a sum file `mindl.sum` to track the hash of the used binaries.

To update these hashes run the tool with `MINDL_UPDATE=true`, e.g.:

```sh
make lint MINDL_UPDATE=true
```

This will download the tool and write the hashes to `mindl.sum` - this
file is meant to be tracked in version control. The `-common` flag for
the download subcommand lets mindl automatically add hashes for common
OS/arch combinations.

To update simply update the version of the tool being used and rerun the
command with `MINDL_UPDATE=true`. mindl will automatically update all
OS/arch combinations for this tool registered in the sum file.

#### -tool

mindl also records the URL and InArchive templates of commonly used tools.
The full list is recorded in [`mindl.CommonTools`](https://pkg.go.dev/github.com/ntnn/mindl@main/pkg/mindl#CommonTools).

When passed to the download subcommand for the `-tool` flag both `-url`
and `-inarchive` default to these values, but can still be overriden if
specified.

```makefile
GOLANGCI_LINT_VER := 2.10.0
GOLANGCI_LINT := hack/tools/golangci-lint-$(GOLANGCI_LINT_VER)

$(GOLANGCI_LINT):
	mkdir -p hack/tools
	go run github.com/ntnn/mindl download -common -out $@ -tool golangci-lint -version $(GOLANGCI_LINT_VER)

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...
```

## `mindl.sum`

`mindl.sum` is a CSV file that stores the SHA-512 hash for each URL,
in-archive path, OS and architecture combination. Similarly to `go.sum`
and similar files it must be committed to version control. mindl
verifies future downloads against the hashes recorded in this file.

## Commands

### `download`

| Flag | Description |
|------|-------------|
| `-url` | URL for the archive to download, templated |
| `-inarchive` | Path of the file to extract from the archive, templated |
| `-out` | Destination path for the extracted binary |
| `-version` | Version string substituted into templates |
| `-common` | Also update hashes for common OS/Arch combinations |
| `-tool` | Default `-url` and `-inarchive` to the values of this common tool |

For the targets added with `-common` check [`mindl.CommonTargets`](https://pkg.go.dev/github.com/ntnn/mindl@main/pkg/mindl#pkg-variables).
For the tools available for `-tool` check [`mindl.CommonTools`](https://pkg.go.dev/github.com/ntnn/mindl@main/pkg/mindl#CommonTools).

#### Templating

`-url` and `-inarchive` are templated. The templates get [`mindl.TemplateData`](https://pkg.go.dev/github.com/ntnn/mindl@main/pkg/mindl#TemplateData) as context.
Additionally convenience functions from [`mindl.FuncMap`](https://pkg.go.dev/github.com/ntnn/mindl@main/pkg/mindl#FuncMap) are available in the templates.

## Downloading for specific OS/Arch

When updating the hash for a specific OS/Arch combination the OS and
Arch can be overridden by setting `MINDL_OS` and `MINDL_ARCH` to the
desired values.

## Similar projects

There are several similar projects that inspired writing this one:

- [uget](https://codeberg.org/xrstf/uget)
- [aqua](https://github.com/aquaproj/aqua)
- [bin](https://github.com/marcosnils/bin)
- [bingo](https://github.com/bwplotka/bingo)

My goal was to be able to use pre-built binaries from release artifacts
with a tool that needs no big setup on the client side.

By mindl only using stdlib it can easily be used as `go tool` without
polluting the dependencies and doesn't incur a maintenance debt (until
the libs used from stdlib are deprecated).
