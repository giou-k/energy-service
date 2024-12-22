SHELL = /bin/bash

build:
	go build -ldflags "-X main.build=local" ./api/services
 	# With `main.build=local`, we define the value of build var inside main.go.

run:
	go run api/services/main.go

# ==============================================================================
# Building containers

VERSION := 0.2
BASE_IMAGE_NAME := giou/energy-service
SERVICE_NAME    := energy-api
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)

KIND            := kindest/node:v1.27.3
KIND_CLUSTER    := cluster-europe
NAMESPACE       := ns-energy-europe
APP             := energy-pod

all: service

service:
	echo $(SERVICE_IMAGE)
	docker build \
		-f infra/docker/dockerfile.service \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# K8s

dev-up:
	kind create cluster --image $(KIND) --name $(KIND_CLUSTER) --config infra/k8s/kind-config.yaml
# todo here we will need to wait for the cluster to come up and then we can load other services that depend on that one,f.e metrics

dev-down:
	kind delete cluster --name $(KIND_CLUSTER)

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-load:
	#cd zarf/k8s/dev/sales; kustomize edit set image service-image=$(SERVICE_IMAGE)
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build infra/k8s/ | kubectl apply -f -
	#kubectl apply -f ./infra/k8s/base.yaml

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 # --max-log-requests=6, 5 is the default value, Specify maximum number of concurrent logs to follow when using by a selector.
# currently the above is not working cause of namespace resources: "No resources found in ns-energy-europe namespace."

# logs based on the selector of base.yaml: `k logs -l app=energy --all-containers=true -f --tail=100`

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

dev-update: all dev-load dev-apply run