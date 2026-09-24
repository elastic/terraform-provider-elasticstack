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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper function to process risk score mapping configuration for all rule types
func (d Data) riskScoreMappingToAPI(ctx context.Context) (kbapi.SecurityDetectionsAPIRiskScoreMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiRiskScoreMapping := convertListFieldToAPI(ctx, d.RiskScoreMapping, path.Root("risk_score_mapping"), &diags,
		func(mapping RiskScoreMappingModel, _ typeutils.ListMeta) struct {
			Field     string                                              `json:"field"`
			Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
			RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
			Value     string                                              `json:"value"`
		} {
			apiMapping := struct {
				Field     string                                              `json:"field"`
				Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
				RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
				Value     string                                              `json:"value"`
			}{
				Field:    mapping.Field.ValueString(),
				Operator: kbapi.SecurityDetectionsAPIRiskScoreMappingOperator(mapping.Operator.ValueString()),
				Value:    mapping.Value.ValueString(),
			}

			// Set optional risk score if provided
			if typeutils.IsKnown(mapping.RiskScore) {
				riskScore := kbapi.SecurityDetectionsAPIRiskScore(mapping.RiskScore.ValueInt64())
				apiMapping.RiskScore = &riskScore
			}

			return apiMapping
		})

	// Return the mappings (any empty mappings were filtered out during creation)
	return apiRiskScoreMapping, diags
}

// convertRiskScoreMappingToModel converts kbapi.SecurityDetectionsAPIRiskScoreMapping to Terraform model
func convertRiskScoreMappingToModel(ctx context.Context, apiRiskScoreMapping kbapi.SecurityDetectionsAPIRiskScoreMapping) (types.List, diag.Diagnostics) {
	return convertAPISliceToListField(ctx, apiRiskScoreMapping, getRiskScoreMappingElementType(),
		func(apiMapping struct {
			Field     string                                              `json:"field"`
			Operator  kbapi.SecurityDetectionsAPIRiskScoreMappingOperator `json:"operator"`
			RiskScore *kbapi.SecurityDetectionsAPIRiskScore               `json:"risk_score,omitempty"`
			Value     string                                              `json:"value"`
		}) RiskScoreMappingModel {
			return RiskScoreMappingModel{
				Field:    types.StringValue(apiMapping.Field),
				Operator: types.StringValue(string(apiMapping.Operator)),
				Value:    types.StringValue(apiMapping.Value),

				// Set optional risk score if provided
				RiskScore: typeutils.IntPointerToInt64Value(apiMapping.RiskScore),
			}
		})
}

func (d *Data) updateRiskScoreMappingFromAPI(ctx context.Context, riskScoreMapping kbapi.SecurityDetectionsAPIRiskScoreMapping) diag.Diagnostics {
	var diags diag.Diagnostics
	d.RiskScoreMapping, diags = updateListFieldFromAPI(ctx, riskScoreMapping,
		types.ListNull(getRiskScoreMappingElementType()),
		convertRiskScoreMappingToModel)
	return diags
}
