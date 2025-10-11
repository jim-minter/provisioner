# TODO: hard-coded container registry
netboot: stage2 login
	go generate ./...
	CGO_ENABLED=0 go build ./cmd/netboot
	docker build -t $(USER).azurecr.io/netboot:latest -f Dockerfile.netboot .
	docker push $(USER).azurecr.io/netboot:latest
	kubectl apply -f manifests
	envsubst <netboot.yaml | kubectl apply -f -
	kubectl delete pod -n netboot -l app=netboot
	kubectl wait --for jsonpath=status.readyReplicas=1 -n netboot replicaset/netboot

# TODO: include netboot, flannel (for now) etc. in diskimage
diskimage:
	[ -e hack/image-builder ] || git clone -b v0.1.46 --depth=1 https://github.com/kubernetes-sigs/image-builder hack
	$(MAKE) -C hack/image-builder build-raw-ubuntu-2404
	echo images are in hack/image-builder/images/capi/output

stage2:
	mkdir -p pkg/tftp/assets/amd64
	docker build -t builder:latest hack/stage2
	docker run --rm --mount type=bind,src=hack/stage2/initramfs-tools,dst=/etc/initramfs-tools --mount type=bind,src=pkg/tftp/assets/amd64,dst=/root/output builder:latest bash -c 'cp /boot/vmlinuz /root/output; mkinitramfs -o /root/output/initrd.img $$(basename /lib/modules/*)'

login:
	@docker login -u 00000000-0000-0000-0000-000000000000 -p $(shell az acr login -n $(USER) --expose-token --query accessToken -o tsv 2>/dev/null) $(USER).azurecr.io

.PHONY: diskimage login stage2 netboot
