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

// Helper function to process threshold configuration for threshold rules
func (d Data) thresholdToAPI(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIThreshold {
	if !typeutils.IsKnown(d.Threshold) {
		return nil
	}

	threshold := typeutils.ObjectTypeToStruct(ctx, d.Threshold, path.Root("threshold"), diags,
		func(item ThresholdModel, meta typeutils.ObjectMeta) kbapi.SecurityDetectionsAPIThreshold {
			threshold := kbapi.SecurityDetectionsAPIThreshold{
				Value: kbapi.SecurityDetectionsAPIThresholdValue(item.Value.ValueInt64()),
			}

			// Handle threshold field(s)
			if typeutils.IsKnown(item.Field) {
				fieldList := typeutils.ListTypeToSliceString(ctx, item.Field, meta.Path.AtName("field"), meta.Diags)
				if len(fieldList) > 0 {
					var thresholdField kbapi.SecurityDetectionsAPIThresholdField
					if len(fieldList) == 1 {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField0(fieldList[0])
						if err != nil {
							meta.Diags.AddError("Error setting threshold field", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					} else {
						err := thresholdField.FromSecurityDetectionsAPIThresholdField1(fieldList)
						if err != nil {
							meta.Diags.AddError("Error setting threshold fields", err.Error())
						} else {
							threshold.Field = thresholdField
						}
					}
				}
			}

			// Handle cardinality (optional)
			if typeutils.IsKnown(item.Cardinality) {
				cardinalityList := typeutils.ListTypeToSlice(ctx, item.Cardinality, meta.Path.AtName("cardinality"), meta.Diags,
					func(item CardinalityModel, _ typeutils.ListMeta) struct {
						Field string `json:"field"`
						Value int    `json:"value"`
					} {
						return struct {
							Field string `json:"field"`
							Value int    `json:"value"`
						}{
							Field: item.Field.ValueString(),
							Value: int(item.Value.ValueInt64()),
						}
					})
				if len(cardinalityList) > 0 {
					threshold.Cardinality = &cardinalityList
				}
			}

			return threshold
		})

	return threshold
}

// convertThresholdToModel converts kbapi.SecurityDetectionsAPIThreshold to the terraform model
func convertThresholdToModel(ctx context.Context, apiThreshold kbapi.SecurityDetectionsAPIThreshold) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Handle threshold field - can be single string or array
	var fieldList types.List
	if singleField, err := apiThreshold.Field.AsSecurityDetectionsAPIThresholdField0(); err == nil {
		// Single field
		fieldList = typeutils.SliceToListTypeString(ctx, []string{singleField}, path.Root("threshold").AtName("field"), &diags)
	} else if multipleFields, err := apiThreshold.Field.AsSecurityDetectionsAPIThresholdField1(); err == nil {
		// Multiple fields
		fieldStrings := make([]string, len(multipleFields))
		copy(fieldStrings, multipleFields)
		fieldList = typeutils.SliceToListTypeString(ctx, fieldStrings, path.Root("threshold").AtName("field"), &diags)
	} else {
		fieldList = typeutils.StringsToListMust(nil)
	}

	// Handle cardinality (optional)
	var cardinalityList types.List
	if apiThreshold.Cardinality != nil && len(*apiThreshold.Cardinality) > 0 {
		cardinalityList = typeutils.SliceToListType(ctx, *apiThreshold.Cardinality, getCardinalityType(), path.Root("threshold").AtName("cardinality"), &diags,
			func(item struct {
				Field string `json:"field"`
				Value int    `json:"value"`
			}, _ typeutils.ListMeta) CardinalityModel {
				return CardinalityModel{
					Field: types.StringValue(item.Field),
					Value: types.Int64Value(int64(item.Value)),
				}
			})
	} else {
		cardinalityList = types.ListNull(getCardinalityType())
	}

	thresholdModel := ThresholdModel{
		Field:       fieldList,
		Value:       types.Int64Value(int64(apiThreshold.Value)),
		Cardinality: cardinalityList,
	}

	thresholdObject, objDiags := types.ObjectValueFrom(ctx, getThresholdType(), thresholdModel)
	diags.Append(objDiags...)
	return thresholdObject, diags
}
