package repo

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"strings"

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
	suite := repo.Suite
	if !strings.ContainsRune(suite, '/') {
		suite = path.Join("dists", suite)
	}

	for _, remotePath := range []string{
		path.Join(suite, "InRelease"),
		path.Join(suite, "Release.gpg"),
		path.Join(suite, "Release"),
	} {
		if err := cache.Get(repo.Base + "/" + remotePath); err != nil && err != errors.StatusCodeError(http.StatusNotFound) && err != errors.StatusCodeError(http.StatusForbidden) {
			return err
		}
	}

	release, err := repo.Release()
	if err != nil {
		return err
	}

	for _, hashEntry := range release.SHA256 {
		dir, file := path.Split(hashEntry.Path)

		if dir != "" && !strings.HasPrefix(dir, repo.Component+"/") ||
			strings.Contains(dir, "/binary-") && !strings.Contains(dir, "/binary-"+repo.Arch) ||
			strings.HasSuffix(dir, "/source/") ||
			strings.HasPrefix(file, "Commands-") && !strings.HasPrefix(file, "Commands-"+repo.Arch) ||
			strings.HasPrefix(file, "Components-") && !strings.HasPrefix(file, "Components-"+repo.Arch) ||
			strings.HasPrefix(file, "Contents-") && !strings.HasPrefix(file, "Contents-"+repo.Arch) ||
			strings.HasPrefix(file, "Translation-") && file != "Translation-en" && !strings.HasPrefix(file, "Translation-en.") ||
			strings.HasPrefix(file, "icons-") {
			continue
		}

		if err := cache.Get(repo.Base + "/" + path.Join(suite, hashEntry.Path)); err != nil && err != errors.StatusCodeError(http.StatusNotFound) && err != errors.StatusCodeError(http.StatusForbidden) {
			return err
		}
	}

	return nil
}

func (repo *Repo) Release() (*Release, error) {
	suite := repo.Suite
	if !strings.ContainsRune(suite, '/') {
		suite = path.Join("dists", suite)
	}

	localPath, err := cache.LocalPath(repo.Base + "/" + path.Join(suite, "Release"))
	if err != nil {
		return nil, err
	}

	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tr := textproto.NewReader(bufio.NewReader(io.MultiReader(f, bytes.NewReader([]byte("\n\n")))))
	h, err := tr.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	return repo.newRelease(h)
}

func (repo *Repo) Packages() (pkgs []*Package, _ error) {
	suite := repo.Suite
	if !strings.ContainsRune(suite, '/') {
		suite = path.Join("dists", suite)
	}

	var localPath string
	for _, remotePath := range []string{
		path.Join(suite, repo.Component, "Packages.gz"),
		path.Join(suite, repo.Component, "binary-"+repo.Arch, "Packages.gz"),
	} {
		var err error
		localPath, err = cache.LocalPath(repo.Base + "/" + remotePath)
		if err != nil {
			return nil, err
		}

		_, err = os.Stat(localPath)
		if err == nil {
			break
		}
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
