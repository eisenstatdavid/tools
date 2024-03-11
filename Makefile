fmt:
	scripts/fmt
.PHONY: fmt

install:
	scripts/install
.PHONY: install

lint:
	go vet ./...
	errcheck ./...
	shellcheck scripts/*
.PHONY: lint

spell:
	find . -name '*.go' -exec cat {} + | spell-check
.PHONY: spell
