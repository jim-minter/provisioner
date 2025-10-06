package repo

import (
	"bufio"
	"compress/gzip"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"path"

	"provisioner/pkg/cache"
	"provisioner/pkg/errors"
)

type Repo struct {
	Base      string
	Suite     string
	Component string
	Arch      string
}

func (repo *Repo) SyncMetadata() error {
	for _, remotePath := range []string{
		path.Join("dists", repo.Suite, "InRelease"),
		path.Join("dists", repo.Suite, "Release.gpg"),
		path.Join("dists", repo.Suite, "Release"),
		path.Join("dists", repo.Suite, repo.Component, "binary-"+repo.Arch, "InRelease"),
		path.Join("dists", repo.Suite, repo.Component, "binary-"+repo.Arch, "Packages.gz"),
		path.Join("dists", repo.Suite, repo.Component, "binary-"+repo.Arch, "Packages.xz"),
		path.Join("dists", repo.Suite, repo.Component, "binary-"+repo.Arch, "Packages"),
		path.Join("dists", repo.Suite, repo.Component, "binary-"+repo.Arch, "Release.gpg"),
		path.Join("dists", repo.Suite, repo.Component, "binary-"+repo.Arch, "Release"),
		path.Join("dists", repo.Suite, repo.Component, "cnf", "Commands-"+repo.Arch+".xz"),
		path.Join("dists", repo.Suite, repo.Component, "dep11", "Components-"+repo.Arch+".yml.gz"),
		path.Join("dists", repo.Suite, repo.Component, "dep11", "Components-"+repo.Arch+".yml.xz"),
		path.Join("dists", repo.Suite, repo.Component, "i18n", "Translation-en.gz"),
		path.Join("dists", repo.Suite, repo.Component, "i18n", "Translation-en.xz"),
	} {
		if err := cache.Get(repo.Base+"/"+remotePath, true); err != nil && err != errors.StatusCodeError(http.StatusNotFound) {
			return err
		}
	}

	return nil
}

func (repo *Repo) Packages() (pkgs []*Package, _ error) {
	localPath, err := cache.LocalPath(repo.Base + "/" + path.Join("dists", repo.Suite, repo.Component, "binary-"+repo.Arch, "Packages.gz"))
	if err != nil {
		return nil, err
	}

	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	tr := textproto.NewReader(bufio.NewReader(r))
	for {
		h, err := tr.ReadMIMEHeader()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		pkg, err := repo.newPackage(h)
		if err != nil {
			return nil, err
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}
