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

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func baseAlertingRuleModel() alertingRuleModel {
	return alertingRuleModel{
		ID:         types.StringValue("default/r1"),
		RuleID:     types.StringValue("r1"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("n"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".index-threshold"),
		Interval:   types.StringValue("1m"),
		Params: jsontypes.NewNormalizedValue(
			`{"index":["i"],"threshold":[1],"thresholdComparator":">","timeField":"@timestamp","timeWindowSize":1,"timeWindowUnit":"m"}`,
		),
		NotifyWhen: types.StringValue("onActionGroupChange"),
	}
}

func artifactsModelWithContent(content string) (types.Object, diag.Diagnostics) {
	ctx := context.Background()
	igObj, diags := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringValue(content),
		ContentPath: types.StringNull(),
		Checksum:    types.StringNull(),
	})
	if diags.HasError() {
		return types.ObjectNull(getArtifactsAttrTypes()), diags
	}
	return types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: getDashboardsAttrTypes()}),
		InvestigationGuide: igObj,
	})
}

func artifactsModelWithContentPath(path string) (types.Object, diag.Diagnostics) {
	ctx := context.Background()
	igObj, diags := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(path),
		Checksum:    types.StringNull(),
	})
	if diags.HasError() {
		return types.ObjectNull(getArtifactsAttrTypes()), diags
	}
	return types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: getDashboardsAttrTypes()}),
		InvestigationGuide: igObj,
	})
}

func artifactsModelWithDashboards(ids ...string) (types.Object, diag.Diagnostics) {
	ctx := context.Background()
	dashboards := make([]dashboardModel, len(ids))
	for i, id := range ids {
		dashboards[i] = dashboardModel{ID: types.StringValue(id)}
	}
	dl, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getDashboardsAttrTypes()}, dashboards)
	if diags.HasError() {
		return types.ObjectNull(getArtifactsAttrTypes()), diags
	}
	return types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         dl,
		InvestigationGuide: types.ObjectNull(getInvestigationGuideAttrTypes()),
	})
}

// === Task 3.5: version gating ===

func Test_alertingRuleModel_toAPIModel_artifactsVersionGate8x(t *testing.T) {
	ctx := context.Background()

	artifactsObj, diags := artifactsModelWithDashboards("d1")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	_, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("8.18.0")))
	require.True(t, convDiags.HasError(), "expected error for artifacts on 8.18.0")
}

func Test_alertingRuleModel_toAPIModel_artifactsVersionGate9x(t *testing.T) {
	ctx := context.Background()

	artifactsObj, diags := artifactsModelWithDashboards("d1")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	_, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("9.0.0")))
	require.True(t, convDiags.HasError(), "expected error for artifacts on 9.0.0")
}

func Test_alertingRuleModel_toAPIModel_artifactsAllowedAt819(t *testing.T) {
	ctx := context.Background()

	artifactsObj, diags := artifactsModelWithDashboards("d1")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	rule, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("8.19.0")))
	require.False(t, convDiags.HasError())
	require.NotNil(t, rule.Artifacts)
	require.Len(t, rule.Artifacts.Dashboards, 1)
}

func Test_alertingRuleModel_toAPIModel_artifactsAllowedAt91(t *testing.T) {
	ctx := context.Background()

	artifactsObj, diags := artifactsModelWithDashboards("d1")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	rule, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("9.1.0")))
	require.False(t, convDiags.HasError())
	require.NotNil(t, rule.Artifacts)
	require.Len(t, rule.Artifacts.Dashboards, 1)
}

func Test_alertingRuleModel_toAPIModel_artifactsContentRequestBody(t *testing.T) {
	ctx := context.Background()

	artifactsObj, diags := artifactsModelWithContent("inline guide text")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	rule, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("8.19.0")))
	require.False(t, convDiags.HasError())
	require.NotNil(t, rule.Artifacts)
	require.NotNil(t, rule.Artifacts.InvestigationGuide)
	require.Equal(t, "inline guide text", rule.Artifacts.InvestigationGuide.Blob)
}

