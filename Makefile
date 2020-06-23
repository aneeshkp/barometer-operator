
.PHONY: all
all: build


.PHONY: format
format:
	./hack/go-fmt.sh

.PHONY: sdk-generate
sdk-generate:
	operator-sdk generate k8s

tidy: ## Update dependencies
	$(Q)go mod tidy -v

.PHONY: vet
vet:
	./hack/go-vet.sh

.PHONY: test
test:
	./hack/go-test.sh

.PHONY: cluster-test
cluster-test:
	operator-sdk test local "./test/e2e"

.PHONY: build
build:
	./hack/go-build.sh
