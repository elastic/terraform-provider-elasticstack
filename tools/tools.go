//go:build tools
// +build tools

package tools

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/goreleaser/goreleaser/v2"
	_ "gopkg.in/yaml.v3"
)
