GO ?= go

check: lint test

WHAT ?= ./...

.PHONY: lint
lint:
	$(GO) vet $(WHAT)

.PHONY: test
test:
	$(GO) test -race -v $(WHAT)
