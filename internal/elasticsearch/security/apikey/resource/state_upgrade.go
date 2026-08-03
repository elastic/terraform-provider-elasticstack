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

package resource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) UpgradeState(_ context.Context) map[int64]fwresource.StateUpgrader {
	return map[int64]fwresource.StateUpgrader{
		0: {
			StateUpgrader: func(_ context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
				stateMap := stateutil.UnmarshalStateMap(req, resp)
				if resp.Diagnostics.HasError() {
					return
				}
				stateutil.NullifyEmptyString(stateMap, attrExpiration, attrMetadata, attrRoleDescriptors)
				backfillOwnerDefault(stateMap)
				stateutil.MarshalStateMap(stateMap, resp)
			},
		},
		1: {
			StateUpgrader: func(_ context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
				stateMap := stateutil.UnmarshalStateMap(req, resp)
				if resp.Diagnostics.HasError() {
					return
				}
				stateutil.NullifyEmptyString(stateMap, attrMetadata, attrRoleDescriptors)
				if v, ok := stateMap[attrType]; !ok || v == nil || v == "" {
					stateMap[attrType] = apikey.DefaultAPIKeyType
				}
				backfillOwnerDefault(stateMap)
				stateutil.MarshalStateMap(stateMap, resp)
			},
		},
		2: {
			// The `owner` attribute was added without any other schema change,
			// so states written at version 2 (by this provider prior to the
			// `owner` attribute existing, or by published releases that
			// already reached schema version 2) need only have `owner`
			// backfilled to reach version 3.
			StateUpgrader: func(_ context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
				stateMap := stateutil.UnmarshalStateMap(req, resp)
				if resp.Diagnostics.HasError() {
					return
				}
				backfillOwnerDefault(stateMap)
				stateutil.MarshalStateMap(stateMap, resp)
			},
		},
	}
}

// backfillOwnerDefault sets the `owner` attribute to its schema default
// (true) when it is absent from prior state. The `owner` attribute was added
// after schema versions 0 and 1 shipped, so any state written before it
// existed (including states written by published provider versions prior to
// this change) has no `owner` key at all. Without this backfill, the first
// plan against such state would compute `owner` via the schema's Default
// (going from null to `true`), which Terraform treats as an in-place update
// and triggers a real Update API Key call - unnecessarily, and fatally on
// Elasticsearch versions older than 8.4 that don't support that endpoint.
func backfillOwnerDefault(m map[string]any) {
	if _, ok := m[attrOwner]; !ok {
		m[attrOwner] = true
	}
}

