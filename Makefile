all: login netboot

login:
	@docker login -u 00000000-0000-0000-0000-000000000000 -p $(shell az acr login -n $(USER) --expose-token --query accessToken -o tsv 2>/dev/null) $(USER).azurecr.io

netboot:
	go generate ./...
	CGO_ENABLED=0 go build ./cmd/netboot
	docker build -t $(USER).azurecr.io/netboot:latest -f Dockerfile.netboot .
	docker push $(USER).azurecr.io/netboot:latest
	kubectl apply -f manifests
	envsubst <netboot.yaml | kubectl apply -f -
	kubectl rollout restart deployment/netboot

.PHONY: all login netboot
