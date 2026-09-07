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

package kibanaoapi

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func ruleWithArtifacts(blob *string) models.AlertingRule {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
	}
	if blob != nil {
		rule.Artifacts = &models.AlertingRuleArtifacts{
			InvestigationGuide: &models.AlertingRuleInvestigationGuide{Blob: *blob},
		}
	}
	return rule
}

func Test_buildRequestBody_omitsArtifactsWhenNil(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  func(models.AlertingRule) ([]byte, error)
	}{
		{"create", func(r models.AlertingRule) ([]byte, error) {
			b, err := buildCreateRequestBody(r)
			if err != nil {
				return nil, err
			}
			return json.Marshal(b)
		}},
		{"update", func(r models.AlertingRule) ([]byte, error) {
			b, err := buildUpdateRequestBody(r)
			if err != nil {
				return nil, err
			}
			return json.Marshal(b)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := tc.raw(ruleWithArtifacts(nil))
			require.NoError(t, err)
			require.NotContains(t, string(raw), `"artifacts"`)
		})
	}
}

func Test_buildRequestBody_includesInvestigationGuideBlob(t *testing.T) {
	blob := "## Runbook"
	for _, tc := range []struct {
		name string
		raw  func(models.AlertingRule) ([]byte, error)
	}{
		{"create", func(r models.AlertingRule) ([]byte, error) {
			b, err := buildCreateRequestBody(r)
			if err != nil {
				return nil, err
			}
			return json.Marshal(b)
		}},
		{"update", func(r models.AlertingRule) ([]byte, error) {
			b, err := buildUpdateRequestBody(r)
			if err != nil {
				return nil, err
			}
			return json.Marshal(b)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := tc.raw(ruleWithArtifacts(&blob))
			require.NoError(t, err)
			require.Contains(t, string(raw), `"artifacts"`)
			require.Contains(t, string(raw), `"investigation_guide"`)
			require.Contains(t, string(raw), `"blob":"## Runbook"`)
		})
	}
}

func Test_ConvertResponseToModel_readsInvestigationGuideBlob(t *testing.T) {
	resp := map[string]any{
		"id":           "r1",
		"name":         "rule",
		"consumer":     "alerts",
		"rule_type_id": ".index-threshold",
		"schedule":     map[string]any{"interval": "1m"},
		"params":       map[string]any{},
		"artifacts": map[string]any{
			"investigation_guide": map[string]any{"blob": "guide body"},
		},
	}

	model, diags := ConvertResponseToModel("default", resp)
	require.False(t, diags.HasError())
	require.NotNil(t, model.Artifacts)
	require.NotNil(t, model.Artifacts.InvestigationGuide)
	require.Equal(t, "guide body", model.Artifacts.InvestigationGuide.Blob)
}

func Test_ConvertResponseToModel_nilArtifactsWhenAbsent(t *testing.T) {
	resp := map[string]any{
		"id":           "r1",
		"name":         "rule",
		"consumer":     "alerts",
		"rule_type_id": ".index-threshold",
		"schedule":     map[string]any{"interval": "1m"},
		"params":       map[string]any{},
	}

	model, diags := ConvertResponseToModel("default", resp)
	require.False(t, diags.HasError())
	require.Nil(t, model.Artifacts)
}

func Test_buildRequestBody_includesDashboards(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		Artifacts: &models.AlertingRuleArtifacts{
			Dashboards: []models.AlertingRuleArtifactDashboard{{ID: "dash-1"}, {ID: "dash-2"}},
		},
	}
	body, err := buildCreateRequestBody(rule)
	require.NoError(t, err)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"artifacts"`)
	require.Contains(t, string(raw), `"dashboards"`)
	require.Contains(t, string(raw), `"id":"dash-1"`)
	require.Contains(t, string(raw), `"id":"dash-2"`)
	// investigation_guide must be omitted when only dashboards are set.
	require.NotContains(t, string(raw), `"investigation_guide"`)
}

func Test_buildRequestBody_includesEmptyInvestigationGuideBlob(t *testing.T) {
	blob := ""
	body, err := buildCreateRequestBody(ruleWithArtifacts(&blob))
	require.NoError(t, err)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"artifacts"`)
	require.Contains(t, string(raw), `"investigation_guide"`)
	require.Contains(t, string(raw), `"blob":""`)
}

func Test_ConvertResponseToModel_emptyArtifactsObject(t *testing.T) {
	resp := map[string]any{
		"id":           "r1",
		"name":         "rule",
		"consumer":     "alerts",
		"rule_type_id": ".index-threshold",
		"schedule":     map[string]any{"interval": "1m"},
		"params":       map[string]any{},
		"artifacts":    map[string]any{},
	}

	model, diags := ConvertResponseToModel("default", resp)
	require.False(t, diags.HasError())
	require.NotNil(t, model.Artifacts)
	require.Nil(t, model.Artifacts.InvestigationGuide)
	require.Empty(t, model.Artifacts.Dashboards)
}

func Test_ConvertResponseToModel_readsDashboards(t *testing.T) {
	resp := map[string]any{
		"id":           "r1",
		"name":         "rule",
		"consumer":     "alerts",
		"rule_type_id": ".index-threshold",
		"schedule":     map[string]any{"interval": "1m"},
		"params":       map[string]any{},
		"artifacts": map[string]any{
			"dashboards": []any{
				map[string]any{"id": "dash-1"},
				map[string]any{"id": "dash-2"},
			},
		},
	}

	model, diags := ConvertResponseToModel("default", resp)
	require.False(t, diags.HasError())
	require.NotNil(t, model.Artifacts)
	require.Nil(t, model.Artifacts.InvestigationGuide)
	require.Len(t, model.Artifacts.Dashboards, 2)
	require.Equal(t, "dash-1", model.Artifacts.Dashboards[0].ID)
	require.Equal(t, "dash-2", model.Artifacts.Dashboards[1].ID)
}
