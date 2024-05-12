
.PHONY: precommit
precommit: build-api build-website lint test test-race


# Build

.PHONY: build-api
build-api:
	go build -o bin/api cmd/api/main.go

.PHONY: build-website
build-website:
	go build -o bin/website cmd/website/main.go


# Run

.PHONY: run-api
run-api:
	go run cmd/api/main.go

.PHONY: run-website
run-website: generate
	go run cmd/website/main.go	


# Test

.PHONY: test
test:
	go test ./...

.PHONY: test-race
test-race:
	go test -race ./...

.PHONY: test-ginkgo
test-ginkgo:
	go install github.com/onsi/ginkgo/v2/ginkgo &&\
	ginkgo ./cmd/acceptance_test/

.PHONY: test-bench
test-bench:
	go test -bench=. ./...


# Lint

.PHONY: lint
lint:
	golangci-lint run


# Generate

.PHONY: generate
generate:
	go install github.com/a-h/templ/cmd/templ@latest && \
	templ generate



# Docker

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build -t html-analyzer .

.PHONY: docker-run-api
docker-run-api:
	docker run -p 8080:8080 -t html-analyzer

.PHONY: docker-run-website
docker-run-website:
	docker run -p 3000:3000 --entrypoint "/website" -t html-analyzer

.PHONY: docker-compose-up
docker-compose-up:
	docker compose up --build
