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

package managedintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// agentlessPolicyFeatures captures Kibana-version-gated capabilities of the
// connected deployment that affect how this resource builds its request
// bodies. It mirrors internal/fleet/integration_policy/capabilities.go's
// integrationPolicyFeatures, trimmed to the one capability this resource
// currently still checks at runtime: `condition` on inputs/streams.
//
// The resource-level MinVersion floor (models.go, 9.5.0 for
// /api/fleet/managed_integrations) now matches policyshape.MinVersionCondition,
// so this separate gate is redundant and will be removed in task 4.2 of the
// fleet-managed-integration OpenSpec change.
type agentlessPolicyFeatures struct {
	SupportsCondition bool
}

// resolveAgentlessPolicyFeatures resolves agentlessPolicyFeatures against the
// connected Kibana. Called from both createAgentlessPolicy and
// updateAgentlessPolicy before building a request body, so that a `condition`
// value the connected Kibana doesn't support is caught as a clean
// attribute-scoped Terraform diagnostic (see validateInputConditionSupport in
// models_convert.go) instead of surfacing as a raw Kibana 400 ("Additional
// properties are not allowed").
func resolveAgentlessPolicyFeatures(ctx context.Context, client *clients.KibanaScopedClient) (agentlessPolicyFeatures, diag.Diagnostics) {
	var diags diag.Diagnostics
	var f agentlessPolicyFeatures

	var bitDiags diag.Diagnostics
	f.SupportsCondition, bitDiags = client.EnforceMinVersion(ctx, policyshape.MinVersionCondition)
	diags.Append(bitDiags...)

	return f, diags
}
