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

package kbschema

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// AttrTypesCache lazily reflects a set of named attr.Type maps out of a
// resource schema exactly once, then serves them by key on every subsequent
// call. Kibana Alerting-based resources build several nested object/block
// attr.Type maps from the same schema.Schema; reflecting into the schema is
// comparatively expensive and the result is immutable for the lifetime of
// the process, so it only needs to happen once per resource.
type AttrTypesCache struct {
	once   sync.Once
	values map[string]map[string]attr.Type
}

// Set records the attr.Type map for key. It is only meaningful when called
// from the populate function passed to Get.
type Set func(key string, types map[string]attr.Type)

// Get returns the cached attr.Type map for key, running populate exactly
// once (across all keys) to fill the cache on first use.
func (c *AttrTypesCache) Get(key string, populate func(set Set)) map[string]attr.Type {
	c.once.Do(func() {
		c.values = make(map[string]map[string]attr.Type)
		populate(func(k string, types map[string]attr.Type) {
			c.values[k] = types
		})
	})
	return c.values[key]
}