func Test_alertingRuleModel_toAPIModel_artifactsContentPathRequestBody(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	guidePath := filepath.Join(tempDir, "guide.md")
	require.NoError(t, os.WriteFile(guidePath, []byte("file guide text\n"), 0o600))

	artifactsObj, diags := artifactsModelWithContentPath(guidePath)
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	rule, convDiags := m.toAPIModel(ctx, version.Must(version.NewVersion("8.19.0")))
	require.False(t, convDiags.HasError())
	require.NotNil(t, rule.Artifacts)
	require.NotNil(t, rule.Artifacts.InvestigationGuide)
	require.Equal(t, "file guide text\n", rule.Artifacts.InvestigationGuide.Blob)
}

// === Task 3.6: read-path mapping ===

func Test_populateArtifactsFromAPI_mapsBlobToContentWhenPriorStateUsedContent(t *testing.T) {
	ctx := context.Background()

	// Build prior state with inline content
	artifactsObj, diags := artifactsModelWithContent("prior content")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	// Simulate API response with a (possibly normalized) blob
	rule := &models.AlertingRule{
		Artifacts: &models.AlertingRuleArtifacts{
			InvestigationGuide: &models.AlertingRuleArtifactInvestigationGuide{
				Blob: "api blob text",
			},
		},
	}

	diags = m.populateArtifactsFromAPI(ctx, rule)
	require.False(t, diags.HasError())

	var am artifactsModel
	diags = m.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	var igm investigationGuideModel
	diags = am.InvestigationGuide.As(ctx, &igm, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	require.Equal(t, "api blob text", igm.Content.ValueString())
	require.True(t, igm.ContentPath.IsNull())
	require.True(t, igm.Checksum.IsNull())
}

func Test_populateArtifactsFromAPI_preservesContentPathWhenPriorStateUsedContentPath(t *testing.T) {
	ctx := context.Background()

	// Build prior state with content_path
	artifactsObj, diags := artifactsModelWithContentPath("/path/to/guide.md")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	// Simulate API response with blob
	rule := &models.AlertingRule{
		Artifacts: &models.AlertingRuleArtifacts{
			InvestigationGuide: &models.AlertingRuleArtifactInvestigationGuide{
				Blob: "api blob text",
			},
		},
	}

	diags = m.populateArtifactsFromAPI(ctx, rule)
	require.False(t, diags.HasError())

	var am artifactsModel
	diags = m.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	var igm investigationGuideModel
	diags = am.InvestigationGuide.As(ctx, &igm, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	// content_path should be preserved, content should remain null
	require.True(t, igm.Content.IsNull())
	require.Equal(t, "/path/to/guide.md", igm.ContentPath.ValueString())
	// checksum is also preserved (it was null in this test, but in real state it would be set)
	require.True(t, igm.Checksum.IsNull())
}

func Test_populateArtifactsFromAPI_preservesDashboardsFromAPI(t *testing.T) {
	ctx := context.Background()

	artifactsObj, diags := artifactsModelWithDashboards("d1", "d2")
	require.False(t, diags.HasError())

	m := baseAlertingRuleModel()
	m.Artifacts = artifactsObj

	rule := &models.AlertingRule{
		Artifacts: &models.AlertingRuleArtifacts{
			Dashboards: []models.AlertingRuleArtifactDashboard{
				{ID: "d1"},
				{ID: "d2"},
			},
		},
	}

	diags = m.populateArtifactsFromAPI(ctx, rule)
	require.False(t, diags.HasError())

	var am artifactsModel
	diags = m.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	var dashboards []dashboardModel
	diags = am.Dashboards.ElementsAs(ctx, &dashboards, false)
	require.False(t, diags.HasError())
	require.Len(t, dashboards, 2)
	require.Equal(t, "d1", dashboards[0].ID.ValueString())
	require.Equal(t, "d2", dashboards[1].ID.ValueString())
}
