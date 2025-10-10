FROM ubuntu:24.04
RUN \
  apt-get update && \
  apt-get dist-upgrade -y && \
  apt-get install -y --no-install-recommends curl initramfs-tools linux-generic zstd && \
  rm -rf /var/lib/apt/lists
