.PHONY: test security security-govulncheck security-osv

SHELL:=/bin/bash

linter:
	go tool golangci-lint run ./...

security: security-govulncheck security-osv

security-govulncheck:
	go tool govulncheck -scan=package -test ./...
	go tool govulncheck -scan=package tool

# Scan the complete selected module graph, including tool dependencies.
security-osv:
	@set -e; \
	scan_dir=$$(mktemp -d); \
	trap 'rm -rf "$$scan_dir"' EXIT; \
	{ \
		printf 'module security-scan\n\nrequire (\n'; \
		go list -m -f '{{if not .Main}}{{if .Replace}}{{if .Replace.Version}}{{.Replace.Path}} {{.Replace.Version}}{{end}}{{else}}{{.Path}} {{.Version}}{{end}}{{end}}' all; \
		printf ')\n'; \
	} > "$$scan_dir/go.mod"; \
	go run github.com/google/osv-scanner/v2/cmd/osv-scanner@v2.7.0 scan source --no-call-analysis=go --lockfile="$$scan_dir/go.mod"

test:
	$(eval args?=./...)
	go test ${args}

benchmark:
	$(eval args?=./...)
	go test -bench=. -run=none -benchmem ${args}
