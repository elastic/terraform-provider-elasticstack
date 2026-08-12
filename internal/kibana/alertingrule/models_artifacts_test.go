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
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kibanacustomtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

// baseModel returns a minimal valid model to attach artifacts to.
func baseModel() alertingRuleModel {
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

func artifactsObject(t *testing.T, ig investigationGuideModel) basetypes.ObjectValue {
	t.Helper()
	ctx := context.Background()
	igObj, diags := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), ig)
	require.False(t, diags.HasError())
	artObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{InvestigationGuide: igObj})
	require.False(t, diags.HasError())
	return artObj
}

func Test_toAPIModel_investigationGuide_inlineContent(t *testing.T) {
	ctx := context.Background()

	m := baseModel()
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringValue("## Runbook\nDo the thing."),
		ContentPath: types.StringNull(),
		Checksum:    types.StringNull(),
	})

	rule, diags := m.toAPIModel(ctx)
	require.False(t, diags.HasError())
	require.NotNil(t, rule.Artifacts)
	require.NotNil(t, rule.Artifacts.InvestigationGuide)
	require.Equal(t, "## Runbook\nDo the thing.", rule.Artifacts.InvestigationGuide.Blob)
}

func Test_toAPIModel_investigationGuide_contentPathReadsFile(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	fpath := filepath.Join(dir, "guide.md")
	require.NoError(t, os.WriteFile(fpath, []byte("file guide body"), 0o600))

	m := baseModel()
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(fpath),
		Checksum:    types.StringUnknown(),
	})

	rule, diags := m.toAPIModel(ctx)
	require.False(t, diags.HasError())
	require.NotNil(t, rule.Artifacts)
	require.NotNil(t, rule.Artifacts.InvestigationGuide)
	require.Equal(t, "file guide body", rule.Artifacts.InvestigationGuide.Blob)
}

func Test_toAPIModel_investigationGuide_contentPathMissingFileErrors(t *testing.T) {
	ctx := context.Background()

	m := baseModel()
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(filepath.Join(t.TempDir(), "does-not-exist.md")),
		Checksum:    types.StringUnknown(),
	})

	_, diags := m.toAPIModel(ctx)
	require.True(t, diags.HasError())
}

func Test_populateArtifactsFromAPI_inlineContentStoresBlob(t *testing.T) {
	ctx := context.Background()

	m := baseModel()
	// Prior state used inline content.
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringValue("old"),
		ContentPath: types.StringNull(),
		Checksum:    types.StringNull(),
	})

	rule := &models.AlertingRule{
		Artifacts: &models.AlertingRuleArtifacts{
			InvestigationGuide: &models.AlertingRuleInvestigationGuide{Blob: "new blob from api"},
		},
	}

	diags := m.populateArtifactsFromAPI(ctx, rule)
	require.False(t, diags.HasError())

	ig, d := m.investigationGuideFrom(ctx)
	require.False(t, d.HasError())
	require.NotNil(t, ig)
	require.Equal(t, "new blob from api", ig.Content.ValueString())
	require.True(t, ig.ContentPath.IsNull())
	require.True(t, ig.Checksum.IsNull())
}

func Test_populateArtifactsFromAPI_contentPathPreservedNotOverwritten(t *testing.T) {
	ctx := context.Background()

	m := baseModel()
	// Prior state used a file path with a recorded checksum.
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue("/tmp/guide.md"),
		Checksum:    types.StringValue("abc123"),
	})

	rule := &models.AlertingRule{
		Artifacts: &models.AlertingRuleArtifacts{
			InvestigationGuide: &models.AlertingRuleInvestigationGuide{Blob: "server side blob"},
		},
	}

	diags := m.populateArtifactsFromAPI(ctx, rule)
	require.False(t, diags.HasError())

	ig, d := m.investigationGuideFrom(ctx)
	require.False(t, d.HasError())
	require.NotNil(t, ig)
	// content_path source: content is not surfaced, path + checksum preserved.
	require.True(t, ig.Content.IsNull())
	require.Equal(t, "/tmp/guide.md", ig.ContentPath.ValueString())
	require.Equal(t, "abc123", ig.Checksum.ValueString())
}

func Test_applyInvestigationGuideChecksum_contentPathComputesSHA(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	fpath := filepath.Join(dir, "guide.md")
	body := []byte("checksum me")
	require.NoError(t, os.WriteFile(fpath, body, 0o600))

	sum := sha256.Sum256(body)
	want := hex.EncodeToString(sum[:])

	m := baseModel()
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(fpath),
		Checksum:    types.StringUnknown(),
	})

	diags := m.applyInvestigationGuideChecksum(ctx)
	require.False(t, diags.HasError())

	ig, d := m.investigationGuideFrom(ctx)
	require.False(t, d.HasError())
	require.NotNil(t, ig)
	require.Equal(t, want, ig.Checksum.ValueString())
}

func Test_GetVersionRequirements_artifactsSetRequires91(t *testing.T) {
	ctx := context.Background()

	m := baseModel()
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringValue("x"),
		ContentPath: types.StringNull(),
		Checksum:    types.StringNull(),
	})

	reqs, diags := m.GetVersionRequirements(ctx)
	require.False(t, diags.HasError())

	var found bool
	want := version.Must(version.NewVersion("9.1.0"))
	for _, r := range reqs {
		if r.MinVersion.Equal(want) {
			found = true
		}
	}
	require.True(t, found, "expected a >= 9.1.0 version requirement when artifacts is set, got %+v", reqs)
}

func Test_GetVersionRequirements_noArtifactsNoRequirement(t *testing.T) {
	ctx := context.Background()

	m := baseModel()
	m.Artifacts = types.ObjectNull(getArtifactsAttrTypes())

	reqs, diags := m.GetVersionRequirements(ctx)
	require.False(t, diags.HasError())

	want := version.Must(version.NewVersion("9.1.0"))
	for _, r := range reqs {
		require.False(t, r.MinVersion.Equal(want), "did not expect the artifacts version gate when artifacts is null")
	}
}

func Test_applyInvestigationGuideChecksum_inlineContentClearsChecksum(t *testing.T) {
	ctx := context.Background()

	m := baseModel()
	m.Artifacts = artifactsObject(t, investigationGuideModel{
		Content:     types.StringValue("inline"),
		ContentPath: types.StringNull(),
		Checksum:    types.StringNull(),
	})

	diags := m.applyInvestigationGuideChecksum(ctx)
	require.False(t, diags.HasError())

	ig, d := m.investigationGuideFrom(ctx)
	require.False(t, d.HasError())
	require.NotNil(t, ig)
	require.True(t, ig.Checksum.IsNull())
}
