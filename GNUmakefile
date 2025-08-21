default: fmt lint install generate

.PHONY: fmt lint test testacc build install generate pushfake

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	go fmt ./...
	TF_ACC=1 go test -v -cover -timeout 120m ./...

pushfake:
	docker build -t nullstone/tf-provider-test:v1 -f fake.Dockerfile .
	docker push nullstone/tf-provider-test:v1
