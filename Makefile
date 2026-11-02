.PHONY: test security security-go-version security-govulncheck security-osv

SHELL:=/bin/bash

linter:
	go tool golangci-lint run ./...

security: security-go-version security-govulncheck security-osv

# Check all known standard library and toolchain advisories for the active Go release.
security-go-version:
	@set -e; \
	go_version=$$(go env GOVERSION); \
	go_version=$${go_version#go}; \
	if [[ ! "$$go_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$$ ]]; then \
		printf 'Cannot check unsupported Go release version: %s\n' "$$go_version" >&2; \
		exit 1; \
	fi; \
	scan_dir=$$(mktemp -d); \
	trap 'rm -rf "$$scan_dir"' EXIT; \
	printf 'Checking Go %s: standard library and toolchain\n' "$$go_version"; \
	printf '{"results":[{"packages":[{"package":{"ecosystem":"Go","name":"stdlib","version":"%s"}},{"package":{"ecosystem":"Go","name":"toolchain","version":"%s"}}]}]}\n' "$$go_version" "$$go_version" > "$$scan_dir/osv-scanner.json"; \
	go run github.com/google/osv-scanner/v2/cmd/osv-scanner@v2.7.0 scan source --no-call-analysis=go --lockfile="osv-scanner:$$scan_dir/osv-scanner.json"

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
