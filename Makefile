IMAGE := "ghcr.io/martinheinz/some-check"
TAG := "devel"

.PHONY: test build build-dev push

build:
	docker build --no-cache --pull -t ${IMAGE}:${TAG} -f Dockerfile .
build-dev:
	docker build -t ${IMAGE}:dev -f Dockerfile .
test: build-dev push
	kubectl delete -f check.yaml
	kubectl apply -f check.yaml
push: build
	docker push ${IMAGE}:${TAG}
