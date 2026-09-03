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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// paramsSchemaSpec contains precomputed key metadata and decode factory for
// one generated params variant. This avoids reflection in runtime validation.
type paramsSchemaSpec struct {
	name                  string
	newTarget             func() any
	requiredKeys          map[string]struct{}
	additionalAllowedKeys []string // keys stripped only for this variant
}

// legacyMonitorStatusParams models the long-standing runtime payload shape that
// Kibana still accepts for monitor status rules, even though the regenerated
// spec now also includes a newer `condition` form.
type legacyMonitorStatusParams struct {
	Availability *struct {
		Range     float32 `json:"range"`
		RangeUnit string  `json:"rangeUnit"`
		Threshold string  `json:"threshold"`
	} `json:"availability,omitempty"`
	Filters *struct {
		Tags            *[]string `json:"tags,omitempty"`
		MonitorType     *[]string `json:"monitor.type,omitempty"`
		ObserverGeoName *[]string `json:"observer.geo.name,omitempty"`
		URLPort         *[]string `json:"url.port,omitempty"`
	} `json:"filters,omitempty"`
	NumTimes                float32  `json:"numTimes"`
	Search                  *string  `json:"search,omitempty"`
	ShouldCheckAvailability bool     `json:"shouldCheckAvailability"`
	ShouldCheckStatus       bool     `json:"shouldCheckStatus"`
	StackVersion            *string  `json:"stackVersion,omitempty"`
	TimerangeCount          *float32 `json:"timerangeCount,omitempty"`
	TimerangeUnit           *string  `json:"timerangeUnit,omitempty"`
}

func mustNewParamsSchemaSpec(newTarget func() any) paramsSchemaSpec {
	target := newTarget()
	name := fmt.Sprintf("%T", target)
	requiredKeys, err := computeRequiredKeys(target)
	if err != nil {
		panic(fmt.Sprintf("alerting_rule: mustNewParamsSchemaSpec(%q): computeRequiredKeys error: %v", name, err))
	}
	if requiredKeys == nil {
		panic(fmt.Sprintf("alerting_rule: mustNewParamsSchemaSpec(%q): computeRequiredKeys returned nil requiredKeys (err=%v)", name, err))
	}
	return paramsSchemaSpec{
		name:         name,
		newTarget:    newTarget,
		requiredKeys: requiredKeys,
	}
}

func mustNewParamsSchemaSpecFromContainer(newContainer func() any) paramsSchemaSpec {
	return mustNewParamsSchemaSpec(func() any {
		container := newContainer()
		containerType := reflect.TypeOf(container)
		if containerType.Kind() != reflect.Pointer {
			panic(fmt.Sprintf("alerting_rule: params container %T must be a pointer", container))
		}

		paramsField, ok := containerType.Elem().FieldByName("Params")
		if !ok {
			panic(fmt.Sprintf("alerting_rule: params container %T is missing a Params field", container))
		}

		if paramsField.Type.Kind() == reflect.Pointer {
			return reflect.New(paramsField.Type.Elem()).Interface()
		}

		return reflect.New(paramsField.Type).Interface()
	})
}

// ValidateConfig is the single validation entry point for rule params. It runs
// during the plan phase so invalid params are caught before any API call. The
// apply-phase toAPIModel intentionally does not re-validate to avoid duplicate
// error messages.
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data alertingRuleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &resp.Diagnostics)

	if !typeutils.IsKnown(data.Params) || !typeutils.IsKnown(data.RuleTypeID) {
		return
	}

	params := typeutils.NormalizedTypeToMap[any](data.Params, path.Root("params"), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	errs := validateRuleParams(data.RuleTypeID.ValueString(), params)
	if len(errs) == 0 {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("params"),
		fmt.Sprintf("Invalid params for rule_type_id %q", data.RuleTypeID.ValueString()),
		formatParamsValidationErrors(errs),
	)
}

