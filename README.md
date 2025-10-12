# Provisioner

**WARNING: this is a work-in-progress and unsupported prototype which is not for
production use.  Feedback is welcomed.**


## Introduction

![Architecture](/docs/images/architecture.png)

Provisions a Kubernetes cluster.  The component parts and process are as
follows:

- The self-contained **disk image**, currently built using [Image
  Builder](https://github.com/kubernetes-sigs/image-builder).  Think Ubuntu
  24.04 + kubeadm + container images.

- The **bootstrap imageserver**.  This serves via HTTP the disk image, as well
  as a cloud-config with instructions to run `kubeadm init`.

- The **"laptop" VM**.  This boots from a second copy of the disk image and
  calls out to the bootstrap imageserver to get its cloud-config.  It runs
  `kubeadm init` to create a single node cluster, then runs `netboot` as a Pod.

- **Netboot** runs DHCP/TFTP/HTTP servers to boot subsequent nodes.  Its
  behavior is controlled by **Machine custom resources**.

- Subsequent nodes (**"node2"** and beyond) network boot via netboot's
  infrastructure.

- Netboot serves a Linux kernel plus a lightweight **stage2 initrd**.  When
  "node2" runs the initrd, it pulls the full disk image from the bootstrap
  imageserver, and reboots.

- When "node2" reboots, it gets its cloud-config from netboot and joins the
  cluster.

Provisioner includes support to test the above using
[libvirt](https://libvirt.org/).


## Let's do it!

1. **Prereqs**.  Use Ubuntu 24.04 and install at least [Go](https://go.dev/dl/),
   [Docker Engine](https://docs.docker.com/engine/install/ubuntu/) and
   [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/).

   Also install:

   ```shell
   sudo apt install ansible e2fsprogs virt-manager
   ```

   You will also need a Docker registry you can push to and pull from.

   Also, if you're running in WSL, switch on **nested virtualization** (needs
   WSL2 and Windows 11, apparently).  See [this
   link](https://learn.microsoft.com/en-us/windows/wsl/wsl-config#wslconfig) for
   more details.

1. **Configuration**.  Copy and edit provisioner.yaml.example.  The minimum you
   need to do needed is populate your registry endpoint and SSH public key so
   you can log into the machines you build.

   ```shell
   cp provisioner.yaml.example provisioner.yaml
   sed -i -e "s|YOUR-SSH-PUBLIC-KEY.*|$(cat ~/.ssh/id_rsa.pub)|" provisioner.yaml
   sed -i -e "s|YOUR-REGISTRY.*|some-value|" provisioner.yaml
   ```

1. **Build** the disk image (about 15 minutes, depending on your internet),
   bootstrap imageserver, netboot and stage2:

   ```shell
   docker login
   make all
   ```

   Also set up your libvirt infrastructure (network, "laptop" and "node2" VMs):

   ```shell
   make libvirt
   ```

1. **Run**!

   ```shell
   # Run the bootstrap imageserver:

   make bootstrap-imageserver && ./bootstrap-imageserver &

   # Run the "laptop" VM:

   virsh start laptop  # use virt-manager to check its console

   # Wait a bit, then get the admin kubeconfig:

   ssh ubuntu@192.168.123.2 sudo cat /etc/kubernetes/admin.conf >admin.conf
   export KUBECONFIG=$PWD/admin.conf
   kubectl get nodes

   # FIXME: netboot isn't yet included in the disk image and doesn't
   # automatically start.  Deploy it manually:

   make netboot-deploy

   # Create a Machine for "node2":

   kubectl create -f - <<'EOF'
   apiVersion: dummy.group/v1alpha1
   kind: Machine
   metadata:
     name: node2
   spec:
     macAddress: 52:54:00:32:0d:b8
     ipAddress: 192.168.123.3
   EOF

   # Now run "node2":

   virsh start node2

   # Wait a bit, then (hopefully) see that "node2" has joined the cluster:

   kubectl get nodes
   ```
