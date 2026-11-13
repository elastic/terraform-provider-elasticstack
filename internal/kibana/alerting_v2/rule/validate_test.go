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

package rule

import (
	"context"
	"maps"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// unknown marks a configuration value as unknown at plan time.
type unknown struct{}

// tfValue builds a tftypes.Value of ty from plain Go values: nil is null,
// unknown{} is unknown, maps are objects whose missing attributes are null,
// and slices are lists.
func tfValue(t *testing.T, ty tftypes.Type, value any) tftypes.Value {
	t.Helper()
	switch v := value.(type) {
	case nil:
		return tftypes.NewValue(ty, nil)
	case unknown:
		return tftypes.NewValue(ty, tftypes.UnknownValue)
	case map[string]any:
		objectType := ty.(tftypes.Object)
		attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
		for name, attrType := range objectType.AttributeTypes {
			attributes[name] = tfValue(t, attrType, v[name])
		}
		return tftypes.NewValue(ty, attributes)
	case []any:
		elements := make([]tftypes.Value, 0, len(v))
		for _, element := range v {
			elements = append(elements, tfValue(t, ty.(tftypes.List).ElementType, element))
		}
		return tftypes.NewValue(ty, elements)
	default:
		return tftypes.NewValue(ty, v)
	}
}

func validateRuleConfig(t *testing.T, attributes map[string]any) resource.ValidateConfigResponse {
	t.Helper()
	ctx := context.Background()
	schema := getSchema(ctx)
	resp := resource.ValidateConfigResponse{}
	newResource().ValidateConfig(ctx, resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: schema, Raw: tfValue(t, schema.Type().TerraformType(ctx), attributes)},
	}, &resp)
	return resp
}

func requireValidationError(t *testing.T, resp resource.ValidateConfigResponse, attributePath string, fragments ...string) {
	t.Helper()
	require.True(t, resp.Diagnostics.HasError(), "expected a validation error")
	var seen []string
	for _, d := range resp.Diagnostics {
		text := d.Summary() + " " + d.Detail()
		seen = append(seen, text)
		if d.(interface{ Path() path.Path }).Path().String() != attributePath {
			continue
		}
		matches := true
		for _, fragment := range fragments {
			matches = matches && strings.Contains(text, fragment)
		}
		if matches {
			return
		}
	}
	require.Failf(t, "no diagnostic matched", "wanted path %q containing %q, got %v", attributePath, fragments, seen)
}

func alertRule(overrides map[string]any) map[string]any {
	attributes := map[string]any{
		"kind":     "alert",
		"metadata": map[string]any{"name": "cpu"},
		"schedule": map[string]any{"every": "1m"},
		"query":    map[string]any{"base": "FROM logs-*"},
		"recovery": map[string]any{"strategy": "no_breach"},
		"no_data":  map[string]any{"strategy": "ignore"},
	}
	maps.Copy(attributes, overrides)
	return attributes
}

func TestValidateConfig_signalRuleCannotSetRecovery(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{"kind": "signal", "no_data": nil}))

	requireValidationError(t, resp, "recovery", "signal rules cannot set `recovery` or `no_data`")
}

func TestValidateConfig_signalRuleCannotSetNoData(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{"kind": "signal", "recovery": nil}))

	requireValidationError(t, resp, "no_data", "signal rules cannot set `recovery` or `no_data`")
}

func TestValidateConfig_alertRuleRequiresNoData(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{"no_data": nil}))

	requireValidationError(t, resp, "no_data", "`no_data` is required for alert rules")
}

func TestValidateConfig_alertRuleRequiresRecovery(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{"recovery": nil}))

	requireValidationError(t, resp, "recovery", "`recovery` is required for alert rules")
}

func TestValidateConfig_validConfigurationsProduceNoDiagnostics(t *testing.T) {
	t.Parallel()

	for name, attributes := range map[string]map[string]any{
		"minimal alert rule":                     alertRule(nil),
		"signal rule":                            alertRule(map[string]any{"kind": "signal", "recovery": nil, "no_data": nil}),
		"unknown kind with lifecycle objects":    alertRule(map[string]any{"kind": unknown{}}),
		"unknown kind without lifecycle objects": alertRule(map[string]any{"kind": unknown{}, "recovery": nil, "no_data": nil}),
		"unknown recovery on an alert rule":      alertRule(map[string]any{"recovery": unknown{}}),
		"unknown recovery on a signal rule":      alertRule(map[string]any{"kind": "signal", "recovery": unknown{}, "no_data": nil}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := validateRuleConfig(t, attributes)
			require.Empty(t, resp.Diagnostics, "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}
}

func TestValidateConfig_routingTagsOnlyOnAlertRules(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"kind":     "signal",
		"metadata": map[string]any{"name": "cpu", "routing_tags": []any{"prod"}},
		"recovery": nil,
		"no_data":  nil,
	}))

	requireValidationError(t, resp, "metadata.routing_tags", "`metadata.routing_tags` is only allowed when `kind` is \"alert\"")
}

