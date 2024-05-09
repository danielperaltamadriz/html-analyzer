
.PHONY: precommit
precommit: build lint test

.PHONY: build
build:
	go build -o bin/home24 cmd/api/main.go

.PHONY: run
run:
	go run cmd/api/main.go

.PHONY: test
test:
	go test ./...

.PHONY: test-ginkgo
test-ginkgo:
	ginkgo ./cmd/acceptance_test/

.PHONY: test-bench
test-bench:
	go test -bench=. ./...

.PHONY: lint
lint:
	golangci-lint run