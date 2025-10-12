package tftp

import (
	"bytes"
	"io"
	"log"
	"net"
	"strings"

	"github.com/pin/tftp"

	"provisioner/pkg/cache"
	"provisioner/pkg/config"
	"provisioner/pkg/tftp/assets"
)

type TFTPServer struct {
	config *config.Config
	cache  *cache.Cache
	ip     net.IP

	server *tftp.Server
}

func New(config *config.Config, cache *cache.Cache, ip net.IP) *TFTPServer {
	ts := &TFTPServer{
		config: config,
		cache:  cache,
		ip:     ip,
	}

	ts.server = tftp.NewServer(ts.handle, nil)

	return ts
}

func (ts *TFTPServer) Serve() error {
	return ts.server.ListenAndServe(":69")
}

func (ts *TFTPServer) handle(filename string, rf io.ReaderFrom) (err error) {
	log.Print(filename)

	var b []byte

	switch filename {
	case "amd64/pxelinux.cfg/default":
		b, err = ts.config.Template(strings.Join([]string{
			"DEFAULT install",
			"LABEL install",
			"  KERNEL vmlinuz",
			"  INITRD initrd.img",
			"  APPEND ip=dhcp stage2=http://{{ .Host.IP }}:8000/ubuntu-2404-kube-v1.32.4.gz ds=nocloud;s=http://" + ts.ip.String() + "/",
		}, "\n"))
		if err != nil {
			return err
		}

	default:
		b, err = assets.Assets.ReadFile(filename)
		if err != nil {
			return err
		}
	}

	_, err = rf.ReadFrom(bytes.NewReader(b))
	return err
}
