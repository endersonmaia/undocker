.PHONY: test
test:
	go test -cover ./...

.INTERMEDIATE: coverage.out
coverage.out: $(shell git ls-files '*.go')
	go test -coverprofile $@ ./...

coverage.html: coverage.out
	go tool cover -html=$< -o $@
