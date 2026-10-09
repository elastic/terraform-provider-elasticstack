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

package alertingrules

import (
	"context"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func ruleIDs(t *testing.T, rules types.List) []string {
	t.Helper()
	var elements []ruleElementModel
	require.False(t, rules.ElementsAs(context.Background(), &elements, false).HasError())
	ids := make([]string, 0, len(elements))
	for _, element := range elements {
		ids = append(ids, element.ID.ValueString())
	}
	return ids
}

func TestRead_searchNeverCallsSingleRuleFetchWithSentinel(t *testing.T) {
	t.Parallel()

	var fetched []string
	var searched bool
	api := ruleAPI{
		get: func(_ context.Context, _ *kibanaoapi.Client, _ string, ruleID string) (*models.AlertingRule, diag.Diagnostics) {
			fetched = append(fetched, ruleID)
			return nil, nil
		},
		find: func(_ context.Context, _ *kibanaoapi.Client, _ string, _ *string) ([]models.AlertingRule, diag.Diagnostics) {
			searched = true
			return nil, nil
		},
	}

	_, _, diags := readRules(context.Background(), nil, alertingRulesSearchResourceID, "default", alertingRulesDataSourceModel{}, api)
	require.False(t, diags.HasError())
	require.True(t, searched)
	require.Empty(t, fetched)
}

func TestRead_omittedSpaceIDUsesDefault(t *testing.T) {
	t.Parallel()

	var queried string
	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			t.Fatal("search must not fetch a single rule")
			return nil, nil
		},
		find: func(_ context.Context, _ *kibanaoapi.Client, spaceID string, _ *string) ([]models.AlertingRule, diag.Diagnostics) {
			queried = spaceID
			return nil, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, alertingRulesSearchResourceID, "", alertingRulesDataSourceModel{}, api)
	require.False(t, diags.HasError())
	require.True(t, found)
	require.Equal(t, "default", queried)
	require.Equal(t, "default", result.SpaceID.ValueString())
	require.Equal(t, "default", result.ID.ValueString())
}

func TestRead_configuredSpaceIDIsUsed(t *testing.T) {
	t.Parallel()

	var queried string
	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			t.Fatal("search must not fetch a single rule")
			return nil, nil
		},
		find: func(_ context.Context, _ *kibanaoapi.Client, spaceID string, _ *string) ([]models.AlertingRule, diag.Diagnostics) {
			queried = spaceID
			return nil, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, alertingRulesSearchResourceID, "ops", alertingRulesDataSourceModel{
		SpaceID: types.StringValue("ops"),
	}, api)
	require.False(t, diags.HasError())
	require.True(t, found)
	require.Equal(t, "ops", queried)
	require.Equal(t, "ops", result.SpaceID.ValueString())
	require.Equal(t, "ops", result.ID.ValueString())
}

func TestRead_lookupIDIsSpaceAndRuleID(t *testing.T) {
	t.Parallel()

	enabled := true
	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			return &models.AlertingRule{RuleID: "abc", Name: "rule", Consumer: "alerts", RuleTypeID: ".index-threshold", Enabled: &enabled}, nil
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			t.Fatal("lookup must not search")
			return nil, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, "abc", "default", alertingRulesDataSourceModel{
		SpaceID: types.StringValue("default"),
		RuleID:  types.StringValue("abc"),
	}, api)
	require.False(t, diags.HasError())
	require.True(t, found)
	require.Equal(t, "default/abc", result.ID.ValueString())
}

func TestRead_lookupReturnsSingleElement(t *testing.T) {
	t.Parallel()

	enabled := true
	api := ruleAPI{
		get: func(_ context.Context, _ *kibanaoapi.Client, spaceID, ruleID string) (*models.AlertingRule, diag.Diagnostics) {
			require.Equal(t, "default", spaceID)
			require.Equal(t, "abc", ruleID)
			return &models.AlertingRule{
				RuleID:     "abc",
				Name:       "rule",
				Consumer:   "alerts",
				RuleTypeID: ".index-threshold",
				Enabled:    &enabled,
			}, nil
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			t.Fatal("lookup must not search")
			return nil, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, "abc", "default", alertingRulesDataSourceModel{
		RuleID: types.StringValue("abc"),
	}, api)
	require.False(t, diags.HasError())
	require.True(t, found)
	require.Len(t, result.Rules.Elements(), 1)
	require.Equal(t, "abc", ruleIDs(t, result.Rules)[0])
}

func TestRead_compositeRuleIDIsFetchedLiterally(t *testing.T) {
	t.Parallel()

	var fetched string
	api := ruleAPI{
		get: func(_ context.Context, _ *kibanaoapi.Client, spaceID, ruleID string) (*models.AlertingRule, diag.Diagnostics) {
			require.Equal(t, "ops", spaceID)
			fetched = ruleID
			return nil, nil
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			t.Fatal("composite rule_id must not be searched")
			return nil, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, "other-space/abc", "ops", alertingRulesDataSourceModel{
		SpaceID: types.StringValue("ops"),
		RuleID:  types.StringValue("other-space/abc"),
	}, api)
	require.False(t, diags.HasError())
	require.False(t, found)
	require.Equal(t, "other-space/abc", fetched)
	require.True(t, result.Rules.IsNull() || len(result.Rules.Elements()) == 0)
}

func TestRead_unknownRuleIDIsNotFound(t *testing.T) {
	t.Parallel()

	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			return nil, nil
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			t.Fatal("unknown rule lookup must not search")
			return nil, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, "missing", "default", alertingRulesDataSourceModel{
		RuleID: types.StringValue("missing"),
	}, api)
	require.False(t, diags.HasError())
	require.False(t, found)
	require.True(t, result.Rules.IsNull() || len(result.Rules.Elements()) == 0)
}

func TestRead_lookupAuthorizationFailureReturnsError(t *testing.T) {
	t.Parallel()

	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unexpected status code from server: got HTTP 403", `{"statusCode":403,"error":"Forbidden","message":"not allowed"}`)}
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			t.Fatal("authorization failure must not fall through to search")
			return nil, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, "abc", "ops", alertingRulesDataSourceModel{
		RuleID: types.StringValue("abc"),
	}, api)
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "not allowed")
	require.False(t, found)
	require.True(t, result.Rules.IsNull() || len(result.Rules.Elements()) == 0)
}

