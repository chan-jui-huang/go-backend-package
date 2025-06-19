.PHONY: test

SHELL:=/bin/bash

linter:
	go tool golangci-lint run ./...

test:
	$(eval args?=./...)
	go test ${args}

benchmark:
	$(eval args?=./...)
	go test -bench=. -run=none -benchmem ${args}
