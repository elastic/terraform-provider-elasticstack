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
				stateutil.NullifyEmptyString(stateMap, "expiration", "metadata", attrRoleDescriptors)
				stateutil.MarshalStateMap(stateMap, resp)
			},
		},
		1: {
			StateUpgrader: func(_ context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
				stateMap := stateutil.UnmarshalStateMap(req, resp)
				if resp.Diagnostics.HasError() {
					return
				}
				stateutil.NullifyEmptyString(stateMap, "metadata", attrRoleDescriptors)
				if v, ok := stateMap[attrType]; !ok || v == nil || v == "" {
					stateMap[attrType] = apikey.DefaultAPIKeyType
				}
				stateutil.MarshalStateMap(stateMap, resp)
			},
		},
	}
}
