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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	kindPath     = path.Root("kind")
	metadataPath = path.Root("metadata")
	queryPath    = path.Root(queryAttribute)
	recoveryPath = path.Root("recovery")
	noDataPath   = path.Root("no_data")

	stateTransitionPath = path.Root("state_transition")
	pendingPath         = stateTransitionPath.AtName("pending")
	recoveringPath      = stateTransitionPath.AtName("recovering")
)

// presence is whether an optional configuration value is unknown at plan
// time, omitted, or configured.
type presence int

const (
	presenceUnknown presence = iota
	presenceAbsent
	presencePresent
)

func presenceOf(value attr.Value) presence {
	switch {
	case value.IsUnknown():
		return presenceUnknown
	case value.IsNull():
		return presenceAbsent
	default:
		return presencePresent
	}
}

// configured is an optional nested object together with its presence, so a
// check can tell an unknown object (skip) from an omitted one (validate).
type configured[T any] struct {
	presence presence
	value    *T
}

func objectAs[T any](ctx context.Context, object types.Object, p path.Path, diags *diag.Diagnostics) configured[T] {
	return configured[T]{presence: presenceOf(object), value: typeutils.ObjectTypeAs[T](ctx, object, p, diags)}
}

// childObjectAs decodes a nested object inside parent. An unknown or absent
// parent gives the child the same presence and no value.
func childObjectAs[T any](ctx context.Context, parent types.Object, name string, p path.Path, diags *diag.Diagnostics) configured[T] {
	if parentPresence := presenceOf(parent); parentPresence != presencePresent {
		return configured[T]{presence: parentPresence}
	}
	child, ok := parent.Attributes()[name].(types.Object)
	if !ok {
		diags.AddAttributeError(p, "Unexpected attribute type",
			"Expected "+p.String()+" to be an object. This is a bug in the provider; please report it.")
		return configured[T]{presence: presenceUnknown}
	}
	return objectAs[T](ctx, child, p, diags)
}

// childPresence reports the presence of an attribute inside parent. An
// unknown parent makes every child unknown, which the framework does not do
// on its own when reading the child's path directly.
func childPresence(parent types.Object, name string) presence {
	if parentPresence := presenceOf(parent); parentPresence != presencePresent {
		return parentPresence
	}
	return presenceOf(parent.Attributes()[name])
}

// ruleConfig is the slice of the configuration the cross-field rules read.
// Each object is fetched on its own so that an unknown object anywhere only
// skips the checks it takes part in.
type ruleConfig struct {
	kind            types.String
	routingTags     presence
	breach          presence
	recovery        configured[recoveryModel]
	noData          configured[noDataModel]
	stateTransition presence
	pending         configured[stateTransitionPhaseModel]
	recovering      configured[stateTransitionPhaseModel]
}

func loadRuleConfig(ctx context.Context, config tfsdk.Config) (ruleConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var cfg ruleConfig
	var metadata, query, recovery, noData, stateTransition types.Object

	diags.Append(config.GetAttribute(ctx, kindPath, &cfg.kind)...)
	diags.Append(config.GetAttribute(ctx, metadataPath, &metadata)...)
	diags.Append(config.GetAttribute(ctx, queryPath, &query)...)
	diags.Append(config.GetAttribute(ctx, recoveryPath, &recovery)...)
	diags.Append(config.GetAttribute(ctx, noDataPath, &noData)...)
	diags.Append(config.GetAttribute(ctx, stateTransitionPath, &stateTransition)...)
	cfg.routingTags = childPresence(metadata, "routing_tags")
	cfg.breach = childPresence(query, "breach")
	cfg.recovery = objectAs[recoveryModel](ctx, recovery, recoveryPath, &diags)
	cfg.noData = objectAs[noDataModel](ctx, noData, noDataPath, &diags)
	cfg.stateTransition = presenceOf(stateTransition)
	cfg.pending = childObjectAs[stateTransitionPhaseModel](ctx, stateTransition, "pending", pendingPath, &diags)
	cfg.recovering = childObjectAs[stateTransitionPhaseModel](ctx, stateTransition, "recovering", recoveringPath, &diags)

	return cfg, diags
}

