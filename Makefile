SCRIPTS = $(shell awk '/#!\/bin\/(ba)?sh/&&FNR==1{print FILENAME}' $(shell git ls-files))
GODEPS = $(shell git ls-files '*.go' go.mod go.sum)
GOBIN = $(shell go env GOPATH)/bin/
GOOSARCHS = $(sort darwin/amd64 linux/amd64)

VSN ?= $(shell git describe --dirty)
VSNHASH = $(shell git rev-parse --verify HEAD)
LDFLAGS = -ldflags "-X main.Version=$(VSN) -X main.VersionHash=$(VSNHASH)"

undocker: ## builds binary for the current architecture
	CGO_ENABLED=0 go build $(LDFLAGS) -o $@

.PHONY: test
test:
	go test -race -cover ./...

define undockertarget
UNDOCKERS += undocker-$(1)-$(2)-$(VSN)
undocker-$(1)-$(2)-$(VSN): $(GODEPS)
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build $(LDFLAGS) -o $$@
endef

$(foreach goosarch,$(GOOSARCHS),\
	$(eval $(call undockertarget,$(word 1,$(subst /, ,$(goosarch))),$(word 2,$(subst /, ,$(goosarch))))))

.PHONY: all
all: $(UNDOCKERS) sha256sum-$(VSN).txt

.PHONY: sha256sum-asc
sha256sum-asc: sha256sum-$(VSN).txt.asc

.PHONY: lint
lint:
	go vet ./...
	$(GOBIN)staticcheck -f stylish ./...
	shellcheck $(SCRIPTS)

.INTERMEDIATE: coverage.out
coverage.out: $(GODEPS)
	go test -coverprofile $@ ./...

coverage.html: coverage.out
	go tool cover -html=$< -o $@

sha256sum-$(VSN).txt: $(UNDOCKERS)
	sha256sum $(UNDOCKERS) > $@

sha256sum-$(VSN).txt.asc: sha256sum-$(VSN).txt
	gpg --clearsign $<

.PHONY: clean
clean:
	rm -f undocker-*-v* coverage.html sha256sum*.txt sha256sum*.txt.asc
