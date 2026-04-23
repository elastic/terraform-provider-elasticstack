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

package pfresource

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const DefaultSpaceID = "default"

type CompositeID = clients.CompositeID

func ParseCompositeID(id string) (*CompositeID, diag.Diagnostics) {
	return clients.CompositeIDFromStrFw(id)
}

func ComposeCompositeID(spaceID string, resourceID string) string {
	return (&clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}).String()
}

func EffectiveSpaceID(spaceID types.String) string {
	if spaceID.IsNull() || spaceID.IsUnknown() || spaceID.ValueString() == "" {
		return DefaultSpaceID
	}
	return spaceID.ValueString()
}
