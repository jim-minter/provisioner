package repo

import (
	"encoding/hex"
	"fmt"
	"net/textproto"
	"strings"
)

type Package struct {
	Package      string
	Version      *Version
	Architecture string
	Provides     []Expression
	PreDepends   []Expression
	Depends      []Expression
	Recommends   []Expression
	Filename     string
	SHA256       []byte
	Task         map[string]struct{}

	Repo *Repo
}

func (repo *Repo) newPackage(h textproto.MIMEHeader) (*Package, error) {
	if h.Get("Package") == "" {
		return nil, fmt.Errorf("package not set")
	}

	version, err := NewVersion(h.Get("Version"))
	if err != nil {
		return nil, err
	}

	provides, err := parseExpressions(h.Get("Provides"))
	if err != nil {
		return nil, err
	}

	preDepends, err := parseExpressions(h.Get("Pre-Depends"))
	if err != nil {
		return nil, err
	}

	depends, err := parseExpressions(h.Get("Depends"))
	if err != nil {
		return nil, err
	}

	recommends, err := parseExpressions(h.Get("Recommends"))
	if err != nil {
		return nil, err
	}

	sha256, err := hex.DecodeString(h.Get("SHA256"))
	if err != nil {
		return nil, err
	}

	task := map[string]struct{}{}
	for t := range strings.SplitSeq(h.Get("Task"), ",") {
		task[strings.TrimSpace(t)] = struct{}{}
	}

	return &Package{
		Package:      h.Get("Package"),
		Version:      version,
		Architecture: h.Get("Architecture"),
		Provides:     provides,
		PreDepends:   preDepends,
		Depends:      depends,
		Recommends:   recommends,
		Filename:     h.Get("Filename"),
		SHA256:       sha256,
		Task:         task,

		Repo: repo,
	}, nil
}