func TestRead_emptySearchIsFoundWithEmptyList(t *testing.T) {
	t.Parallel()

	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			t.Fatal("empty search must not fetch a single rule")
			return nil, nil
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			return []models.AlertingRule{}, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, alertingRulesSearchResourceID, "default", alertingRulesDataSourceModel{}, api)
	require.False(t, diags.HasError())
	require.True(t, found)
	require.False(t, result.Rules.IsNull())
	require.Empty(t, result.Rules.Elements())
}

func TestRead_filterIsPassedUnmodified(t *testing.T) {
	t.Parallel()

	const filter = `alert.attributes.name: "a-rule"`
	enabled := true
	var got *string
	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			t.Fatal("filtered search must not fetch a single rule")
			return nil, nil
		},
		find: func(_ context.Context, _ *kibanaoapi.Client, _ string, filter *string) ([]models.AlertingRule, diag.Diagnostics) {
			got = filter
			return []models.AlertingRule{{
				RuleID:     "rule-1",
				Name:       "a-rule",
				Consumer:   "alerts",
				RuleTypeID: ".index-threshold",
				Enabled:    &enabled,
			}}, nil
		},
	}

	result, found, diags := readRules(context.Background(), nil, alertingRulesSearchResourceID, "default", alertingRulesDataSourceModel{
		Filter: types.StringValue(filter),
	}, api)
	require.False(t, diags.HasError())
	require.True(t, found)
	require.NotNil(t, got)
	require.Equal(t, filter, *got)
	require.Equal(t, []string{"rule-1"}, ruleIDs(t, result.Rules))
}

func TestRead_searchAuthorizationFailureReturnsError(t *testing.T) {
	t.Parallel()

	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			t.Fatal("authorization failure must not fetch a single rule")
			return nil, nil
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Unexpected status code from server: got HTTP 403",
				`{"statusCode":403,"error":"Forbidden","message":"not allowed"}`,
			)}
		},
	}

	result, found, diags := readRules(context.Background(), nil, alertingRulesSearchResourceID, "ops", alertingRulesDataSourceModel{}, api)
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "not allowed")
	require.False(t, found)
	require.True(t, result.Rules.IsNull() || len(result.Rules.Elements()) == 0)
}

func TestRead_secondLookupReflectsLaterExecution(t *testing.T) {
	t.Parallel()

	first, err := time.Parse(time.RFC3339, "2026-09-18T16:15:41.450Z")
	require.NoError(t, err)
	later, err := time.Parse(time.RFC3339, "2026-09-18T16:16:41.450Z")
	require.NoError(t, err)
	status := "ok"
	enabled := true
	dates := []*time.Time{&first, &later}
	var calls int
	api := ruleAPI{
		get: func(context.Context, *kibanaoapi.Client, string, string) (*models.AlertingRule, diag.Diagnostics) {
			when := dates[calls]
			calls++
			return &models.AlertingRule{
				RuleID:     "abc",
				Name:       "rule",
				Consumer:   "alerts",
				RuleTypeID: ".index-threshold",
				Enabled:    &enabled,
				ExecutionStatus: models.AlertingRuleExecutionStatus{
					Status:            &status,
					LastExecutionDate: when,
				},
			}, nil
		},
		find: func(context.Context, *kibanaoapi.Client, string, *string) ([]models.AlertingRule, diag.Diagnostics) {
			t.Fatal("lookup must not search")
			return nil, nil
		},
	}
	config := alertingRulesDataSourceModel{RuleID: types.StringValue("abc")}

	firstResult, found, diags := readRules(context.Background(), nil, "abc", "default", config, api)
	require.False(t, diags.HasError())
	require.True(t, found)
	secondResult, found, diags := readRules(context.Background(), nil, "abc", "default", config, api)
	require.False(t, diags.HasError())
	require.True(t, found)

	require.Equal(t, "2026-09-18T16:15:41.450Z", executionDates(t, firstResult.Rules)[0])
	require.Equal(t, "2026-09-18T16:16:41.450Z", executionDates(t, secondResult.Rules)[0])
}

func executionDates(t *testing.T, rules types.List) []string {
	t.Helper()
	var elements []ruleElementModel
	require.False(t, rules.ElementsAs(context.Background(), &elements, false).HasError())
	dates := make([]string, 0, len(elements))
	for _, element := range elements {
		dates = append(dates, element.LastExecutionDate.ValueString())
	}
	return dates
}
