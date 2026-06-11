GO_CACHE := $(CURDIR)/.gocache/go-build

.PHONY: test
test:
	GOCACHE=$(GO_CACHE) go test ./...
