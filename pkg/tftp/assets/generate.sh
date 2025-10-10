#!/bin/sh -e

# TODO: EFI
if ! [ -e amd64/pxelinux.0 ]; then
  curl -s https://releases.ubuntu.com/24.04.3/ubuntu-24.04.3-netboot-amd64.tar.gz | tar -xz ./amd64/pxelinux.0 ./amd64/ldlinux.c32
fi