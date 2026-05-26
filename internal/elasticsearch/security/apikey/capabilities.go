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

package apikey

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// APIKeyCapabilities records which Elasticsearch API key features the connected
// cluster supports. Persisted in Terraform private state for plan-time gating.
//
//nolint:revive // Name matches persisted private-state JSON field prefix and OpenSpec design.
type APIKeyCapabilities struct {
	SupportsUpdate          bool `json:"SupportsUpdate"`
	SupportsRoleDescriptors bool `json:"SupportsRoleDescriptors"`
	SupportsRestriction     bool `json:"SupportsRestriction"`
}

func resolveAPIKeyCapabilities(ctx context.Context, client *clients.ElasticsearchScopedClient) (APIKeyCapabilities, diag.Diagnostics) {
	var diags diag.Diagnostics
	var caps APIKeyCapabilities

	var bitDiags diag.Diagnostics
	caps.SupportsUpdate, bitDiags = client.EnforceMinVersion(ctx, MinVersionWithUpdate)
	diags.Append(bitDiags...)
	caps.SupportsRoleDescriptors, bitDiags = client.EnforceMinVersion(ctx, MinVersionReturningRoleDescriptors)
	diags.Append(bitDiags...)
	caps.SupportsRestriction, bitDiags = client.EnforceMinVersion(ctx, MinVersionWithRestriction)
	diags.Append(bitDiags...)

	return caps, diags
}

// ResolveAPIKeyCapabilities resolves API key feature support from the live cluster.
func ResolveAPIKeyCapabilities(ctx context.Context, client *clients.ElasticsearchScopedClient) (APIKeyCapabilities, diag.Diagnostics) {
	return resolveAPIKeyCapabilities(ctx, client)
}

func synthesizeAPIKeyCapabilitiesFromVersion(ver *version.Version) APIKeyCapabilities {
	return APIKeyCapabilities{
		SupportsUpdate:          !ver.LessThan(MinVersionWithUpdate),
		SupportsRoleDescriptors: !ver.LessThan(MinVersionReturningRoleDescriptors),
		SupportsRestriction:     !ver.LessThan(MinVersionWithRestriction),
	}
}

// SynthesizeAPIKeyCapabilitiesFromVersion derives capability flags from a cluster version.
func SynthesizeAPIKeyCapabilitiesFromVersion(ver *version.Version) APIKeyCapabilities {
	return synthesizeAPIKeyCapabilitiesFromVersion(ver)
}
