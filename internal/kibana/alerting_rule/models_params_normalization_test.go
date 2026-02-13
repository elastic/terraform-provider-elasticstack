package alerting_rule

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPopulateFromAPI_OmitsAPIInjectedKeysAbsentFromPriorState(t *testing.T) {
	// Prior state has only these keys — the API will return extras that should be stripped.
	model := alertingRuleModel{
		RuleTypeID: types.StringValue(".es-query"),
		Params: jsontypes.NewNormalizedValue(`{
			"groupBy":"top",
			"termSize":10,
			"termField":"some-field",
			"timeWindowSize":10,
			"timeWindowUnit":"m",
			"threshold":[10],
			"thresholdComparator":">",
			"index":["cluster-elasticsearch-*"],
			"timeField":"@timestamp",
			"searchType":"esQuery",
			"size":10,
			"esQuery":"{}",
			"excludeHitsFromPreviousRun":true
		}`),
	}

	apiRule := &models.AlertingRule{
		RuleID:     "rule-id",
		SpaceID:    "default",
		Name:       "Test rule",
		Consumer:   "alerts",
		RuleTypeID: ".es-query",
		Schedule:   models.AlertingRuleSchedule{Interval: "10m"},
		Params: map[string]interface{}{
			// API injects aggType and someNewDefault — user never had them.
			"aggType":                    "count",
			"someNewDefault":             "injected",
			"groupBy":                    "top",
			"termSize":                   float64(10),
			"termField":                  "some-field",
			"timeWindowSize":             float64(10),
			"timeWindowUnit":             "m",
			"threshold":                  []interface{}{float64(10)},
			"thresholdComparator":        ">",
			"index":                      []interface{}{"cluster-elasticsearch-*"},
			"timeField":                  "@timestamp",
			"searchType":                 "esQuery",
			"size":                       float64(10),
			"esQuery":                    "{}",
			"excludeHitsFromPreviousRun": true,
		},
	}

	diags := model.populateFromAPI(context.Background(), apiRule)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %v", diags)
	}

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(model.Params.ValueString()), &got); err != nil {
		t.Fatalf("failed to decode model params: %v", err)
	}

	for _, key := range []string{"aggType", "someNewDefault"} {
		if _, exists := got[key]; exists {
			t.Errorf("expected key %q to be stripped from state (absent in prior params), but it was present", key)
		}
	}
	// Keys that WERE in the prior state must remain.
	expectedKeys := []string{
		"groupBy",
		"searchType",
		"size",
		"esQuery",
		"termSize",
		"termField",
		"threshold",
		"thresholdComparator",
		"index",
		"timeField",
		"timeWindowSize",
		"timeWindowUnit",
		"excludeHitsFromPreviousRun",
	}
	for _, key := range expectedKeys {
		if _, exists := got[key]; !exists {
			t.Errorf("expected key %q to remain in state, but it was missing", key)
		}
	}
}