// ruleTypeParamsOverrides contains explicit validation overrides for rule types
// where OpenAPI does not match Kibana runtime behavior or where params are
// nested unions requiring multiple variant attempts.
var ruleTypeParamsOverrides = map[string][]paramsSchemaSpec{
	ruleTypeIndexThreshold: {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsIndexThresholdCreateRuleBodyAlerting{} }),
	},
	ruleTypeESQuery: {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsEsQueryCreateRuleBodyAlerting{} }),
	},
	ruleTypeUptimeMonitorStatus: {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsXpackUptimeAlertsMonitorstatusCreateRuleBodyAlerting{} }),
		mustNewParamsSchemaSpec(func() any { return &legacyMonitorStatusParams{} }),
	},
	"logs.alert.document.count": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.KibanaHTTPAPIsLogsAlertDocumentCountCreateRuleBodyAlertingParams0{} }),
		mustNewParamsSchemaSpec(func() any { return &kbapi.KibanaHTTPAPIsLogsAlertDocumentCountCreateRuleBodyAlertingParams1{} }),
	},
}

var ruleTypeAdditionalAllowedParamsKeys = map[string][]string{}

// validateParamsViaDiscriminator validates params for a rule type by dispatching
// through kbapi.AlertingRuleAPIBody.ValueByDiscriminator(), extracting the typed
// Params value via reflection, and re-decoding the practitioner's params with
// DisallowUnknownFields(). Unknown discriminators return nil (pass-through).
func validateParamsViaDiscriminator(ruleTypeID string, params map[string]any) []string {
	// Build a minimal stub rule body sufficient for discriminator dispatch.
	stub := map[string]any{
		attrRuleTypeID: ruleTypeID,
		attrParams:     params,
	}
	stubRaw, err := json.Marshal(stub)
	if err != nil {
		return []string{fmt.Sprintf("failed to marshal stub rule body: %v", err)}
	}

	var body kbapi.AlertingRuleAPIBody
	if err := body.UnmarshalJSON(stubRaw); err != nil {
		return []string{fmt.Sprintf("failed to unmarshal stub rule body: %v", err)}
	}

	typedVal, err := body.ValueByDiscriminator()
	if err != nil {
		if strings.Contains(err.Error(), "unknown discriminator value") {
			return nil
		}
		return []string{fmt.Sprintf("failed to dispatch by discriminator: %v", err)}
	}

	// Extract the Params field from the discriminated value via reflection.
	typedReflect := reflect.ValueOf(typedVal)
	if typedReflect.Kind() == reflect.Pointer {
		typedReflect = typedReflect.Elem()
	}
	paramsField := typedReflect.FieldByName("Params")
	if !paramsField.IsValid() {
		return []string{fmt.Sprintf("discriminated type %T has no Params field", typedVal)}
	}

	paramsType := paramsField.Type()

	// Marshal practitioner params for strict re-decode.
	rawParams, err := json.Marshal(params)
	if err != nil {
		return []string{fmt.Sprintf("failed to marshal params for validation: %v", err)}
	}

	// Strip globally allowed keys before strict decode.
	allowedKeys := ruleTypeAdditionalAllowedParamsKeys[ruleTypeID]
	if len(allowedKeys) > 0 {
		rawParams, err = stripKeys(rawParams, allowedKeys)
		if err != nil {
			return []string{fmt.Sprintf("failed to pre-process params for validation: %v", err)}
		}
	}

	// Create a fresh target for strict decode.
	var paramsTarget any
	if paramsType.Kind() == reflect.Pointer {
		paramsTarget = reflect.New(paramsType.Elem()).Interface()
	} else {
		paramsTarget = reflect.New(paramsType).Interface()
	}

	decoder := json.NewDecoder(bytes.NewReader(rawParams))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(paramsTarget); err != nil {
		return []string{fmt.Sprintf("params did not match schema for rule type %q: %v", ruleTypeID, err)}
	}

	// Required-keys heuristic: create a fresh zero value to discover
	// non-omitempty fields.
	var paramsZero any
	if paramsType.Kind() == reflect.Pointer {
		paramsZero = reflect.New(paramsType.Elem()).Interface()
	} else {
		paramsZero = reflect.New(paramsType).Interface()
	}

	requiredKeys, err := computeRequiredKeys(paramsZero)
	if err != nil {
		return []string{fmt.Sprintf("failed to compute required keys for rule type %q: %v", ruleTypeID, err)}
	}

	missingKeys := missingRequiredKeys(requiredKeys, params, ruleTypeAdditionalRequiredParamsKeys[ruleTypeID])
	if len(missingKeys) > 0 {
		return []string{fmt.Sprintf("missing required params keys for rule type %q: %s", ruleTypeID, strings.Join(missingKeys, ", "))}
	}

	postDecodeErrs := validateRuleParamsPostDecode(ruleTypeID, params)
	if len(postDecodeErrs) > 0 {
		return postDecodeErrs
	}

	return nil
}

