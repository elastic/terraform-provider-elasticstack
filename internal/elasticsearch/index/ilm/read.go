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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ilmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, state.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, found, diags := readILM(ctx, client, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		tflog.Warn(ctx, fmt.Sprintf(`ILM policy "%s" not found, removing from state`, state.Name.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &finalModel)...)
}

func readILM(ctx context.Context, client *clients.APIClient, prior ilmModel) (ilmModel, bool, diag.Diagnostics) {
	model := newNullModel()
	model.ID = prior.ID
	model.ElasticsearchConnection = prior.ElasticsearchConnection

	compID, diags := clients.CompositeIDFromStrFw(prior.ID.ValueString())
	if diags.HasError() {
		return model, false, diags
	}
	policyID := compID.ResourceID

	ilmDef, diags := elasticsearch.GetIlm(ctx, client, policyID)
	if diags.HasError() {
		return model, false, diags
	}
	if ilmDef == nil {
		return model, false, nil
	}

	model.Name = types.StringValue(policyID)
	model.ModifiedDate = types.StringValue(ilmDef.Modified)
	if ilmDef.Policy.Metadata != nil {
		metadata, err := json.Marshal(ilmDef.Policy.Metadata)
		if err != nil {
			return model, false, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to marshal ILM metadata", err.Error()),
			}
		}
		model.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else {
		model.Metadata = jsontypes.NewNormalizedNull()
	}

	for _, phaseName := range []string{"hot", "warm", "cold", "frozen", "delete"} {
		priorPhase, d := getPriorPhase(ctx, prior.phaseByName(phaseName))
		if d.HasError() {
			return model, false, d
		}

		if phase, ok := ilmDef.Policy.Phases[phaseName]; ok {
			list, d := flattenPhase(ctx, phaseName, phase, priorPhase)
			if d.HasError() {
				return model, false, d
			}
			model.setPhase(phaseName, list)
		} else {
			model.setPhase(phaseName, types.ListNull(phaseElementType()))
		}
	}

	return model, true, nil
}

func getPriorPhase(ctx context.Context, list types.List) (*phaseModel, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var phases []phaseModel
	diags := list.ElementsAs(ctx, &phases, false)
	if diags.HasError() || len(phases) == 0 {
		return nil, diags
	}
	return &phases[0], nil
}
