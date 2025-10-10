# Provisioner

**WARNING: this is a work-in-progress and unsupported prototype which is not for
production use.  Feedback is welcomed.**

## netboot

`netboot` runs the minimal DHCP/TFTP/HTTP (cloud-init/cache) infrastructure
necessary to build a new server without internet access.  It must be run as root
as it binds ports below 1024.

It responds to DHCP queries based on the existence of Machine.dummy.group
objects.

```shell
$ make

$ kubectl create -f - <<'EOF'
apiVersion: dummy.group/v1alpha1
kind: Machine
metadata:
  name: hostname
spec:
  diskImageUrl: http://<ip>/ubuntu-2404-kube-v1.32.4.gz
  macAddress: 11:22:33:44:55:66
  ipAddress: 1.2.3.4
EOF
```
