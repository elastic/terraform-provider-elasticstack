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

// Helper function to process investigation fields configuration for all rule types
func (d Data) investigationFieldsToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIInvestigationFields, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiFieldNames := toAPIList(ctx, d.InvestigationFields, path.Root("investigation_fields"), &diags,
		func(fieldName string, _ typeutils.ListMeta) kbapi.SecurityDetectionsAPINonEmptyString {
			return kbapi.SecurityDetectionsAPINonEmptyString(fieldName)
		})

	if apiFieldNames == nil {
		return nil, diags
	}

	return &kbapi.SecurityDetectionsAPIInvestigationFields{
		FieldNames: apiFieldNames,
	}, diags
}

// convertInvestigationFieldsToModel converts kbapi.SecurityDetectionsAPIInvestigationFields to Terraform model
func convertInvestigationFieldsToModel(ctx context.Context, apiInvestigationFields *kbapi.SecurityDetectionsAPIInvestigationFields) (types.List, diag.Diagnostics) {
	var fieldNames []kbapi.SecurityDetectionsAPINonEmptyString
	if apiInvestigationFields != nil {
		fieldNames = apiInvestigationFields.FieldNames
	}

	return toModelList(ctx, fieldNames, types.StringType,
		func(fieldName kbapi.SecurityDetectionsAPINonEmptyString) string {
			return string(fieldName)
		})
}

// Helper function to update investigation fields from API response
func (d *Data) updateInvestigationFieldsFromAPI(ctx context.Context, investigationFields *kbapi.SecurityDetectionsAPIInvestigationFields) diag.Diagnostics {
	var diags diag.Diagnostics

	investigationFieldsValue, investigationFieldsDiags := convertInvestigationFieldsToModel(ctx, investigationFields)
	diags.Append(investigationFieldsDiags...)
	if diags.HasError() {
		return diags
	}
	d.InvestigationFields = investigationFieldsValue

	return diags
}
