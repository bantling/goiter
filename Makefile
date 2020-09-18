reviveConfigFile = revive.toml
reviceTestConfigFile = revive_test.toml

.PHONY: all
all: check-go check-path install-tools compile lint format test

.PHONY: check-go
check-go:
	which go > /dev/null || { \
		echo "go is not installed"; \
		exit 1; \
	}

.PHONY: check-path
check-path:
	echo $$PATH | grep -q "`go env GOPATH`/bin" || { \
		echo "`go env GOPATH`/bin is not in PATH"; \
		exit 1; \
	}

.PHONY: install-tools
install-tools:
	which revive 2> /dev/null || go get -u github.com/mgechev/revive

.PHONY: compile
compile:
	modOpt=""; \
	[ -n "$(mod)" ] && modOpt="-mod=$(mod)"; \
	go build "$$modOpt" ./...

.PHONY: lint
lint:
	go vet ./... || exit $$?; \
	reviveConfigOpt=""; \
	[ -f "$(reviveConfigFile)" ] && reviveConfigOpt="-config $(reviveConfigFile)"; \
	revive $$reviveConfigOpt $$(go list -f '{{.GoFiles}}' | tr -d '[]') || exit $$?; \
	reviveTestConfigOpt="$$reviveConfigOpt"; \
	[ -f "$(reviceTestConfigFile)" ] && reviveTestConfigOpt="-config $(reviceTestConfigFile)"; \
	revive $$reviveTestConfigOpt $$(go list -f '{{.TestGoFiles}}' | tr -d '[]')

.PHONY: format
format:
	gofmt -s -w $$(go list -f '{{.GoFiles}} {{.TestGoFiles}}' | tr -d '[]')

.PHONY: test
test:
	modOpt=""; \
	[ -n "$(mod)" ] && modOpt="-mod=$(mod)"; \
	testOpt=""; \
	[ -n "$(run)" ] && testOpt="-run $(run)"; \
	go test "$$modOpt" -v -count=1 $$testOpt ./...
