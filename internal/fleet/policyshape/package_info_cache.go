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

package policyshape

import (
	"sync"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

// knownPackages is a process-wide cache of Fleet package registry metadata
// keyed by PackageCacheKey ("<name>-<version>"). It is shared by every
// resource that uses VarsJSONType, since the package info is identical
// regardless of which resource fetched it.
var knownPackages sync.Map

// LookupPackageInfo returns the cached package info for cacheKey. It satisfies
// PackageInfoLookupFunc.
func LookupPackageInfo(cacheKey string) (kbapi.KibanaHTTPAPIsGetPackageInfo, bool) {
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

// StorePackageInfo records pkg under cacheKey for future LookupPackageInfo calls.
func StorePackageInfo(cacheKey string, pkg kbapi.KibanaHTTPAPIsGetPackageInfo) {
	knownPackages.Store(cacheKey, pkg)
}

// DeletePackageInfo removes any cached entry for cacheKey.
func DeletePackageInfo(cacheKey string) {
	knownPackages.Delete(cacheKey)
}
