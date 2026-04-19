GO ?= go
WHAT ?= ./...

TOOLS_DIR = hack/tools

GOLANGCI_LINT_VER := 2.11.0
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint-$(GOLANGCI_LINT_VER)

GORELEASER_VER := 2.15.3
GORELEASER := $(TOOLS_DIR)/goreleaser-$(GORELEASER_VER)
GORELEASER_FLAGS := --clean

check: lint test example

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_FLAGS) $(WHAT)

.PHONY: lint-fix
lint-fix: override GOLANGCI_LINT_FLAGS := $(GOLANGCI_LINT_FLAGS) --fix
lint-fix: lint

.PHONY: test
test:
	$(GO) test -race $(WHAT)

.PHONY: example
example:
	cd example; git clean -fdx . ; make tools

.PHONY: snapshot
snapshot: $(GORELEASER)
	$(GORELEASER) release $(GORELEASER_FLAGS) --snapshot --skip announce,publish

.PHONY: release
release: $(GORELEASER)
	$(GORELEASER) release $(GORELEASER_FLAGS)

$(GOLANGCI_LINT):
	mkdir -p $(TOOLS_DIR)
	$(GO) run . download -common -out $@ -tool golangci-lint -version $(GOLANGCI_LINT_VER)

$(GORELEASER):
	mkdir -p $(TOOLS_DIR)
	$(GO) run . download -common -out $@ -tool goreleaser -version $(GORELEASER_VER)
