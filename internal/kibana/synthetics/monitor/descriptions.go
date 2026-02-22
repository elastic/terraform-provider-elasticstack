package monitor

import _ "embed"

//go:embed descriptions/space_id.md
var spaceIDDescription string

//go:embed descriptions/namespace.md
var namespaceDescription string

//go:embed descriptions/retest_on_failure.md
var retestOnFailureDescription string

//go:embed descriptions/proxy_url.md
var proxyURLDescription string

//go:embed descriptions/proxy_use_local_resolver.md
var proxyUseLocalResolverDescription string
