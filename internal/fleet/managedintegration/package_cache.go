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

package managedintegration

import (
	"sync"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

// knownPackages caches Fleet package registry metadata keyed by
// policyshape.PackageCacheKey ("<name>-<version>"). It backs the `vars_json`
// attribute's policyshape.VarsJSONType default-population logic wired in
// getSchema (schema.go).
//
// Task 4 (schema) only needs the read-side adapter below so the shared
// policyshape.VarsJSONType can be constructed at schema-definition time.
// Populating the cache -- calling the Fleet package-info API and storing the
// result via knownPackages.Store, mirroring
// internal/fleet/integration_policy/resource.go's getPackageInfo -- is Task
// 5's responsibility (create.go/read.go).
var knownPackages sync.Map

// lookupCachedPackageInfo adapts knownPackages to
// policyshape.PackageInfoLookupFunc.
func lookupCachedPackageInfo(cacheKey string) (kbapi.KibanaHTTPAPIsGetPackageInfo, bool) {
	value, ok := knownPackages.Load(cacheKey)
	if !ok {
		return kbapi.KibanaHTTPAPIsGetPackageInfo{}, false
	}
	pkg, ok := value.(kbapi.KibanaHTTPAPIsGetPackageInfo)
	if !ok {
		return kbapi.KibanaHTTPAPIsGetPackageInfo{}, false
	}
	return pkg, true
}
