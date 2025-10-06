package tftp

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"github.com/pin/tftp"

	"provisioner/pkg/cache"
	"provisioner/pkg/repo/ubuntu"
)

type TFTPServer struct {
	mac string
	ip  string

	root   map[string][]byte
	server *tftp.Server
}

func New(mac net.HardwareAddr, ip net.IP) (*TFTPServer, error) {
	ts := &TFTPServer{
		mac: mac.String(),
		ip:  ip.String(),
	}

	isoRemotePath, err := cache.RemotePath(ubuntu.NobleServerISOURL)
	if err != nil {
		return nil, err
	}

	netbootLocalPath, err := cache.LocalPath(ubuntu.NobleServerNetbootURL)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(netbootLocalPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	ts.root = map[string][]byte{}

	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}

		if h.Typeflag == tar.TypeReg {
			ts.root["/"+path.Clean(h.Name)] = b
		}
	}

	// https://documentation.ubuntu.com/server/how-to/installation/how-to-netboot-the-server-installer-on-amd64

	ts.root["/amd64/pxelinux.cfg/01-"+strings.ReplaceAll(ts.mac, ":", "-")] = []byte(strings.Join([]string{
		"DEFAULT install",
		"LABEL install",
		"  KERNEL linux",
		"  APPEND autoinstall cloud-config-url=/dev/null ds=nocloud;s=http://" + ts.ip + "/ initrd=initrd ip=dhcp url=http://" + ts.ip + "/" + isoRemotePath,
	}, "\n"))

	ts.server = tftp.NewServer(ts.handle, nil)

	return ts, nil
}

func (ts *TFTPServer) Serve() error {
	return ts.server.ListenAndServe(net.JoinHostPort(ts.ip, "69"))
}

func (ts *TFTPServer) handle(filename string, rf io.ReaderFrom) error {
	log.Print(filename)

	b, ok := ts.root[filename]
	if !ok {
		return os.ErrNotExist
	}

	_, err := rf.ReadFrom(bytes.NewReader(b))
	return err
}
