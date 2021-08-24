GODEPS = $(shell git ls-files '*.go' go.mod go.sum)
GOBIN = $(shell go env GOPATH)/bin/

GOOSARCHS = $(sort \
				linux/amd64 \
				linux/arm64 \
				darwin/amd64 \
				darwin/arm64 \
				windows/amd64/.exe \
			)

VERSION = $(shell git describe --dirty)
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

define undockertarget
UNDOCKERS += undocker-$(1)-$(2)-$(VERSION)-$(firstword $(3))
undocker-$(1)-$(2)-$(VERSION)$(firstword $(3)): $(GODEPS)
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build $(LDFLAGS) -o $$@
endef

$(foreach goosarch,$(GOOSARCHS),\
	$(eval $(call undockertarget,$(word 1,$(subst /, ,$(goosarch))),$(word 2,$(subst /, ,$(goosarch))),$(word 3,$(subst /, ,$(goosarch))))))

.PHONY: all
all: $(UNDOCKERS) coverage.html

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

sha256sum.txt: $(UNDOCKERS)
	sha256sum $(UNDOCKERS) > $@

sha256sum.txt.asc: sha256sum.txt
	gpg --clearsign $<

.PHONY: clean
clean:
	rm -f $(UNDOCKERS) coverage.html sha256sum.txt sha256sum.txt.asc
