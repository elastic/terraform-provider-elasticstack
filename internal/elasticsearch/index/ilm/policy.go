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

package ilm

import (
	"context"
	"encoding/json"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func policyFromModel(ctx context.Context, m *tfModel, serverVersion *version.Version) (*models.Policy, diag.Diagnostics) {
	var diags diag.Diagnostics
	meta := ""
	if !m.Metadata.IsNull() && !m.Metadata.IsUnknown() {
		meta = m.Metadata.ValueString()
	}
	phases := make(map[string]map[string]any)
	for _, ph := range supportedIlmPhases {
		po := m.phaseObject(ph)
		pm, d := phaseObjectToExpandMap(ctx, po)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if len(pm) > 0 {
			phases[ph] = pm
		}
	}
	return expandIlmPolicy(m.Name.ValueString(), meta, phases, serverVersion)
}

func readPolicyIntoModel(ctx context.Context, ilmDef *models.PolicyDefinition, prior *tfModel, policyName string) (*tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := &tfModel{
		ID:                      prior.ID,
		ElasticsearchConnection: prior.ElasticsearchConnection,
		Name:                    types.StringValue(policyName),
		ModifiedDate:            types.StringValue(ilmDef.Modified),
	}

	if ilmDef.Policy.Metadata != nil {
		b, err := json.Marshal(ilmDef.Policy.Metadata)
		if err != nil {
			diags.AddError("Failed to marshal metadata", err.Error())
			return nil, diags
		}
		out.Metadata = jsontypes.NewNormalizedValue(string(b))
	} else {
		out.Metadata = prior.Metadata
	}

	for _, ph := range supportedIlmPhases {
		if v, ok := ilmDef.Policy.Phases[ph]; ok {
			obj, d := flattenPhase(ctx, ph, v, prior.phaseObject(ph))
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			switch ph {
			case ilmPhaseHot:
				out.Hot = obj
			case ilmPhaseWarm:
				out.Warm = obj
			case ilmPhaseCold:
				out.Cold = obj
			case ilmPhaseFrozen:
				out.Frozen = obj
			case ilmPhaseDelete:
				out.Delete = obj
			}
		} else {
			nullObj := phaseObjectNull(ph)
			switch ph {
			case ilmPhaseHot:
				out.Hot = nullObj
			case ilmPhaseWarm:
				out.Warm = nullObj
			case ilmPhaseCold:
				out.Cold = nullObj
			case ilmPhaseFrozen:
				out.Frozen = nullObj
			case ilmPhaseDelete:
				out.Delete = nullObj
			}
		}
	}

	return out, diags
}
