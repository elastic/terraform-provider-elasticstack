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

package datastreamlifecycle

import (
	"context"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Name                    types.String `tfsdk:"name"`
	DataRetention           types.String `tfsdk:"data_retention"`
	ExpandWildcards         types.String `tfsdk:"expand_wildcards"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	Downsampling            types.List   `tfsdk:"downsampling"`
}

type downsamplingTfModel struct {
	After         types.String `tfsdk:"after"`
	FixedInterval types.String `tfsdk:"fixed_interval"`
}

func (model tfModel) GetID() (*clients.CompositeID, diag.Diagnostics) {
	compID, sdkDiags := clients.CompositeIDFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	return compID, nil
}

func (model tfModel) toAPIModel(ctx context.Context) (models.LifecycleSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiModel := models.LifecycleSettings{
		DataRetention: model.DataRetention.ValueString(),
		Enabled:       model.Enabled.ValueBool(),
	}

	if !model.Downsampling.IsNull() && !model.Downsampling.IsUnknown() && len(model.Downsampling.Elements()) > 0 {
		downsampling := make([]downsamplingTfModel, len(model.Downsampling.Elements()))
		if d := model.Downsampling.ElementsAs(ctx, &downsampling, true); d.HasError() {
			return models.LifecycleSettings{}, d
		}

		apiModel.Downsampling = make([]models.Downsampling, len(downsampling))
		for i, ds := range downsampling {
			apiModel.Downsampling[i] = models.Downsampling{
				After:         ds.After.ValueString(),
				FixedInterval: ds.FixedInterval.ValueString(),
			}
		}
	}

	return apiModel, diags
}

func (model *tfModel) populateFromAPI(ctx context.Context, ds []estypes.DataStreamWithLifecycle) diag.Diagnostics {
	actualRetention := model.DataRetention.ValueString()
	actualDownsampling := make([]downsamplingTfModel, len(model.Downsampling.Elements()))
	if diags := model.Downsampling.ElementsAs(ctx, &actualDownsampling, true); diags.HasError() {
		return nil
	}

	for _, lf := range ds {
		apiRetention := ""
		if s, ok := lf.Lifecycle.DataRetention.(string); ok {
			apiRetention = s
		}
		if apiRetention != actualRetention {
			model.DataRetention = types.StringValue(apiRetention)
		}
		var updateDownsampling bool
		apiDs := lf.Lifecycle.Downsampling
		if len(apiDs) != len(actualDownsampling) {
			updateDownsampling = true
		} else {
			for i, dstf := range actualDownsampling {
				after := ""
				fixedInterval := ""
				if len(apiDs) > i {
					if s, ok := apiDs[i].After.(string); ok {
						after = s
					}
					fixedInterval = apiDs[i].Config.FixedInterval
				}
				if dstf.After.ValueString() != after || dstf.FixedInterval.ValueString() != fixedInterval {
					updateDownsampling = true
					break
				}
			}
		}
		if updateDownsampling {
			listValue, diags := convertDownsamplingToModel(ctx, apiDs)
			if diags.HasError() {
				return diags
			}
			model.Downsampling = listValue
		}
	}
	return nil
}

func convertDownsamplingToModel(ctx context.Context, apiDownsamplings []estypes.DownsamplingRound) (types.List, diag.Diagnostics) {
	var downsamplings []downsamplingTfModel

	for _, apiDs := range apiDownsamplings {
		after := ""
		if s, ok := apiDs.After.(string); ok {
			after = s
		}
		downsamplings = append(downsamplings, downsamplingTfModel{
			After:         types.StringValue(after),
			FixedInterval: types.StringValue(apiDs.Config.FixedInterval),
		})
	}

	return types.ListValueFrom(ctx, downsamplingElementType(), downsamplings)
}
