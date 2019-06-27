#!/bin/sh
REGISTRY=aneeshkp
IMAGE=barometer-operator
TAG=2.0.0

if [[ -z ${CI} ]]; then
	./hack/go-test.sh
	operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o build/_output/bin/barometer-operator github.com/aneeshkp/barometer-operator/cmd/manager
fi