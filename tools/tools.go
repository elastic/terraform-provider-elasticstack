//go:build tools
// +build tools

package tools

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
