package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Network struct {
		IPNet      IPNet
		Gateway    net.IP
		Nameserver net.IP
	}

	Host struct {
		IP net.IP
	}

	Laptop struct {
		IP net.IP
	}

	AuthorizedKeys []string
}

func Load() (*Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var b []byte
	for {
		b, err = os.ReadFile(filepath.Join(dir, "provisioner.yaml"))
		if err == nil {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("provisioner.yaml not found")
		}
		dir = parent
	}

	var config Config
	if err = yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
