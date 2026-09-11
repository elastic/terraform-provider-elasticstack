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

package clients

import "github.com/hashicorp/terraform-plugin-framework/types"

// DefaultSpaceID is the Kibana default space identifier used when the resource
// does not target a specific space.
const DefaultSpaceID = "default"

// EffectiveSpaceID returns spaceID when non-empty, or DefaultSpaceID as the
// fallback for resources that do not target a specific space.
func EffectiveSpaceID(spaceID string) string {
	if spaceID == "" {
		return DefaultSpaceID
	}
	return spaceID
}

// EffectiveSpaceIDFromValue resolves a Terraform space_id attribute to its
// effective Kibana space string. Null and unknown values fall back to
// DefaultSpaceID; known values (including empty strings) are resolved via
// EffectiveSpaceID. This is the single canonical bridge between a
// types.String space_id attribute and the effective space string, and should
// be used by every Kibana space-scoped resource and data source instead of
// re-implementing the null/unknown/empty handling locally.
func EffectiveSpaceIDFromValue(spaceID types.String) string {
	if spaceID.IsNull() || spaceID.IsUnknown() {
		return DefaultSpaceID
	}
	return EffectiveSpaceID(spaceID.ValueString())
}
