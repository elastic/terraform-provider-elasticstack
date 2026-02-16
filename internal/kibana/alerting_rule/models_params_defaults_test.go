package alerting_rule

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
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

func TestToAPIModel_IndexThresholdBackfillsAggTypeCountWhenOmitted(t *testing.T) {
	m := alertingRuleModel{
		RuleID:     types.StringValue("rule-id"),
		SpaceID:    types.StringValue("default"),
		Name:       types.StringValue("name"),
		Consumer:   types.StringValue("alerts"),
		RuleTypeID: types.StringValue(".index-threshold"),
		Interval:   types.StringValue("1m"),
		NotifyWhen: types.StringValue("onActiveAlert"),
		Params: jsontypes.NewNormalizedValue(`{
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
	if got := apiRule.Params["aggType"]; got != "count" {
		t.Fatalf("expected params.aggType to be %q, got %#v", "count", got)
	}
	if _, exists := apiRule.Params["aggField"]; exists {
		t.Fatalf("expected params.aggField to be removed when aggType defaults to count, got %#v", apiRule.Params["aggField"])
	}
}

func TestToAPIModel_IndexThresholdDoesNotOverrideExplicitAggType(t *testing.T) {
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
	if got := apiRule.Params["aggType"]; got != "avg" {
		t.Fatalf("expected params.aggType to remain %q, got %#v", "avg", got)
	}
	if got := apiRule.Params["aggField"]; got != "version" {
		t.Fatalf("expected params.aggField to remain %q, got %#v", "version", got)
	}
}

func TestToAPIModel_IndexThresholdStripsFilterKueryForOldVersions(t *testing.T) {
	serverVersion := version.Must(version.NewVersion("8.5.3"))
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
			"filterKuery":"host.name: *",
			"timeWindowSize":10,
			"timeWindowUnit":"s",
			"threshold":[10],
			"thresholdComparator":">",
			"index":["test-index"],
			"timeField":"@timestamp"
		}`),
	}

	apiRule, diags := m.toAPIModel(context.TODO(), serverVersion)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if _, exists := apiRule.Params["filterKuery"]; exists {
		t.Fatalf("expected params.filterKuery to be stripped for old Kibana versions, got %#v", apiRule.Params["filterKuery"])
	}
}

func TestToAPIModel_IndexThresholdKeepsFilterKueryForNewVersions(t *testing.T) {
	serverVersion := version.Must(version.NewVersion("8.6.0"))
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
			"filterKuery":"host.name: *",
			"timeWindowSize":10,
			"timeWindowUnit":"s",
			"threshold":[10],
			"thresholdComparator":">",
			"index":["test-index"],
			"timeField":"@timestamp"
		}`),
	}

	apiRule, diags := m.toAPIModel(context.TODO(), serverVersion)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := apiRule.Params["filterKuery"]; got != "host.name: *" {
		t.Fatalf("expected params.filterKuery to be preserved for new Kibana versions, got %#v", got)
	}
}