var ruleTypeAdditionalRequiredParamsKeys = map[string][]string{
	// Kibana rejects `.es-query` params without `size` even when the generated
	// DSL variant currently models it as optional.
	ruleTypeESQuery: {paramsKeySize},
}

func validateRuleParams(ruleTypeID string, params map[string]any) []string {
	specs, ok := ruleTypeParamsOverrides[ruleTypeID]
	if !ok {
		return validateParamsViaDiscriminator(ruleTypeID, params)
	}

	raw, err := json.Marshal(params)
	if err != nil {
		return []string{fmt.Sprintf("failed to marshal params for validation: %v", err)}
	}

	var best validationCandidate
	validationRaw, err := stripKeys(raw, ruleTypeAdditionalAllowedParamsKeys[ruleTypeID])
	if err != nil {
		return []string{fmt.Sprintf("failed to pre-process params for validation: %v", err)}
	}
	for _, spec := range specs {
		specRaw := validationRaw
		if len(spec.additionalAllowedKeys) > 0 {
			specRaw, err = stripKeys(specRaw, spec.additionalAllowedKeys)
			if err != nil {
				return []string{fmt.Sprintf("failed to pre-process params for validation: %v", err)}
			}
		}
		target := spec.newTarget()
		decoder := json.NewDecoder(bytes.NewReader(specRaw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			best.consider(false, fmt.Sprintf("params did not match %s schema for rule type %q: %v", spec.name, ruleTypeID, err))
			continue
		}

		missingKeys := missingRequiredKeys(spec.requiredKeys, params, ruleTypeAdditionalRequiredParamsKeys[ruleTypeID])
		if len(missingKeys) == 0 {
			postDecodeErrs := validateRuleParamsPostDecode(ruleTypeID, params)
			if len(postDecodeErrs) == 0 {
				return nil
			}

			best.consider(true, formatParamsValidationErrors(postDecodeErrs))
			continue
		}

		best.consider(true, fmt.Sprintf("missing required params keys for rule type %q: %s", ruleTypeID, strings.Join(missingKeys, ", ")))
	}

	if !best.hasValue || best.err == "" {
		return nil
	}

	return []string{best.err}
}

func validateRuleParamsPostDecode(ruleTypeID string, params map[string]any) []string {
	if ruleTypeID == ruleTypeIndexThreshold {
		if index, ok := params["index"]; ok && !isJSONArrayLike(index) {
			return []string{fmt.Sprintf("invalid params for rule type %q: index must be an array of strings", ruleTypeID)}
		}
	}

	return nil
}

type validationCandidate struct {
	hasValue bool
	decoded  bool
	err      string
}

func (c *validationCandidate) consider(decoded bool, err string) {
	if !c.hasValue || betterValidationCandidate(decoded, c.decoded) {
		c.hasValue = true
		c.decoded = decoded
		c.err = err
	}
}

func betterValidationCandidate(decoded bool, currentDecoded bool) bool {
	if decoded != currentDecoded {
		// Prefer candidates that decoded successfully so key-level diagnostics
		// are surfaced over generic schema-mismatch decode errors.
		return decoded
	}
	// Keep stable variant order for deterministic tie-breaking.
	return false
}

func isJSONArrayLike(v any) bool {
	if v == nil {
		return false
	}

	kind := reflect.TypeOf(v).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func missingRequiredKeys(requiredKeys map[string]struct{}, params map[string]any, additionalRequiredKeys []string) []string {
	if len(requiredKeys) == 0 && len(additionalRequiredKeys) == 0 {
		return nil
	}

	allRequired := make(map[string]struct{}, len(requiredKeys)+len(additionalRequiredKeys))
	for key := range requiredKeys {
		allRequired[key] = struct{}{}
	}
	for _, key := range additionalRequiredKeys {
		allRequired[key] = struct{}{}
	}

	missing := make([]string, 0)
	for key := range allRequired {
		if _, ok := params[key]; !ok {
			missing = append(missing, key)
		}
	}

	slices.Sort(missing)
	return missing
}

func computeRequiredKeys(target any) (map[string]struct{}, error) {
	// Marshal zero-value struct and decode to map to discover non-omitempty keys.
	// This relies on the Go JSON serialization convention: fields tagged with
	// `omitempty` are omitted when they hold their zero value, so any key that
	// survives marshaling a zero-value struct is treated as required. This is a
	// heuristic — if a generated type has a non-pointer field that Kibana treats
	// as optional, it will appear "required" here. Use ruleTypeAdditionalRequiredParamsKeys
	// (or ruleTypeAdditionalAllowedParamsKeys) to patch individual mismatches.
	raw, err := json.Marshal(target)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	if values == nil {
		return nil, fmt.Errorf("decoded JSON was null (expected object)")
	}

	requiredKeys := make(map[string]struct{}, len(values))
	for key := range values {
		requiredKeys[key] = struct{}{}
	}
	return requiredKeys, nil
}

func stripKeys(raw []byte, keys []string) ([]byte, error) {
	if len(keys) == 0 {
		return raw, nil
	}

	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	for _, key := range keys {
		delete(values, key)
	}

	stripped, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	return stripped, nil
}

func formatParamsValidationErrors(errs []string) string {
	return strings.Join(errs, "\n")
}

const frequencyExclusivityDetail = "Rule-level notify_when and throttle cannot be combined with actions[*].frequency " +
	"(per-action notification). Use either rule-level notify_when/throttle or per-action frequency blocks, not both. " +
	"Kibana does not allow these parameters when notify_when or throttle are defined at the rule level."

func validateNotifyWhenThrottleFrequencyExclusivity(ctx context.Context, data *alertingRuleModel, diags *diag.Diagnostics) {
	if !configActionsIncludeKnownFrequencyBlock(ctx, data.Actions, diags) {
		return
	}
	ruleNotify := ruleLevelNotifyWhenExclusive(data.NotifyWhen)
	ruleThrottle := ruleLevelThrottleExclusive(data.Throttle.StringValue)
	if !ruleNotify && !ruleThrottle {
		return
	}
	if ruleNotify {
		diags.AddAttributeError(path.Root("notify_when"), "Cannot combine rule-level notify_when with actions.frequency", frequencyExclusivityDetail)
		return
	}
	diags.AddAttributeError(path.Root("throttle"), "Cannot combine rule-level throttle with actions.frequency", frequencyExclusivityDetail)
}

func ruleLevelNotifyWhenExclusive(v types.String) bool {
	return typeutils.IsKnown(v) && !v.IsNull() && strings.TrimSpace(v.ValueString()) != ""
}

func ruleLevelThrottleExclusive(v basetypes.StringValue) bool {
	return typeutils.IsKnown(v) && !v.IsNull() && strings.TrimSpace(v.ValueString()) != ""
}

func configActionsIncludeKnownFrequencyBlock(ctx context.Context, actions types.List, diags *diag.Diagnostics) bool {
	if !typeutils.IsKnown(actions) || actions.IsNull() {
		return false
	}
	var elems []actionModel
	diags.Append(actions.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return false
	}
	for _, a := range elems {
		if typeutils.IsKnown(a.Frequency) && !a.Frequency.IsNull() {
			return true
		}
	}
	return false
}
