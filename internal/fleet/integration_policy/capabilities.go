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

package integrationpolicy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type integrationPolicyFeatures struct {
	SupportsPolicyIDs bool
	SupportsOutputID  bool
	SupportsCondition bool
}

func resolveIntegrationPolicyFeatures(ctx context.Context, client *clients.KibanaScopedClient) (integrationPolicyFeatures, diag.Diagnostics) {
	var diags diag.Diagnostics
	var f integrationPolicyFeatures

	var bitDiags diag.Diagnostics
	f.SupportsPolicyIDs, bitDiags = client.EnforceMinVersion(ctx, MinVersionPolicyIDs)
	diags.Append(bitDiags...)
	f.SupportsOutputID, bitDiags = client.EnforceMinVersion(ctx, MinVersionOutputID)
	diags.Append(bitDiags...)
	f.SupportsCondition, bitDiags = client.EnforceMinVersion(ctx, MinVersionCondition)
	diags.Append(bitDiags...)

	return f, diags
}
