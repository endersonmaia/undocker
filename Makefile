GODEPS = $(shell git ls-files '*.go' go.mod go.sum)
GOBIN = $(shell go env GOPATH)/bin/

GOOSARCHS = $(sort \
			darwin/amd64 \
			darwin/arm64 \
			linux/amd64 \
			linux/arm64 \
			windows/amd64/.exe \
			windows/arm64/.exe)

VSN = $(shell git describe --dirty)
VSNHASH = $(shell git rev-parse --verify HEAD)
LDFLAGS = -ldflags "-X main.Version=$(VSN) -X main.VersionHash=$(VSNHASH)"

.PHONY: test
test:
	go test -race -cover ./...

define undockertarget
UNDOCKERS += undocker-$(1)-$(2)-$(VSN)$(firstword $(3))
undocker-$(1)-$(2)-$(VSN)$(firstword $(3)): $(GODEPS)
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build $(LDFLAGS) -o $$@
endef

$(foreach goosarch,$(GOOSARCHS),\
	$(eval $(call undockertarget,$(word 1,$(subst /, ,$(goosarch))),$(word 2,$(subst /, ,$(goosarch))),$(word 3,$(subst /, ,$(goosarch))))))

.PHONY: all
all: $(UNDOCKERS)

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
	rm -f undocker-*-v* coverage.html sha256sum.txt sha256sum.txt.asc
