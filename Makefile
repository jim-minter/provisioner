all: stage2 netboot hack/image-builder/images/capi/output bootstrap-imageserver libvirt

stage2:
	mkdir -p pkg/tftp/assets/amd64
	docker build -t builder:latest hack/stage2
	docker run --rm --mount type=bind,src=hack/stage2/initramfs-tools,dst=/etc/initramfs-tools --mount type=bind,src=pkg/tftp/assets/amd64,dst=/root/output builder:latest bash -c 'cp /boot/vmlinuz /root/output; mkinitramfs -o /root/output/initrd.img $$(basename /lib/modules/*)'

# TODO: hard-coded container registry
netboot: stage2
	go generate ./...
	CGO_ENABLED=0 go build ./cmd/netboot
	docker build -t $(USER).azurecr.io/netboot:latest -f Dockerfile.netboot .
	@docker login -u 00000000-0000-0000-0000-000000000000 -p $(shell az acr login -n $(USER) --expose-token --query accessToken -o tsv 2>/dev/null) $(USER).azurecr.io
	docker push $(USER).azurecr.io/netboot:latest

netboot-deploy: netboot
	kubectl apply -f manifests
	kubectl create configmap -n netboot netboot --from-file=provisioner.yaml --dry-run=client -o yaml | kubectl apply -f -
	go run ./cmd/config <netboot.yaml | kubectl apply -f -
	kubectl delete pod -n netboot -l app=netboot
	kubectl wait --for jsonpath=status.readyReplicas=1 -n netboot replicaset/netboot

# TODO: include netboot, flannel (for now) etc. in diskimage
hack/image-builder/images/capi/output:
	[ -e hack/image-builder ] || git clone -b v0.1.46 --depth=1 https://github.com/kubernetes-sigs/image-builder hack/image-builder
	$(MAKE) -C hack/image-builder build-raw-ubuntu-2404

bootstrap-imageserver: hack/image-builder/images/capi/output
	go build ./cmd/bootstrap-imageserver

libvirt: hack/image-builder/images/capi/output
	$(MAKE) -C hack/development/libvirt

.PHONY: all bootstrap-imageserver stage2 netboot libvirt
