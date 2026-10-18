GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: test
test:
	go test -v -race ./...

.PHONY: lint
lint: $(GOBIN)/golangci-lint
	golangci-lint run

$(GOBIN)/golangci-lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

.PHONY: clean
clean:
	go clean
