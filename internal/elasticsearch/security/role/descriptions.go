package role

import _ "embed"

//go:embed descriptions/allow_restricted_indices.md
var allowRestrictedIndicesDescription string

//go:embed descriptions/remote_indices.md
var remoteIndicesDescription string
