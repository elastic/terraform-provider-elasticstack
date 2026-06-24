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

package alertingrule

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kibanacustomtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func Test_artifactsVersionSupported(t *testing.T) {
	v818, _ := version.NewVersion("8.18.0")
	v819, _ := version.NewVersion("8.19.0")
	v900, _ := version.NewVersion("9.0.0")
	v910, _ := version.NewVersion("9.1.0")

	require.False(t, artifactsVersionSupported(v818))
	require.True(t, artifactsVersionSupported(v819))
	require.False(t, artifactsVersionSupported(v900))
	require.True(t, artifactsVersionSupported(v910))
}

func Test_alertingRuleModel_toAPIModel_artifactsContent(t *testing.T) {
	ctx := context.Background()

	igObj, diags := types.ObjectValueFrom(ctx, investigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringValue("guide"),
		ContentPath: types.StringNull(),
		Checksum:    types.StringNull(),
	})
	require.False(t, diags.HasError())

	dashList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}}, []artifactDashboardModel{
		{ID: types.StringValue("dash-a")},
	})
	require.False(t, diags.HasError())

	artObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         dashList,
		InvestigationGuide: igObj,
	})
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artObj

	rule, convDiags := m.toAPIModel(ctx)
	require.False(t, convDiags.HasError())
	require.NotNil(t, rule.Artifacts)
	require.Len(t, rule.Artifacts.Dashboards, 1)
	require.Equal(t, "dash-a", rule.Artifacts.Dashboards[0].ID)
	require.Equal(t, "guide", rule.Artifacts.InvestigationGuide.Blob)
}

func Test_alertingRuleModel_toAPIModel_artifactsContentPath(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "guide.md")
	require.NoError(t, os.WriteFile(filePath, []byte("from file"), 0o600))

	igObj, diags := types.ObjectValueFrom(ctx, investigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(filePath),
		Checksum:    types.StringNull(),
	})
	require.False(t, diags.HasError())

	artObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}}),
		InvestigationGuide: igObj,
	})
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artObj

	rule, convDiags := m.toAPIModel(ctx)
	require.False(t, convDiags.HasError())
	require.Equal(t, "from file", rule.Artifacts.InvestigationGuide.Blob)
}

func Test_populateArtifactsFromAPI_preservesWhenAPIOmits(t *testing.T) {
	ctx := context.Background()

	igObj, diags := types.ObjectValueFrom(ctx, investigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringValue("kept"),
		ContentPath: types.StringNull(),
		Checksum:    types.StringNull(),
	})
	require.False(t, diags.HasError())

	artObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}}),
		InvestigationGuide: igObj,
	})
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artObj

	popDiags := m.populateArtifactsFromAPI(ctx, &models.AlertingRule{
		Name:       "n",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
	})
	require.False(t, popDiags.HasError())

	var out artifactsModel
	diags = m.Artifacts.As(ctx, &out, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())
	var outIG investigationGuideModel
	diags = out.InvestigationGuide.As(ctx, &outIG, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())
	require.Equal(t, "kept", outIG.Content.ValueString())
}

func Test_populateArtifactsFromAPI_mapsBlobToContent(t *testing.T) {
	ctx := context.Background()

	m := baseAlertingRuleModel()
	m.Artifacts = types.ObjectNull(getArtifactsAttrTypes())

	popDiags := m.populateArtifactsFromAPI(ctx, &models.AlertingRule{
		Artifacts: &models.AlertingRuleArtifacts{
			InvestigationGuide: &models.AlertingRuleArtifactInvestigationGuide{Blob: "api blob"},
		},
	})
	require.False(t, popDiags.HasError())

	var out artifactsModel
	diags := m.Artifacts.As(ctx, &out, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())
	var outIG investigationGuideModel
	diags = out.InvestigationGuide.As(ctx, &outIG, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())
	require.Equal(t, "api blob", outIG.Content.ValueString())
}

func Test_populateArtifactsFromAPI_omitsEmptyAPIArtifacts(t *testing.T) {
	ctx := context.Background()

	m := baseAlertingRuleModel()
	m.Artifacts = types.ObjectUnknown(getArtifactsAttrTypes())

	popDiags := m.populateArtifactsFromAPI(ctx, &models.AlertingRule{
		Artifacts: &models.AlertingRuleArtifacts{},
	})
	require.False(t, popDiags.HasError())
	require.True(t, m.Artifacts.IsNull())
}

func baseAlertingRuleModel() alertingRuleModel {
	return alertingRuleModel{
		ID:         types.StringValue("default/r1"),
		RuleID:     types.StringValue("r1"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("n"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".index-threshold"),
		Interval:   kibanacustomtypes.NewAlertingDurationValue("1m"),
		Params: jsontypes.NewNormalizedValue(
			`{"index":["i"],"threshold":[1],"thresholdComparator":">","timeField":"@timestamp","timeWindowSize":1,"timeWindowUnit":"m"}`,
		),
		NotifyWhen: types.StringValue("onActionGroupChange"),
	}
}
