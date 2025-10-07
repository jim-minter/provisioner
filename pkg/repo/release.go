package repo

import (
	"encoding/hex"
	"net/textproto"
	"slices"
	"strconv"
	"strings"
)

type HashEntry struct {
	Hash []byte
	Size uint64
	Path string
}

type Release struct {
	SHA256 []*HashEntry

	Repo *Repo
}

func (repo *Repo) newRelease(h textproto.MIMEHeader) (*Release, error) {
	var sha256 []*HashEntry
	for line := range slices.Chunk(strings.Fields(h.Get("SHA256")), 3) {
		hash, err := hex.DecodeString(line[0])
		if err != nil {
			return nil, err
		}

		size, err := strconv.ParseUint(line[1], 10, 64)
		if err != nil {
			return nil, err
		}

		sha256 = append(sha256, &HashEntry{
			Hash: hash,
			Size: size,
			Path: line[2],
		})
	}

	return &Release{
		SHA256: sha256,

		Repo: repo,
	}, nil
}
