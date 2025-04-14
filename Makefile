REGISTRY ?= gcr.io/k8s-minikube
TAG ?= v0.0.13

.PHONY: push-latest-dev-image
push-latest-dev-image: 
	docker login gcr.io/k8s-minikube
	docker buildx create --name multiarch --bootstrap
	docker buildx build --push --builder multiarch --platform linux/amd64,linux/arm64 -t $(REGISTRY)/triage-party:$(TAG) -t $(REGISTRY)/triage-party:latest .
	docker buildx rm multiarch