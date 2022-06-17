default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	go fmt ./...
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

pushfake:
	docker build -t nullstone/tf-provider-test:v1 -f fake.Dockerfile .
	docker push nullstone/tf-provider-test:v1
