package assets

import (
	"embed"
)

//go:generate ./generate.sh

//go:embed amd64
var Assets embed.FS
