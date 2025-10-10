package tftp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/pin/tftp"

	"provisioner/pkg/cache"
	"provisioner/pkg/tftp/assets"
)

var rxMac = regexp.MustCompile(`^amd64/pxelinux\.cfg/01-([0-9a-f]{2}-[0-9a-f]{2}-[0-9a-f]{2}-[0-9a-f]{2}-[0-9a-f]{2}-[0-9a-f]{2})$`)

type TFTPServer struct {
	cache *cache.Cache
	ip    net.IP

	server *tftp.Server
}

func New(cache *cache.Cache, ip net.IP) *TFTPServer {
	ts := &TFTPServer{
		cache: cache,
		ip:    ip,
	}

	ts.server = tftp.NewServer(ts.handle, nil)

	return ts
}

func (ts *TFTPServer) Serve() error {
	return ts.server.ListenAndServe(":69")
}

func (ts *TFTPServer) handle(filename string, rf io.ReaderFrom) (err error) {
	ctx := context.Background()

	log.Print(filename)

	var b []byte

	m := rxMac.FindStringSubmatch(filename)
	if m != nil {
		machine, err := ts.cache.Get(ctx, cache.ByMAC, strings.ReplaceAll(m[1], "-", ":"))
		if err != nil {
			log.Print(err)
			return err
		}

		if machine.Spec.DiskImageURL == "" {
			err = fmt.Errorf("spec.diskImageUrl not set")
			log.Print(err)
			return err
		}

		u, err := url.Parse(machine.Spec.DiskImageURL)
		if err != nil {
			log.Print(err)
			return err
		}

		b = []byte(strings.Join([]string{
			"DEFAULT install",
			"LABEL install",
			"  KERNEL vmlinuz",
			"  INITRD initrd.img",
			"  APPEND ip=dhcp stage2=" + u.String() + " ds=nocloud;s=http://" + ts.ip.String() + "/",
		}, "\n"))

	} else {
		b, err = assets.Assets.ReadFile(filename)
		if err != nil {
			return err
		}
	}

	_, err = rf.ReadFrom(bytes.NewReader(b))
	return err
}
