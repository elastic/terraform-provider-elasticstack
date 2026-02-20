package agentpolicy

import _ "embed"

//go:embed descriptions/host_name_format.md
var hostNameFormatDescription string

//go:embed descriptions/inactivity_timeout.md
var inactivityTimeoutDescription string

//go:embed descriptions/global_data_tags.md
var globalDataTagsDescription string
