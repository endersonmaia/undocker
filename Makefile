GODEPS = $(shell git ls-files '*.go' go.mod go.sum)
GOBIN = $(shell go env GOPATH)/bin/

GOOSARCHS = linux/amd64 \
			linux/arm64 \
			darwin/amd64 \
			darwin/arm64 \
			windows/amd64/.exe

define undockertarget
TARGETS += undocker-$(strip $(1))-$(strip $(2))$(firstword $(3))
undocker-$(strip $(1))-$(strip $(2))$(firstword $(3)): $(GODEPS)
	CGO_ENABLED=0 GOOS=$(strip $(1)) GOARCH=$(strip $(2)) go build -o $$@
endef

$(foreach goosarch,$(GOOSARCHS), \
	$(eval $(call undockertarget,\
		$(word 1,$(subst /, ,$(goosarch))),\
		$(word 2,$(subst /, ,$(goosarch))),\
		$(word 3,$(subst /, ,$(goosarch))),\
)))

.PHONY: all
all: $(TARGETS) coverage.html

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
	rm -f coverage.html $(TARGETS)
