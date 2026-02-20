package security

import _ "embed"

//go:embed descriptions/role_allow_restricted_indices.md
var roleAllowRestrictedIndicesDescription string

//go:embed descriptions/role_remote_indices.md
var roleRemoteIndicesDescription string

//go:embed descriptions/user_data_source.md
var userDataSourceDescription string
