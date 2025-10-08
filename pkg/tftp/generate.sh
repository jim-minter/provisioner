#!/bin/sh -e

if ! [ -e ubuntu-24.04.3-netboot-amd64.tar.gz ]; then
  curl -so .ubuntu-24.04.3-netboot-amd64.tar.gz https://releases.ubuntu.com/24.04.3/ubuntu-24.04.3-netboot-amd64.tar.gz
  mv .ubuntu-24.04.3-netboot-amd64.tar.gz ubuntu-24.04.3-netboot-amd64.tar.gz
fi