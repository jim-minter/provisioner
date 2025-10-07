# Provisioner

**WARNING: this is a work-in-progress and unsupported prototype which is not for
production use.  Feedback is welcomed.**

## apt-sync

`apt-sync` caches locally the relatively minimal set of binaries necessary to
subsequently build a new Ubuntu server without internet access.

You can configure the remote (both Ubuntu and third party) apt repos and package
names that you'll use.  `apt-sync` determines the closure of the most recent
versions of those packages and downloads them, as well as the ISO and netboot
file for Ubuntu 24.04.

Packages and manifests are not re-signed; they remain with the original signing
keys.

```shell
$ go build ./cmd/apt-sync
$ ./apt-sync
```

As of writing, the cache for a minimal install + updates + cri-o + k8s is 4.5GB,
3.2GB of which is the ISO.

## netboot

`netboot` runs the minimal DHCP/TFTP/HTTP (cloud-init/cache) infrastructure
necessary to build a new server without internet access.  It must be run as root
as it binds ports below 1024.

It assumes ownership of a network interface with an arbitrary IP address in an
otherwise empty subnet, and allocates the subsequent IP address for the server
to be built.

**WARNING: the client in question will be completely wiped.**  DHCP is bound to
the specified client MAC address for safety.

The client boots and runs an Ubuntu autoinstall.  Once it reboots, login is
available via the console and SSH using the `ubuntu` user and the supplied
password, from which you can `sudo` to root.

```shell
$ go build ./cmd/netboot
$ sudo ./netboot -interface virbr0 -mac 11:22:33:44:55:66 -password <secure-pw>
```