func TestValidateConfig_stateTransitionOnlyOnAlertRules(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"kind":             "signal",
		"state_transition": map[string]any{"pending": map[string]any{"count": int64(2)}},
		"recovery":         nil,
		"no_data":          nil,
	}))

	requireValidationError(t, resp, "state_transition", "`state_transition` is only allowed when `kind` is \"alert\"")
}

func TestValidateConfig_recoverySegmentOnlyWithConditionStrategy(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"recovery": map[string]any{"strategy": "no_breach", "segment": "WHERE x < 1"},
	}))

	requireValidationError(t, resp, "recovery.segment", "`segment` is only allowed with `strategy = \"condition\"`")
}

func TestValidateConfig_conditionStrategyRequiresSegment(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"query":    map[string]any{"base": "FROM logs-*", "breach": map[string]any{"segment": "WHERE x > 1"}},
		"recovery": map[string]any{"strategy": "condition"},
	}))

	requireValidationError(t, resp, "recovery.segment", "`segment` is required for `strategy = \"condition\"`")
}

func TestValidateConfig_queryStrategyRequiresQuery(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"recovery": map[string]any{"strategy": "query"},
	}))

	requireValidationError(t, resp, "recovery.query", "`query` is required for `strategy = \"query\"`")
}

func TestValidateConfig_recoveryQueryOnlyWithQueryStrategy(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"recovery": map[string]any{"strategy": "manual", "query": "FROM x"},
	}))

	requireValidationError(t, resp, "recovery.query", "`query` is only allowed with `strategy = \"query\"`")
}

func TestValidateConfig_noDataQueryNotAllowedWithIgnore(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"no_data": map[string]any{"strategy": "ignore", "query": "FROM x"},
	}))

	requireValidationError(t, resp, "no_data.query", "`query` is not allowed with `strategy = \"ignore\"`")
}

func TestValidateConfig_conditionRecoveryRequiresBreach(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"recovery": map[string]any{"strategy": "condition", "segment": "WHERE x < 1"},
	}))

	requireValidationError(t, resp, "recovery.strategy", "`recovery.strategy = \"condition\"` requires `query.breach`")
}

func TestValidateConfig_nonIgnoreNoDataRequiresBreachOrQuery(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"no_data": map[string]any{"strategy": "resolve"},
	}))

	requireValidationError(t, resp, "no_data.strategy", "a no-data strategy other than `ignore` requires `query.breach` or `no_data.query`")
}

func TestValidateConfig_manualRecoveryForbidsRecoveringPhase(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"recovery":         map[string]any{"strategy": "manual"},
		"state_transition": map[string]any{"recovering": map[string]any{"count": int64(1)}},
	}))

	requireValidationError(t, resp, "state_transition.recovering", "`state_transition.recovering` is not allowed with `recovery.strategy = \"manual\"`")
}

func TestValidateConfig_operatorRequiresBothThresholds(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"state_transition": map[string]any{"pending": map[string]any{"count": int64(2), "operator": "and"}},
	}))

	requireValidationError(t, resp, "state_transition.pending.operator", "`operator` requires both `count` and `timeframe`")
}

func TestValidateConfig_phaseRequiresCountOrTimeframe(t *testing.T) {
	t.Parallel()

	resp := validateRuleConfig(t, alertRule(map[string]any{
		"state_transition": map[string]any{"recovering": map[string]any{}},
	}))

	requireValidationError(t, resp, "state_transition.recovering", "at least one of `count` and `timeframe` is required")
}

