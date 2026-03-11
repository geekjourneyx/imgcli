CLI_NAME := imgcli
VERSION := 1.0.0
BIN := bin/$(CLI_NAME)
LDFLAGS := -X github.com/geekjourneyx/imgcli/pkg/version.Version=$(VERSION)

.PHONY: build fmt vet lint test release-check real-smoke version

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN) .

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	CGO_ENABLED=1 go test -count=1 ./...

release-check:
	./scripts/release_check.sh

real-smoke:
	./scripts/real_smoke.sh

version:
	@echo $(VERSION)
