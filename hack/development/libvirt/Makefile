all: network laptop node2

laptop: vmlinuz
	virsh destroy laptop || true
	virsh undefine laptop || true
	qemu-img convert -O qcow2 ../../image-builder/images/capi/output/ubuntu-2404-kube-v1.32.4/ubuntu-2404-kube-v1.32.4 laptop.qcow2
	qemu-img resize laptop.qcow2 25G
	go run ../../../cmd/config <laptop.xml | virsh define /dev/stdin

vmlinuz:
	$(eval loop=$(shell sudo losetup --show -f ../../image-builder/images/capi/output/ubuntu-2404-kube-v1.32.4/ubuntu-2404-kube-v1.32.4))
	sudo partprobe $(loop)
	sudo debugfs $(loop)p1 -R "cat /boot/vmlinuz-6.8.0-85-generic" >vmlinuz
	sudo debugfs $(loop)p1 -R "cat /boot/initrd.img-6.8.0-85-generic" >initrd.img
	sudo losetup -d $(loop)

node2:
	virsh destroy node2 || true
	virsh undefine node2 || true
	qemu-img create -f qcow2 node2.qcow2 25G
	go run ../../../cmd/config <node2.xml | virsh define /dev/stdin

network:
	virsh net-destroy apollo || true
	virsh net-undefine apollo || true
	go run ../../../cmd/config <network/apollo.xml | virsh net-define /dev/stdin
	virsh net-start apollo
	virsh net-autostart apollo

network-disconnected:
	virsh net-destroy apollo || true
	virsh net-undefine apollo || true
	go run ../../../cmd/config <network/apollo-disconnected.xml | virsh net-define /dev/stdin
	virsh net-start apollo
	virsh net-autostart apollo

.PHONY: all laptop network network-disconnected
