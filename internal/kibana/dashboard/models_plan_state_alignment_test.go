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

package dashboard

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_alignDashboardStateFromPlanPanels_preservesCommonPanelFields(t *testing.T) {
	planPanels := []panelModel{
		{
			MosaicConfig: &mosaicConfigModel{
				Title:       types.StringValue("Sample Mosaic"),
				Description: types.StringValue("Test mosaic visualization"),
			},
		},
		{
			EsqlControlConfig: &esqlControlConfigModel{
				EsqlQuery:        types.StringValue("FROM logs-* | KEEP host.name"),
				Title:            types.StringValue("Fields Control"),
				AvailableOptions: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("option_a")}),
			},
		},
		{
			TagcloudConfig: &tagcloudConfigModel{
				Title:       types.StringValue("Sample Tagcloud"),
				Description: types.StringValue("Test tagcloud visualization"),
				TagByJSON:   mustTagcloudJSON(`{"operation":"terms","fields":["host.name"],"limit":10}`),
			},
		},
	}

	statePanels := []panelModel{
		{
			MosaicConfig: &mosaicConfigModel{
				Title:       types.StringValue(""),
				Description: types.StringValue(""),
			},
		},
		{
			EsqlControlConfig: &esqlControlConfigModel{
				EsqlQuery:        types.StringValue(""),
				Title:            types.StringValue(""),
				AvailableOptions: types.ListNull(types.StringType),
			},
		},
		{
			TagcloudConfig: &tagcloudConfigModel{
				Title:       types.StringValue(""),
				Description: types.StringValue(""),
				TagByJSON: mustTagcloudJSON(
					`{"operation":"terms","fields":["host.name"],"limit":10,"rank_by":{"type":"metric","metric_index":0,"direction":"desc"}}`,
				),
			},
		},
	}

	for i := range min(len(planPanels), len(statePanels)) {
		alignPanelStateFromPlan(&planPanels[i], &statePanels[i])
	}

	assert.Equal(t, planPanels[0].MosaicConfig.Title, statePanels[0].MosaicConfig.Title)
	assert.Equal(t, planPanels[0].MosaicConfig.Description, statePanels[0].MosaicConfig.Description)
	assert.Equal(t, planPanels[1].EsqlControlConfig.EsqlQuery, statePanels[1].EsqlControlConfig.EsqlQuery)
	assert.Equal(t, planPanels[1].EsqlControlConfig.Title, statePanels[1].EsqlControlConfig.Title)
	assert.Equal(t, planPanels[1].EsqlControlConfig.AvailableOptions, statePanels[1].EsqlControlConfig.AvailableOptions)
	assert.Equal(t, planPanels[2].TagcloudConfig.Title, statePanels[2].TagcloudConfig.Title)
	assert.Equal(t, planPanels[2].TagcloudConfig.Description, statePanels[2].TagcloudConfig.Description)
	assert.Equal(t, planPanels[2].TagcloudConfig.TagByJSON.ValueString(), statePanels[2].TagcloudConfig.TagByJSON.ValueString())
}

func mustTagcloudJSON(v string) customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsValue(v, populateTagcloudTagByDefaults)
}
