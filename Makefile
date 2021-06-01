GODEPS = $(shell git ls-files '*.go' go.mod go.sum)
GOBIN = $(shell go env GOPATH)/bin/

.PHONY: all
all: undocker coverage.html

undocker: $(GODEPS)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: lint
lint: vet staticcheck

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	$(GOBIN)staticcheck -f stylish ./...

.INTERMEDIATE: coverage.out
coverage.out: $(GODEPS)
	go test -coverprofile $@ ./...

coverage.html: coverage.out
	go tool cover -html=$< -o $@

.PHONY: clean
clean:
	rm -f coverage.html undocker
