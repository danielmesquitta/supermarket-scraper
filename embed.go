package root

import (
	"embed"
)

//go:embed .env*
var Env embed.FS
