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

// PackageInfoCache is a sync.Map-backed cache of Fleet package registry
// metadata keyed by PackageCacheKey ("<name>-<version>"). Its Lookup method
// satisfies PackageInfoLookupFunc as a method value.
//
// Each resource that uses VarsJSONType owns its own PackageInfoCache
// instance rather than this package owning shared cache state across
// resources (see PackageInfoLookupFunc's doc comment). The zero value is
// ready to use; the cache must not be copied after first use.
type PackageInfoCache struct {
	m sync.Map
}

// Lookup adapts the cache to PackageInfoLookupFunc.
func (c *PackageInfoCache) Lookup(cacheKey string) (kbapi.KibanaHTTPAPIsGetPackageInfo, bool) {
	value, ok := c.m.Load(cacheKey)
	if !ok {
		return kbapi.KibanaHTTPAPIsGetPackageInfo{}, false
	}
	pkg, ok := value.(kbapi.KibanaHTTPAPIsGetPackageInfo)
	if !ok {
		return kbapi.KibanaHTTPAPIsGetPackageInfo{}, false
	}
	return pkg, true
}

// Store records pkg under cacheKey for future Lookup calls.
func (c *PackageInfoCache) Store(cacheKey string, pkg kbapi.KibanaHTTPAPIsGetPackageInfo) {
	c.m.Store(cacheKey, pkg)
}

// Delete removes any cached entry for cacheKey.
func (c *PackageInfoCache) Delete(cacheKey string) {
	c.m.Delete(cacheKey)
}
