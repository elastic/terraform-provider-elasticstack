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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
		Tags []string `json:"tags,omitempty"`
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

	var params map[string]any
	if err := json.Unmarshal([]byte(data.Params.ValueString()), &params); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("params"),
			"Invalid params JSON",
			err.Error(),
		)
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

var ruleTypeParamsSpecs = map[string][]paramsSchemaSpec{
	"apm.rules.anomaly": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsApmAnomalyCreateRuleBodyAlerting{} }),
	},
	"apm.error_rate": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsApmErrorRateCreateRuleBodyAlerting{} }),
	},
	"apm.transaction_duration": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsApmTransactionDurationCreateRuleBodyAlerting{} }),
	},
	"apm.transaction_error_rate": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsApmTransactionErrorRateCreateRuleBodyAlerting{} }),
	},
	".index-threshold": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsIndexThresholdCreateRuleBodyAlerting{} }),
	},
	"metrics.alert.inventory.threshold": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsMetricsAlertInventoryThresholdCreateRuleBodyAlerting{} }),
	},
	"metrics.alert.threshold": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsMetricsAlertThresholdCreateRuleBodyAlerting{} }),
	},
	"slo.rules.burnRate": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsSloRulesBurnrateCreateRuleBodyAlerting{} }),
	},
	"xpack.uptime.alerts.tls": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsXpackSyntheticsAlertsTlsCreateRuleBodyAlerting{} }),
	},
	"xpack.uptime.alerts.monitorStatus": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsXpackSyntheticsAlertsMonitorstatusCreateRuleBodyAlerting{} }),
		mustNewParamsSchemaSpec(func() any { return &legacyMonitorStatusParams{} }),
	},
	".es-query": {
		mustNewParamsSchemaSpecFromContainer(func() any { return &kbapi.KibanaHTTPAPIsEsQueryCreateRuleBodyAlerting{} }),
	},
	"logs.alert.document.count": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.KibanaHTTPAPIsLogsAlertDocumentCountCreateRuleBodyAlertingParams0{} }),
		mustNewParamsSchemaSpec(func() any { return &kbapi.KibanaHTTPAPIsLogsAlertDocumentCountCreateRuleBodyAlertingParams1{} }),
	},
}

var ruleTypeAdditionalAllowedParamsKeys = map[string][]string{}

var ruleTypeAdditionalRequiredParamsKeys = map[string][]string{
	// Kibana rejects `.es-query` params without `size` even when the generated
	// DSL variant currently models it as optional.
	".es-query": {"size"},
}

func validateRuleParams(ruleTypeID string, params map[string]any) []string {
	specs, ok := ruleTypeParamsSpecs[ruleTypeID]
	if !ok {
		// Backward compatible fallback for custom/unknown rules.
		return nil
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
	if ruleTypeID == ".index-threshold" {
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
