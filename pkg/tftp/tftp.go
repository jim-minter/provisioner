package tftp

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pin/tftp"
)

//go:generate ./generate.sh

//go:embed ubuntu-24.04.3-netboot-amd64.tar.gz
var netboot []byte

var root = map[string][]byte{}

func init() {
	r, err := gzip.NewReader(bytes.NewReader(netboot))
	if err != nil {
		panic(err)
	}

	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		b, err := io.ReadAll(tr)
		if err != nil {
			panic(err)
		}

		if h.Typeflag == tar.TypeReg {
			root["/"+path.Clean(h.Name)] = b
		}
	}

	if err = r.Close(); err != nil {
		panic(err)
	}

	root["/amd64/pxelinux.cfg/default"] = []byte(strings.Join([]string{
		"DEFAULT install",
		"LABEL install",
		"  KERNEL linux",
		"  INITRD initrd",
		"  APPEND root=/dev/ram0 ramdisk_size=1500000 ip=dhcp", // "ds=nocloud;s=http://$ip/ url=http://$ip/..."
	}, "\n"))
}

type TFTPServer struct {
	server *tftp.Server
}

func New() *TFTPServer {
	ts := &TFTPServer{}

	ts.server = tftp.NewServer(ts.handle, nil)

	return ts
}

func (ts *TFTPServer) Serve() error {
	return ts.server.ListenAndServe(":69")
}

func (ts *TFTPServer) handle(filename string, rf io.ReaderFrom) error {
	log.Print(filename)

	b, ok := root[filename]
	if !ok {
		return os.ErrNotExist
	}

	_, err := rf.ReadFrom(bytes.NewReader(b))
	return err
}
