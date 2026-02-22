package securityuser

import _ "embed"

//go:embed resource-description.md
var userResourceDescription string

//go:embed descriptions/username.md
var usernameDescription string

//go:embed descriptions/password_hash.md
var passwordHashDescription string

//go:embed descriptions/password_wo.md
var passwordWriteOnlyDescription string

//go:embed descriptions/password_wo_version.md
var passwordWriteOnlyVersionDescription string
