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

package securitydetectionrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Helper function to convert MITRE ATT&CK threat data from Terraform to API format
func (d Data) threatToAPI(ctx context.Context) (kbapi.SecurityDetectionsAPIThreatArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.Threat) || len(d.Threat.Elements()) == 0 {
		return nil, diags
	}

	threats := make([]ThreatModel, len(d.Threat.Elements()))
	threatDiags := d.Threat.ElementsAs(ctx, &threats, false)
	diags.Append(threatDiags...)
	if threatDiags.HasError() {
		return nil, diags
	}

	apiThreats := make(kbapi.SecurityDetectionsAPIThreatArray, 0)
	for _, threat := range threats {
		apiThreat := kbapi.SecurityDetectionsAPIThreat{
			Framework: threat.Framework.ValueString(),
		}

		// Convert tactic
		var tacticModel ThreatTacticModel
		tacticDiags := threat.Tactic.As(ctx, &tacticModel, basetypes.ObjectAsOptions{})
		diags.Append(tacticDiags...)
		if tacticDiags.HasError() {
			continue
		}

		apiThreat.Tactic = kbapi.SecurityDetectionsAPIThreatTactic{
			Id:        tacticModel.ID.ValueString(),
			Name:      tacticModel.Name.ValueString(),
			Reference: tacticModel.Reference.ValueString(),
		}

		// Convert techniques (optional)
		if typeutils.IsKnown(threat.Technique) && len(threat.Technique.Elements()) > 0 {
			techniques := make([]ThreatTechniqueModel, len(threat.Technique.Elements()))
			techniqueDiags := threat.Technique.ElementsAs(ctx, &techniques, false)
			diags.Append(techniqueDiags...)
			if techniqueDiags.HasError() {
				continue
			}

			apiTechniques := make([]kbapi.SecurityDetectionsAPIThreatTechnique, 0)
			for _, technique := range techniques {
				apiTechnique := kbapi.SecurityDetectionsAPIThreatTechnique{
					Id:        technique.ID.ValueString(),
					Name:      technique.Name.ValueString(),
					Reference: technique.Reference.ValueString(),
				}

				// Convert subtechniques (optional)
				if typeutils.IsKnown(technique.Subtechnique) && len(technique.Subtechnique.Elements()) > 0 {
					subtechniques := make([]ThreatSubtechniqueModel, len(technique.Subtechnique.Elements()))
					subtechniqueDiags := technique.Subtechnique.ElementsAs(ctx, &subtechniques, false)
					diags.Append(subtechniqueDiags...)
					if subtechniqueDiags.HasError() {
						continue
					}

					apiSubtechniques := make([]kbapi.SecurityDetectionsAPIThreatSubtechnique, 0)
					for _, subtechnique := range subtechniques {
						apiSubtechnique := kbapi.SecurityDetectionsAPIThreatSubtechnique{
							Id:        subtechnique.ID.ValueString(),
							Name:      subtechnique.Name.ValueString(),
							Reference: subtechnique.Reference.ValueString(),
						}
						apiSubtechniques = append(apiSubtechniques, apiSubtechnique)
					}
					apiTechnique.Subtechnique = &apiSubtechniques
				}

				apiTechniques = append(apiTechniques, apiTechnique)
			}
			apiThreat.Technique = &apiTechniques
		}

		apiThreats = append(apiThreats, apiThreat)
	}

	return apiThreats, diags
}

// convertThreatToModel converts kbapi.SecurityDetectionsAPIThreatArray to Terraform model
func convertThreatToModel(ctx context.Context, apiThreats kbapi.SecurityDetectionsAPIThreatArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiThreats) == 0 {
		return types.ListNull(getThreatElementType()), diags
	}

	threats := make([]ThreatModel, 0)

	for _, apiThreat := range apiThreats {
		threat := ThreatModel{
			Framework: types.StringValue(apiThreat.Framework),
		}

		// Convert tactic
		tacticModel := ThreatTacticModel{
			ID:        types.StringValue(apiThreat.Tactic.Id),
			Name:      types.StringValue(apiThreat.Tactic.Name),
			Reference: types.StringValue(apiThreat.Tactic.Reference),
		}

		tacticObj, tacticDiags := types.ObjectValueFrom(ctx, getThreatTacticType(), tacticModel)
		diags.Append(tacticDiags...)
		if tacticDiags.HasError() {
			continue
		}
		threat.Tactic = tacticObj

		// Convert techniques (optional)
		if apiThreat.Technique != nil && len(*apiThreat.Technique) > 0 {
			techniques := make([]ThreatTechniqueModel, 0)

			for _, apiTechnique := range *apiThreat.Technique {
				technique := ThreatTechniqueModel{
					ID:        types.StringValue(apiTechnique.Id),
					Name:      types.StringValue(apiTechnique.Name),
					Reference: types.StringValue(apiTechnique.Reference),
				}

				// Convert subtechniques (optional)
				if apiTechnique.Subtechnique != nil && len(*apiTechnique.Subtechnique) > 0 {
					subtechniques := make([]ThreatSubtechniqueModel, 0)

					for _, apiSubtechnique := range *apiTechnique.Subtechnique {
						subtechnique := ThreatSubtechniqueModel{
							ID:        types.StringValue(apiSubtechnique.Id),
							Name:      types.StringValue(apiSubtechnique.Name),
							Reference: types.StringValue(apiSubtechnique.Reference),
						}
						subtechniques = append(subtechniques, subtechnique)
					}

					subtechniquesList, subtechniquesListDiags := types.ListValueFrom(ctx, getThreatSubtechniqueElementType(), subtechniques)
					diags.Append(subtechniquesListDiags...)
					if !subtechniquesListDiags.HasError() {
						technique.Subtechnique = subtechniquesList
					}
				} else {
					technique.Subtechnique = types.ListNull(getThreatSubtechniqueElementType())
				}

				techniques = append(techniques, technique)
			}

			techniquesList, techniquesListDiags := types.ListValueFrom(ctx, getThreatTechniqueElementType(), techniques)
			diags.Append(techniquesListDiags...)
			if !techniquesListDiags.HasError() {
				threat.Technique = techniquesList
			}
		} else {
			threat.Technique = types.ListNull(getThreatTechniqueElementType())
		}

		threats = append(threats, threat)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getThreatElementType(), threats)
	diags.Append(listDiags...)
	return listValue, diags
}

func (d *Data) updateThreatFromAPI(ctx context.Context, threat *kbapi.SecurityDetectionsAPIThreatArray) diag.Diagnostics {
	var diags diag.Diagnostics
	var slice kbapi.SecurityDetectionsAPIThreatArray
	if threat != nil {
		slice = *threat
	}
	d.Threat, diags = updateListFieldFromAPI(ctx, slice,
		types.ListNull(getThreatElementType()),
		convertThreatToModel)
	return diags
}