// ValidateConfig enforces the recovery and no_data variant contracts and the
// cross-field rules the API rejects with HTTP 400, so they fail at plan time.
// Every check is skipped while a value it depends on is unknown.
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	cfg, diags := loadRuleConfig(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(cfg.validate()...)
}

func (c ruleConfig) validate() diag.Diagnostics {
	var diags diag.Diagnostics
	c.validateAlertOnlyAttributes(&diags)
	c.validateLifecycleObjects(&diags)
	c.validateRecoveryVariant(&diags)
	c.validateNoDataVariant(&diags)
	c.validateConditionRecoveryHasBreach(&diags)
	c.validateNoDataCanDetectAbsence(&diags)
	c.validateManualRecoveryHasNoRecoveringPhase(&diags)
	validateStateTransitionPhase(&diags, pendingPath, c.pending.value)
	validateStateTransitionPhase(&diags, recoveringPath, c.recovering.value)
	return diags
}

// validateAlertOnlyAttributes enforces rules 1 and 2: state_transition and
// metadata.routing_tags are only allowed on alert rules.
func (c ruleConfig) validateAlertOnlyAttributes(diags *diag.Diagnostics) {
	if !typeutils.IsKnown(c.kind) || c.kind.ValueString() == kindAlert {
		return
	}
	if c.stateTransition == presencePresent {
		diags.AddAttributeError(stateTransitionPath, "State transition on a signal rule",
			"`state_transition` is only allowed when `kind` is \"alert\": signal rules do not create alerts, so there is no alert status to transition.")
	}
	if c.routingTags == presencePresent {
		diags.AddAttributeError(routingTagsPath, "Routing tags on a signal rule",
			"`metadata.routing_tags` is only allowed when `kind` is \"alert\": signal rules do not create alerts, so there is nothing to route to action policies.")
	}
}

// validateLifecycleObjects enforces rule 3: alert rules need both recovery
// and no_data, signal rules neither.
func (c ruleConfig) validateLifecycleObjects(diags *diag.Diagnostics) {
	if !typeutils.IsKnown(c.kind) {
		return
	}
	for _, lifecycle := range []struct {
		presence presence
		path     path.Path
	}{{c.recovery.presence, recoveryPath}, {c.noData.presence, noDataPath}} {
		switch {
		case c.kind.ValueString() == kindSignal && lifecycle.presence == presencePresent:
			diags.AddAttributeError(lifecycle.path, "Lifecycle configuration on a signal rule",
				"signal rules cannot set `recovery` or `no_data`: `"+lifecycle.path.String()+"` is only allowed when `kind` is \"alert\".")
		case c.kind.ValueString() == kindAlert && lifecycle.presence == presenceAbsent:
			diags.AddAttributeError(lifecycle.path, "Missing lifecycle configuration on an alert rule",
				"`"+lifecycle.path.String()+"` is required for alert rules: set it, or set `kind` to \"signal\".")
		}
	}
}

// validateRecoveryVariant enforces the recovery union contract: segment iff
// strategy is condition, query iff strategy is query.
func (c ruleConfig) validateRecoveryVariant(diags *diag.Diagnostics) {
	recovery := c.recovery.value
	if recovery == nil || !typeutils.IsKnown(recovery.Strategy) {
		return
	}
	strategy := recovery.Strategy.ValueString()
	validateVariantField(diags, recoveryPath, "segment", recovery.Segment, strategy, recoveryStrategyCondition,
		"the ES|QL clause appended to `query.base` that returns recovered groups")
	validateVariantField(diags, recoveryPath, queryAttribute, recovery.Query, strategy, recoveryStrategyQuery,
		"the independent ES|QL query that returns recovered groups")
}

// validateVariantField requires field exactly when strategy is
// requiredStrategy and forbids it otherwise. An unknown field is skipped.
func validateVariantField(diags *diag.Diagnostics, object path.Path, name string, field types.String, strategy, requiredStrategy, purpose string) {
	switch {
	case typeutils.IsKnown(field) && strategy != requiredStrategy:
		diags.AddAttributeError(object.AtName(name), "Variant field does not match "+object.String()+".strategy",
			"`"+name+"` is only allowed with `strategy = \""+requiredStrategy+"\"`, got `strategy = \""+strategy+"\"`.")
	case field.IsNull() && strategy == requiredStrategy:
		diags.AddAttributeError(object.AtName(name), "Missing "+object.String()+"."+name,
			"`"+name+"` is required for `strategy = \""+requiredStrategy+"\"`: "+purpose+".")
	}
}

