// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
