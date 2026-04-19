GO ?= go
WHAT ?= ./...

TOOLS_DIR = hack/tools

GOLANGCI_LINT_VER := 2.11.0
GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(TOOLS_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER)

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

$(GOLANGCI_LINT):
	mkdir -p $(TOOLS_DIR)
	$(GO) run . download -common -out $@ -tool golangci-lint -version $(GOLANGCI_LINT_VER)
