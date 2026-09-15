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

// Helper function to process required fields configuration for all rule types
func (d Data) requiredFieldsToAPI(ctx context.Context) (*[]kbapi.SecurityDetectionsAPIRequiredFieldInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.RequiredFields) || len(d.RequiredFields.Elements()) == 0 {
		return nil, diags
	}

	apiRequiredFields := typeutils.ListTypeToSlice(ctx, d.RequiredFields, path.Root("required_fields"), &diags,
		func(field RequiredFieldModel, _ typeutils.ListMeta) kbapi.SecurityDetectionsAPIRequiredFieldInput {

			return kbapi.SecurityDetectionsAPIRequiredFieldInput{
				Name: field.Name.ValueString(),
				Type: field.Type.ValueString(),
			}
		})

	return &apiRequiredFields, diags
}

// convertRequiredFieldsToModel converts kbapi.SecurityDetectionsAPIRequiredFieldArray to Terraform model
func convertRequiredFieldsToModel(ctx context.Context, apiRequiredFields kbapi.SecurityDetectionsAPIRequiredFieldArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiRequiredFields) == 0 {
		return types.ListNull(getRequiredFieldElementType()), diags
	}

	fields := make([]RequiredFieldModel, 0)

	for _, apiField := range apiRequiredFields {
		field := RequiredFieldModel{
			Name: types.StringValue(apiField.Name),
			Type: types.StringValue(apiField.Type),
			Ecs:  types.BoolValue(apiField.Ecs),
		}

		fields = append(fields, field)
	}

	listValue, listDiags := types.ListValueFrom(ctx, getRequiredFieldElementType(), fields)
	diags.Append(listDiags...)
	return listValue, diags
}

func (d *Data) updateRequiredFieldsFromAPI(ctx context.Context, requiredFields *kbapi.SecurityDetectionsAPIRequiredFieldArray) diag.Diagnostics {
	var diags diag.Diagnostics
	var slice kbapi.SecurityDetectionsAPIRequiredFieldArray
	if requiredFields != nil {
		slice = *requiredFields
	}
	d.RequiredFields, diags = updateListFieldFromAPI(ctx, slice,
		types.ListNull(getRequiredFieldElementType()),
		convertRequiredFieldsToModel)
	return diags
}
