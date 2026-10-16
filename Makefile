.PHONY: test
test:
	go test -v -race ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: clean
clean:
	go clean
