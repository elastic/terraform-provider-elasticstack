package alerting_rule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// paramsSchemaSpec contains precomputed key metadata and decode factory for
// one generated params variant. This avoids reflection in runtime validation.
type paramsSchemaSpec struct {
	name         string
	newTarget    func() any
	requiredKeys map[string]struct{}
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

	if !utils.IsKnown(data.Params) || !utils.IsKnown(data.RuleTypeID) {
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
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyApmAnomaly{} }),
	},
	"apm.error_rate": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyApmErrorCount{} }),
	},
	"apm.transaction_duration": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyApmTransactionDuration{} }),
	},
	"apm.transaction_error_rate": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyApmTransactionErrorRate{} }),
	},
	".index-threshold": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsIndexThresholdRule{} }),
	},
	"metrics.alert.inventory.threshold": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyInfraInventory{} }),
	},
	"metrics.alert.threshold": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyInfraMetricThreshold{} }),
	},
	"slo.rules.burnRate": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertySloBurnRate{} }),
	},
	"xpack.uptime.alerts.tls": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertySyntheticsUptimeTls{} }),
	},
	"xpack.uptime.alerts.monitorStatus": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertySyntheticsMonitorStatus{} }),
	},
	".es-query": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsEsQueryDslRule{} }),
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsEsQueryEsqlRule{} }),
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsEsQueryKqlRule{} }),
	},
	"logs.alert.document.count": {
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyLogThreshold0{} }),
		mustNewParamsSchemaSpec(func() any { return &kbapi.ParamsPropertyLogThreshold1{} }),
	},
}

var ruleTypeAdditionalAllowedParamsKeys = map[string][]string{
	// The generated type currently models legacy single-window fields, while
	// Kibana accepts modern multi-window payloads under `windows`.
	// TODO: remove when upstream Kibana schema models modern window payloads.
	// Tracking: https://github.com/elastic/kibana/issues/252451
	"slo.rules.burnRate": {"windows", "dependencies"},
	// Kibana supports passing selected hit fields to actions, but that key is
	// currently missing from generated `.es-query` params models.
	// TODO: remove when upstream Kibana schema includes this key.
	// Tracking: https://github.com/elastic/kibana/issues/252451
	".es-query": {"sourceFields"},
	// Kibana accepts this convenience field alongside filterQuery in metrics
	// threshold rules.
	"metrics.alert.threshold": {"filterQueryText"},
	// Kibana accepts these APM error-rate params, but generated schema currently
	// misses them in provider validation targets.
	"apm.error_rate": {"searchConfiguration", "useKqlFilter"},
	// Kibana accepts stackVersion in uptime monitorStatus params.
	"xpack.uptime.alerts.monitorStatus": {"stackVersion"},
}

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
		target := spec.newTarget()
		decoder := json.NewDecoder(bytes.NewReader(validationRaw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			best.consider(false, fmt.Sprintf("extra param detected in params field for rule type %q: %v", ruleTypeID, err))
			continue
		}

		missingKeys := missingRequiredKeys(spec.requiredKeys, params, ruleTypeAdditionalRequiredParamsKeys[ruleTypeID])
		if len(missingKeys) == 0 {
			return nil
		}

		best.consider(true, fmt.Sprintf("missing required params keys for rule type %q: %s", ruleTypeID, strings.Join(missingKeys, ", ")))
	}

	if !best.hasValue || best.err == "" {
		return nil
	}

	return []string{best.err}
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