// Each row would trip the named rule if the unknown value were treated as
// known (absent or present), so a diagnostic here means an early return is
// missing. Plan-only configurations cannot express unknowns, so this is the
// only coverage of the skip-on-unknown contract.
func TestValidateConfig_unknownInputsSuppressChecks(t *testing.T) {
	t.Parallel()

	breach := map[string]any{"base": "FROM logs-*", "breach": map[string]any{"segment": "WHERE x > 1"}}
	condition := map[string]any{"strategy": "condition", "segment": "WHERE x < 1"}

	for name, attributes := range map[string]map[string]any{
		"rules 1 and 2: unknown kind with state_transition and routing_tags": alertRule(map[string]any{
			"kind":             unknown{},
			"metadata":         map[string]any{"name": "cpu", "routing_tags": []any{"prod"}},
			"state_transition": map[string]any{"pending": map[string]any{"count": int64(1)}},
		}),
		"rule 1: unknown state_transition on a signal rule": alertRule(map[string]any{
			"kind": "signal", "state_transition": unknown{}, "recovery": nil, "no_data": nil,
		}),
		"rule 2: unknown routing_tags on a signal rule": alertRule(map[string]any{
			"kind": "signal", "metadata": map[string]any{"name": "cpu", "routing_tags": unknown{}}, "recovery": nil, "no_data": nil,
		}),
		"rule 2: unknown metadata on a signal rule": alertRule(map[string]any{
			"kind": "signal", "metadata": unknown{}, "recovery": nil, "no_data": nil,
		}),
		"rule 3: unknown no_data on an alert rule": alertRule(map[string]any{"no_data": unknown{}}),
		"rule 3: unknown no_data on a signal rule": alertRule(map[string]any{"kind": "signal", "recovery": nil, "no_data": unknown{}}),
		"recovery variant and rule 4: unknown strategy with a segment and no breach": alertRule(map[string]any{
			"recovery": map[string]any{"strategy": unknown{}, "segment": "WHERE x < 1"},
		}),
		"recovery variant: unknown strategy with a query": alertRule(map[string]any{
			"recovery": map[string]any{"strategy": unknown{}, "query": "FROM recovered"},
		}),
		"recovery variant: unknown segment with the no_breach strategy": alertRule(map[string]any{
			"recovery": map[string]any{"strategy": "no_breach", "segment": unknown{}},
		}),
		"recovery variant: unknown segment with the condition strategy": alertRule(map[string]any{
			"query": breach, "recovery": map[string]any{"strategy": "condition", "segment": unknown{}},
		}),
		"recovery variant: unknown query with the query strategy": alertRule(map[string]any{
			"recovery": map[string]any{"strategy": "query", "query": unknown{}},
		}),
		"no_data variant and rule 5: unknown strategy with a query and no breach": alertRule(map[string]any{
			"no_data": map[string]any{"strategy": unknown{}, "query": "FROM logs-*"},
		}),
		"no_data variant: unknown query with the ignore strategy": alertRule(map[string]any{
			"no_data": map[string]any{"strategy": "ignore", "query": unknown{}},
		}),
		"rule 4: unknown query with condition recovery": alertRule(map[string]any{
			"query": unknown{}, "recovery": condition,
		}),
		"rule 4: unknown breach with condition recovery": alertRule(map[string]any{
			"query": map[string]any{"base": "FROM logs-*", "breach": unknown{}}, "recovery": condition,
		}),
		"rule 5: unknown no_data query with resolve and no breach": alertRule(map[string]any{
			"no_data": map[string]any{"strategy": "resolve", "query": unknown{}},
		}),
		"rule 5: unknown breach with resolve and no query": alertRule(map[string]any{
			"query": map[string]any{"base": "FROM logs-*", "breach": unknown{}}, "no_data": map[string]any{"strategy": "resolve"},
		}),
		"rule 6: unknown state_transition with manual recovery": alertRule(map[string]any{
			"recovery": map[string]any{"strategy": "manual"}, "state_transition": unknown{},
		}),
		"rule 6: unknown recovering phase with manual recovery": alertRule(map[string]any{
			"recovery": map[string]any{"strategy": "manual"}, "state_transition": map[string]any{"recovering": unknown{}},
		}),
		"rule 7: unknown pending count with an operator": alertRule(map[string]any{
			"state_transition": map[string]any{"pending": map[string]any{"count": unknown{}, "operator": "and"}},
		}),
		"rule 7: unknown pending timeframe with an operator": alertRule(map[string]any{
			"state_transition": map[string]any{"pending": map[string]any{"timeframe": unknown{}, "operator": "and"}},
		}),
		"rule 7: unknown operator with a single threshold": alertRule(map[string]any{
			"state_transition": map[string]any{"pending": map[string]any{"count": int64(1), "operator": unknown{}}},
		}),
		"rule 7: unknown recovering phase without thresholds": alertRule(map[string]any{
			"state_transition": map[string]any{"recovering": unknown{}},
		}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := validateRuleConfig(t, attributes)
			require.Empty(t, resp.Diagnostics, "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}
}
