default: build

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

test:
	go test ./... -v $(TESTARGS) -timeout 120m

testacc:
	TF_ACC=1 go test ./internal/... -v $(TESTARGS) -timeout 120m

generate:
	go generate ./...

.PHONY: build install lint test testacc generate
