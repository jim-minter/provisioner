package cache

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"provisioner/pkg/errors"
)

const cache = "cache"

func LocalPath(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return filepath.Join(cache, u.Hostname(), u.Path), nil
}

func RemotePath(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return filepath.Join(u.Hostname(), u.Path), nil
}

func Get(url string, force bool) error {
	localPath, err := LocalPath(url)
	if err != nil {
		return err
	}

	localDir, localFile := path.Split(localPath)
	localPathTemp := path.Join(localDir, "."+localFile)

	if _, err := os.Stat(localPath); err == nil && !force {
		return nil
	}

	log.Print(url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.StatusCodeError(resp.StatusCode)
	}

	if err := os.MkdirAll(localDir, 0777); err != nil {
		return err
	}

	f, err := os.Create(localPathTemp)
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(localPathTemp)

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	return os.Rename(localPathTemp, localPath)
}

func Clean() error {
	return filepath.WalkDir(cache, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		return os.RemoveAll(path)
	})
}
