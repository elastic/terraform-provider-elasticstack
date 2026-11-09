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

type SeverityMappingModel struct {
	Field    types.String `tfsdk:"field"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	Severity types.String `tfsdk:"severity"`
}

// Helper function to process severity mapping configuration for all rule types
func (d Data) severityMappingToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPISeverityMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiSeverityMapping := toAPIList(ctx, d.SeverityMapping, path.Root("severity_mapping"), &diags,
		func(mapping SeverityMappingModel, _ typeutils.ListMeta) struct {
			Field    string                                             `json:"field"`
			Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
			Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
			Value    string                                             `json:"value"`
		} {
			return struct {
				Field    string                                             `json:"field"`
				Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
				Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
				Value    string                                             `json:"value"`
			}{
				Field:    mapping.Field.ValueString(),
				Operator: kbapi.SecurityDetectionsAPISeverityMappingOperator(mapping.Operator.ValueString()),
				Severity: kbapi.SecurityDetectionsAPISeverity(mapping.Severity.ValueString()),
				Value:    mapping.Value.ValueString(),
			}
		})

	if apiSeverityMapping == nil {
		return nil, diags
	}

	// Convert to the expected slice type
	severityMappingSlice := make(kbapi.SecurityDetectionsAPISeverityMapping, len(apiSeverityMapping))
	copy(severityMappingSlice, apiSeverityMapping)

	return &severityMappingSlice, diags
}

// convertSeverityMappingToModel converts kbapi.SecurityDetectionsAPISeverityMapping to Terraform model
func convertSeverityMappingToModel(ctx context.Context, apiSeverityMapping *kbapi.SecurityDetectionsAPISeverityMapping) (types.List, diag.Diagnostics) {
	var slice kbapi.SecurityDetectionsAPISeverityMapping
	if apiSeverityMapping != nil {
		slice = *apiSeverityMapping
	}

	return toModelList(ctx, slice, getSeverityMappingElementType(),
		func(apiMapping struct {
			Field    string                                             `json:"field"`
			Operator kbapi.SecurityDetectionsAPISeverityMappingOperator `json:"operator"`
			Severity kbapi.SecurityDetectionsAPISeverity                `json:"severity"`
			Value    string                                             `json:"value"`
		}) SeverityMappingModel {
			return SeverityMappingModel{
				Field:    types.StringValue(apiMapping.Field),
				Operator: types.StringValue(string(apiMapping.Operator)),
				Value:    types.StringValue(apiMapping.Value),
				Severity: types.StringValue(string(apiMapping.Severity)),
			}
		})
}

func (d *Data) updateSeverityMappingFromAPI(ctx context.Context, severityMapping *kbapi.SecurityDetectionsAPISeverityMapping) diag.Diagnostics {
	var diags diag.Diagnostics
	if severityMapping != nil && len(*severityMapping) > 0 {
		d.SeverityMapping, diags = convertSeverityMappingToModel(ctx, severityMapping)
	} else {
		d.SeverityMapping = types.ListNull(getSeverityMappingElementType())
	}
	return diags
}
