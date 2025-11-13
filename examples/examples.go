package examples

import (
	"embed"
)

//go:embed data-sources
var DataSources embed.FS

//go:embed resources
var Resources embed.FS