func TestPopulateFromAPI_OmitsNestedAPIInjectedKeysAbsentFromPriorState(t *testing.T) {
	// Prior state includes `criteria` but omits nested defaults (e.g. criteria[].aggType).
	model := alertingRuleModel{
		RuleTypeID: types.StringValue(".es-query"),
		Params: jsontypes.NewNormalizedValue(`{
			"searchType":"esQuery",
			"size":10,
			"criteria":[
				{
					"threshold":[10],
					"comparator":">",
					"timeSize":5,
					"timeUnit":"m"
				}
			]
		}`),
	}

	apiRule := &models.AlertingRule{
		RuleID:     "rule-id",
		SpaceID:    "default",
		Name:       "Test rule",
		Consumer:   "alerts",
		RuleTypeID: ".es-query",
		Schedule:   models.AlertingRuleSchedule{Interval: "10m"},
		Params: map[string]interface{}{
			"searchType": "esQuery",
			"size":       float64(10),
			"criteria": []interface{}{
				map[string]interface{}{
					"threshold":  []interface{}{float64(10)},
					"comparator": ">",
					"timeSize":   float64(5),
					"timeUnit":   "m",
					// Injected nested default — user never had it.
					"aggType": "count",
				},
			},
		},
	}

	diags := model.populateFromAPI(context.Background(), apiRule)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %v", diags)
	}

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(model.Params.ValueString()), &got); err != nil {
		t.Fatalf("failed to decode model params: %v", err)
	}

	criteria, ok := got["criteria"].([]interface{})
	if !ok || len(criteria) != 1 {
		t.Fatalf("expected criteria to be a single-element array, got: %#v", got["criteria"])
	}
	crit0, ok := criteria[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected criteria[0] to be an object, got: %#v", criteria[0])
	}
	if _, exists := crit0["aggType"]; exists {
		t.Fatalf("expected nested injected key criteria[0].aggType to be stripped, but it was present")
	}
	// Keys that WERE in the prior state must remain.
	for _, key := range []string{"threshold", "comparator", "timeSize", "timeUnit"} {
		if _, exists := crit0[key]; !exists {
			t.Fatalf("expected criteria[0].%s to remain in state, but it was missing", key)
		}
	}
}

func TestPopulateFromAPI_KeepsAllKeysWhenPresentInPriorState(t *testing.T) {
	model := alertingRuleModel{
		RuleTypeID: types.StringValue(".es-query"),
		Params:     jsontypes.NewNormalizedValue(`{"searchType":"esQuery","size":10,"aggType":"count","groupBy":"all","esQuery":"{}"}`),
	}

	apiRule := &models.AlertingRule{
		RuleID:     "rule-id",
		SpaceID:    "default",
		Name:       "Test rule",
		Consumer:   "alerts",
		RuleTypeID: ".es-query",
		Schedule:   models.AlertingRuleSchedule{Interval: "10m"},
		Params: map[string]interface{}{
			"searchType": "esQuery",
			"size":       float64(10),
			"aggType":    "count",
			"groupBy":    "all",
			"esQuery":    "{}",
		},
	}

	diags := model.populateFromAPI(context.Background(), apiRule)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %v", diags)
	}

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(model.Params.ValueString()), &got); err != nil {
		t.Fatalf("failed to decode model params: %v", err)
	}

	for _, key := range []string{"searchType", "size", "aggType", "groupBy", "esQuery"} {
		if _, exists := got[key]; !exists {
			t.Errorf("expected key %q to remain in state (was in prior params), but it was missing", key)
		}
	}
}

func TestPopulateFromAPI_FirstCreate_KeepsAllAPIParams(t *testing.T) {
	// No prior state (null params) — first create. All API params should be kept.
	model := alertingRuleModel{
		RuleTypeID: types.StringNull(),
		Params:     jsontypes.NewNormalizedNull(),
	}

	apiRule := &models.AlertingRule{
		RuleID:     "rule-id",
		SpaceID:    "default",
		Name:       "Test rule",
		Consumer:   "alerts",
		RuleTypeID: ".es-query",
		Schedule:   models.AlertingRuleSchedule{Interval: "10m"},
		Params: map[string]interface{}{
			"searchType": "esQuery",
			"size":       float64(10),
			"aggType":    "count",
			"groupBy":    "all",
			"esQuery":    "{}",
		},
	}

	diags := model.populateFromAPI(context.Background(), apiRule)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %v", diags)
	}

	var got map[string]interface{}
	if err := json.Unmarshal([]byte(model.Params.ValueString()), &got); err != nil {
		t.Fatalf("failed to decode model params: %v", err)
	}

	for _, key := range []string{"searchType", "size", "aggType", "groupBy", "esQuery"} {
		if _, exists := got[key]; !exists {
			t.Errorf("expected key %q to be present on first create, but it was missing", key)
		}
	}
}
