#!/bin/sh

if [[ -z ${CI} ]]; then
    ./hack/go-vet.sh
    ./hack/go-fmt.sh
    ./hack/catalog-source.sh
fi

#local test
GO111MODULE=on go test `go list ./test/... | grep -v e2e`