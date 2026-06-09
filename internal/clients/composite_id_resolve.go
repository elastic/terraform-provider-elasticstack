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

// ResolveCompositeSpaceAndID resolves a Kibana space ID and resource ID from a
// raw ID (which may be a bare resource ID or a composite "<space_id>/<resource_id>"
// string) and an optional explicit space ID attribute.
//
// Resolution rules:
//  1. Start with DefaultSpaceID as the space.
//  2. If configSpaceID is known and non-empty, use it as the space.
//  3. If rawID is a composite "<space>/<resource>" string and no explicit space
//     was provided, use the space from the composite ID.
//  4. Always extract the bare resource ID from the composite string when present.
func ResolveCompositeSpaceAndID(configSpaceID types.String, rawID string) (spaceID, resourceID string) {
	spaceExplicit := !configSpaceID.IsNull() && !configSpaceID.IsUnknown() && configSpaceID.ValueString() != ""
	spaceID = DefaultSpaceID
	if spaceExplicit {
		spaceID = configSpaceID.ValueString()
	}

	resourceID = rawID
	if compID, compDiags := CompositeIDFromStr(rawID); !compDiags.HasError() {
		resourceID = compID.ResourceID
		if !spaceExplicit {
			spaceID = compID.ClusterID
		}
	}

	return spaceID, resourceID
}
