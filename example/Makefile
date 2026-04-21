GO ?= go
WHAT ?= ./...

TOOLS_DIR = hack/tools

GOLANGCI_LINT_VER := 2.10.0
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint-$(GOLANGCI_LINT_VER)

MKCERT_VER := 1.4.4
MKCERT := $(TOOLS_DIR)/mkcert-$(MKCERT_VER)

GORELEASER_VER := 2.15.3
GORELEASER := $(TOOLS_DIR)/goreleaser-$(GORELEASER_VER)


KUBECTL_VER := 1.35.1
KUBECTL := $(TOOLS_DIR)/kubectl-$(KUBECTL_VER)

SETUP_ENVTEST_VER := 0.23.3
SETUP_ENVTEST := $(TOOLS_DIR)/setup-envtest-$(SETUP_ENVTEST_VER)

check: lint test
tools: $(GOLANGCI_LINT) $(MKCERT) $(GORELEASER) $(KUBECTL) $(SETUP_ENVTEST)

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_FLAGS) $(WHAT)

.PHONY: lint-fix
lint-fix: override GOLANGCI_LINT_FLAGS := $(GOLANGCI_LINT_FLAGS) --fix
lint-fix: lint

.PHONY: test
test:
	$(GO) test -race -v $(WHAT)

$(GOLANGCI_LINT):
	mkdir -p $(TOOLS_DIR)
	$(GO) tool github.com/ntnn/mindl download -common -out $@ \
		-url 'https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}.{{.OSArchive}}' \
		-inarchive golangci-lint-{{.Version}}-{{.OS}}-{{.Arch}}/golangci-lint{{.Exe}} \
		-version $(GOLANGCI_LINT_VER)

$(MKCERT):
	mkdir -p $(TOOLS_DIR)
	$(GO) tool github.com/ntnn/mindl download -common -out $@ -tool mkcert -version $(MKCERT_VER)

$(GORELEASER):
	mkdir -p $(TOOLS_DIR)
	$(GO) tool github.com/ntnn/mindl download -common -out $@ -tool goreleaser -version $(GORELEASER_VER)

$(KUBECTL):
	mkdir -p $(TOOLS_DIR)
	$(GO) tool github.com/ntnn/mindl download -common -out $@ -tool kubectl -version $(KUBECTL_VER)

$(SETUP_ENVTEST):
	mkdir -p $(TOOLS_DIR)
	$(GO) tool github.com/ntnn/mindl download -common -out $@ -tool setup-envtest -version $(SETUP_ENVTEST_VER)
