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

package anomalydetectionjob

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func baselineTFModelForBuildFromPlan(ctx context.Context, t *testing.T) TFModel {
	t.Helper()
	mpcAttrTypes := getModelPlotConfigAttrTypes(ctx)
	alAttrTypes := getAnalysisLimitsAttrTypes(ctx)

	mpcObj := types.ObjectValueMust(mpcAttrTypes, map[string]attr.Value{
		"enabled":             types.BoolValue(false),
		"annotations_enabled": types.BoolValue(false),
		"terms":               types.StringValue(""),
	})
	alObj := types.ObjectValueMust(alAttrTypes, map[string]attr.Value{
		"categorization_examples_limit": types.Int64Null(),
		"model_memory_limit":            customtypes.NewMemorySizeValue("128mb"),
	})

	return TFModel{
		JobID:                                types.StringValue("job-unit-test"),
		Description:                          types.StringValue("steady description"),
		Groups:                               types.SetNull(types.StringType),
		ModelPlotConfig:                      mpcObj,
		AnalysisLimits:                       alObj,
		AllowLazyOpen:                        types.BoolNull(),
		BackgroundPersistInterval:            types.StringNull(),
		CustomSettings:                       jsontypes.NewNormalizedNull(),
		DailyModelSnapshotRetentionAfterDays: types.Int64Null(),
		ModelSnapshotRetentionDays:           types.Int64Null(),
		RenormalizationWindowDays:            types.Int64Null(),
		ResultsRetentionDays:                 types.Int64Null(),
	}
}

func TestUpdateAPIModel_BuildFromPlan_noChangesWhenUpdatablesMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	base := baselineTFModelForBuildFromPlan(ctx, t)

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &base, &base)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, hasChanges)
	require.Nil(t, u.Description)
	require.Nil(t, u.AllowLazyOpen)
}

func TestUpdateAPIModel_BuildFromPlan_partialBodyWhenDescriptionChanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baselineTFModelForBuildFromPlan(ctx, t)

	plan := prior
	plan.Description = types.StringValue("updated description")

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &plan, &prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, hasChanges)
	require.NotNil(t, u.Description)
	require.Equal(t, "updated description", *u.Description)
	require.Nil(t, u.Groups)
	require.Nil(t, u.ModelPlotConfig)
	require.Nil(t, u.AnalysisLimits)
}

func TestUpdateAPIModel_BuildFromPlan_emptyCustomSettingsSendsEmptyObject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baselineTFModelForBuildFromPlan(ctx, t)
	prior.CustomSettings = jsontypes.NewNormalizedValue(`{"created_by":"advanced-wizard"}`)

	plan := prior
	plan.CustomSettings = jsontypes.NewNormalizedValue("{}")

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &plan, &prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, hasChanges)
	require.JSONEq(t, "{}", string(u.CustomSettings))

	body, err := json.Marshal(u)
	require.NoError(t, err)
	require.Contains(t, string(body), `"custom_settings":{}`)
}

func TestUpdateAPIModel_BuildFromPlan_nullCustomSettingsIsOmitted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baselineTFModelForBuildFromPlan(ctx, t)
	prior.CustomSettings = jsontypes.NewNormalizedValue(`{"created_by":"advanced-wizard"}`)

	plan := prior
	plan.CustomSettings = jsontypes.NewNormalizedNull()

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &plan, &prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, hasChanges)
	require.Nil(t, u.CustomSettings)

	body, err := json.Marshal(u)
	require.NoError(t, err)
	require.NotContains(t, string(body), "custom_settings")
}

func TestUpdateAPIModel_BuildFromPlan_jsonNullCustomSettingsIsOmitted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baselineTFModelForBuildFromPlan(ctx, t)
	prior.CustomSettings = jsontypes.NewNormalizedValue(`{"created_by":"advanced-wizard"}`)

	plan := prior
	plan.CustomSettings = jsontypes.NewNormalizedValue("null")

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &plan, &prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, hasChanges)
	require.Nil(t, u.CustomSettings)

	body, err := json.Marshal(u)
	require.NoError(t, err)
	require.NotContains(t, string(body), "custom_settings")
}

func TestUpdateAPIModel_BuildFromPlan_nonEmptyCustomSettingsReplacesBag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baselineTFModelForBuildFromPlan(ctx, t)
	prior.CustomSettings = jsontypes.NewNormalizedValue(`{"department":"ops","created_by":"advanced-wizard"}`)

	plan := prior
	plan.CustomSettings = jsontypes.NewNormalizedValue(`{"department":"ops"}`)

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &plan, &prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, hasChanges)
	require.JSONEq(t, `{"department":"ops"}`, string(u.CustomSettings))

	body, err := json.Marshal(u)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, map[string]any{"department": "ops"}, decoded["custom_settings"])
}

func TestAPIModel_toPutJobRequest_emptyCustomSettingsSendsEmptyObject(t *testing.T) {
	t.Parallel()

	a := APIModel{
		JobID: "job-unit-test",
		AnalysisConfig: AnalysisConfigAPIModel{
			BucketSpan: "15m",
			Detectors:  []DetectorAPIModel{{Function: "count"}},
		},
		DataDescription: DataDescriptionAPIModel{
			TimeField: "@timestamp",
		},
		CustomSettings: map[string]any{},
	}

	req := a.toPutJobRequest()
	require.JSONEq(t, "{}", string(req.CustomSettings))
}
