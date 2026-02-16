package alerting_rule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToAPIModel_IndexThresholdBackfillsGroupByAllWhenOmitted(t *testing.T) {
	m := alertingRuleModel{
		RuleID:     types.StringValue("rule-id"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("name"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".index-threshold"),
		Interval:   types.StringValue("1m"),
		NotifyWhen: types.StringValue("onActiveAlert"),
		Params: jsontypes.NewNormalizedValue(`{
			"aggType":"avg",
			"aggField":"version",
			"timeWindowSize":10,
			"timeWindowUnit":"s",
			"threshold":[10],
			"thresholdComparator":">",
			"index":["test-index"],
			"timeField":"@timestamp"
		}`),
	}

	apiRule, diags := m.toAPIModel(context.TODO(), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if apiRule.Params == nil {
		t.Fatalf("expected apiRule.Params to be set")
	}
	if got := apiRule.Params["groupBy"]; got != "all" {
		t.Fatalf("expected params.groupBy to be %q, got %#v", "all", got)
	}
}

func TestToAPIModel_IndexThresholdDoesNotOverrideExplicitGroupBy(t *testing.T) {
	m := alertingRuleModel{
		RuleID:     types.StringValue("rule-id"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("name"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".index-threshold"),
		Interval:   types.StringValue("1m"),
		NotifyWhen: types.StringValue("onActiveAlert"),
		Params:     jsontypes.NewNormalizedValue(`{"groupBy":"top","termField":["host.name"],"termSize":10,"index":["test-index"],"timeField":"@timestamp","timeWindowSize":10,"timeWindowUnit":"s","threshold":[10],"thresholdComparator":">"}`),
	}

	apiRule, diags := m.toAPIModel(context.TODO(), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := apiRule.Params["groupBy"]; got != "top" {
		t.Fatalf("expected params.groupBy to remain %q, got %#v", "top", got)
	}
}

func TestToAPIModel_NonIndexThresholdDoesNotBackfillGroupBy(t *testing.T) {
	m := alertingRuleModel{
		RuleID:     types.StringValue("rule-id"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("name"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".es-query"),
		Interval:   types.StringValue("1m"),
		NotifyWhen: types.StringValue("onActiveAlert"),
		Params:     jsontypes.NewNormalizedValue(`{"searchType":"esQuery","size":10,"esQuery":"{}"}`),
	}

	apiRule, diags := m.toAPIModel(context.TODO(), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if _, exists := apiRule.Params["groupBy"]; exists {
		t.Fatalf("expected params.groupBy to be absent for non-.index-threshold rule types, got %#v", apiRule.Params["groupBy"])
	}
}
