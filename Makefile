
.PHONY: precommit
precommit: build lint test

.PHONY: build
build:
	go build -o bin/home24 cmd/api/main.go

.PHONY: run-api
run-api:
	go run cmd/api/main.go

.PHONY: run-website
run-website: generate
	go run cmd/website/main.go	

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

.PHONY: generate
generate:
	go install github.com/a-h/templ/cmd/templ@latest && \
	templ generate 