// validateNoDataVariant enforces the no_data union contract: query is not
// allowed with the ignore strategy.
func (c ruleConfig) validateNoDataVariant(diags *diag.Diagnostics) {
	noData := c.noData.value
	if noData == nil || !typeutils.IsKnown(noData.Strategy) || noData.Query.IsUnknown() {
		return
	}
	if noData.Strategy.ValueString() == noDataStrategyIgnore && !noData.Query.IsNull() {
		diags.AddAttributeError(noDataPath.AtName(queryAttribute), "No-data query with the ignore strategy",
			"`query` is not allowed with `strategy = \"ignore\"`: the ignore strategy never checks for missing data.")
	}
}

// validateConditionRecoveryHasBreach enforces rule 4: the condition recovery
// strategy needs query.breach to define what breaching means.
func (c ruleConfig) validateConditionRecoveryHasBreach(diags *diag.Diagnostics) {
	recovery := c.recovery.value
	if recovery == nil || !typeutils.IsKnown(recovery.Strategy) || c.breach == presenceUnknown {
		return
	}
	if recovery.Strategy.ValueString() == recoveryStrategyCondition && c.breach == presenceAbsent {
		diags.AddAttributeError(recoveryPath.AtName("strategy"), "Condition recovery without a breach clause",
			"`recovery.strategy = \"condition\"` requires `query.breach`: without a breach clause there is no condition for the group to stop breaching.")
	}
}

// validateNoDataCanDetectAbsence enforces rule 5: a no_data strategy other
// than ignore needs query.breach or no_data.query, otherwise "no data" and
// "no breach" are indistinguishable.
func (c ruleConfig) validateNoDataCanDetectAbsence(diags *diag.Diagnostics) {
	noData := c.noData.value
	if noData == nil || !typeutils.IsKnown(noData.Strategy) || noData.Query.IsUnknown() || c.breach == presenceUnknown {
		return
	}
	if noData.Strategy.ValueString() != noDataStrategyIgnore && noData.Query.IsNull() && c.breach == presenceAbsent {
		diags.AddAttributeError(noDataPath.AtName("strategy"), "No-data strategy cannot detect missing data",
			"a no-data strategy other than `ignore` requires `query.breach` or `no_data.query`: without either, a group with no data is indistinguishable from a group that no longer breaches.")
	}
}

// validateManualRecoveryHasNoRecoveringPhase enforces rule 6: an alert that
// only recovers manually never passes through the recovering phase.
func (c ruleConfig) validateManualRecoveryHasNoRecoveringPhase(diags *diag.Diagnostics) {
	recovery := c.recovery.value
	if recovery == nil || !typeutils.IsKnown(recovery.Strategy) {
		return
	}
	if recovery.Strategy.ValueString() == recoveryStrategyManual && c.recovering.presence == presencePresent {
		diags.AddAttributeError(recoveringPath, "Recovering phase with manual recovery",
			"`state_transition.recovering` is not allowed with `recovery.strategy = \"manual\"`: a manually recovered alert never enters the recovering phase.")
	}
}

// validateStateTransitionPhase enforces rule 7 on one phase: at least one of
// count and timeframe is required, and operator is only allowed when both
// are set.
func validateStateTransitionPhase(diags *diag.Diagnostics, phasePath path.Path, phase *stateTransitionPhaseModel) {
	if phase == nil || phase.Count.IsUnknown() || phase.Timeframe.IsUnknown() || phase.Operator.IsUnknown() {
		return
	}
	if phase.Count.IsNull() && phase.Timeframe.IsNull() {
		diags.AddAttributeError(phasePath, "State transition phase without a threshold",
			"at least one of `count` and `timeframe` is required in `"+phasePath.String()+"`.")
		return
	}
	if !phase.Operator.IsNull() && (phase.Count.IsNull() || phase.Timeframe.IsNull()) {
		diags.AddAttributeError(phasePath.AtName("operator"), "Operator without both thresholds",
			"`operator` requires both `count` and `timeframe`: it only describes how the two thresholds combine.")
	}
}
