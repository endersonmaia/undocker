SCRIPTS = $(shell awk '/#!\/bin\/(ba)?sh/&&FNR==1{print FILENAME}' $(shell git ls-files))
GODEPS = $(shell git ls-files '*.go' go.mod go.sum)

VSN ?= $(shell git describe --dirty)
VSNHASH = $(shell git rev-parse --verify HEAD)
LDFLAGS = -ldflags "-X main.Version=$(VSN) -X main.VersionHash=$(VSNHASH)"

undocker: ## builds binary for the current architecture
	go build $(LDFLAGS) -o $@

.PHONY: test
test: coverage.out

.PHONY: lint
lint:
	go vet ./...
	$(shell go env GOPATH)/bin/staticcheck -f stylish ./...
	shellcheck $(SCRIPTS)

.INTERMEDIATE: coverage.out
coverage.out: $(GODEPS)
	go test -race -cover -coverprofile $@ ./...

coverage.html: coverage.out
	go tool cover -html=$< -o $@

.PHONY: clean
clean:
	rm -f undocker-*-v* coverage.html
