package alerting_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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
	newTarget    func() interface{}
	requiredKeys map[string]struct{}
	allowedKeys  map[string]struct{}
}

func mustNewParamsSchemaSpec(name string, newTarget func() interface{}) paramsSchemaSpec {
	requiredKeys, allowedKeys := paramsSchemaKeys(newTarget())
	return paramsSchemaSpec{
		name:         name,
		newTarget:    newTarget,
		requiredKeys: requiredKeys,
		allowedKeys:  allowedKeys,
	}
}

func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data alertingRuleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.IsKnown(data.Params) || !utils.IsKnown(data.RuleTypeID) {
		return
	}

	var params map[string]interface{}
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
		"Invalid params for rule_type_id",
		formatParamsValidationErrors(errs),
	)
}

var ruleTypeParamsSpecs = map[string][]paramsSchemaSpec{
	"apm.rules.anomaly": {
		mustNewParamsSchemaSpec("params_property_apm_anomaly", func() interface{} { return &kbapi.ParamsPropertyApmAnomaly{} }),
	},
	"apm.error_rate": {
		mustNewParamsSchemaSpec("params_property_apm_error_count", func() interface{} { return &kbapi.ParamsPropertyApmErrorCount{} }),
	},
	"apm.transaction_duration": {
		mustNewParamsSchemaSpec("params_property_apm_transaction_duration", func() interface{} { return &kbapi.ParamsPropertyApmTransactionDuration{} }),
	},
	"apm.transaction_error_rate": {
		mustNewParamsSchemaSpec("params_property_apm_transaction_error_rate", func() interface{} { return &kbapi.ParamsPropertyApmTransactionErrorRate{} }),
	},
	".index-threshold": {
		mustNewParamsSchemaSpec("params_index_threshold_rule", func() interface{} { return &kbapi.ParamsIndexThresholdRule{} }),
	},
	"metrics.alert.inventory.threshold": {
		mustNewParamsSchemaSpec("params_property_infra_inventory", func() interface{} { return &kbapi.ParamsPropertyInfraInventory{} }),
	},
	"metrics.alert.threshold": {
		mustNewParamsSchemaSpec("params_property_infra_metric_threshold", func() interface{} { return &kbapi.ParamsPropertyInfraMetricThreshold{} }),
	},
	"slo.rules.burnRate": {
		mustNewParamsSchemaSpec("params_property_slo_burn_rate", func() interface{} { return &kbapi.ParamsPropertySloBurnRate{} }),
	},
	"xpack.uptime.alerts.tls": {
		mustNewParamsSchemaSpec("params_property_synthetics_uptime_tls", func() interface{} { return &kbapi.ParamsPropertySyntheticsUptimeTls{} }),
	},
	"xpack.uptime.alerts.monitorStatus": {
		mustNewParamsSchemaSpec("params_property_synthetics_monitor_status", func() interface{} { return &kbapi.ParamsPropertySyntheticsMonitorStatus{} }),
	},
	".es-query": {
		mustNewParamsSchemaSpec("params_es_query_dsl_rule", func() interface{} { return &kbapi.ParamsEsQueryDslRule{} }),
		mustNewParamsSchemaSpec("params_es_query_esql_rule", func() interface{} { return &kbapi.ParamsEsQueryEsqlRule{} }),
		mustNewParamsSchemaSpec("params_es_query_kql_rule", func() interface{} { return &kbapi.ParamsEsQueryKqlRule{} }),
	},
	"logs.alert.document.count": {
		mustNewParamsSchemaSpec("params_property_log_threshold_0", func() interface{} { return &kbapi.ParamsPropertyLogThreshold0{} }),
		mustNewParamsSchemaSpec("params_property_log_threshold_1", func() interface{} { return &kbapi.ParamsPropertyLogThreshold1{} }),
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

func validateRuleParams(ruleTypeID string, params map[string]interface{}) []string {
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
	for _, spec := range specs {
		target := spec.newTarget()
		if err := json.Unmarshal(raw, target); err != nil {
			best.consider(false, []string{fmt.Sprintf("params do not match expected generated schema %q: %v", spec.name, err)})
			continue
		}

		missingKeys := missingRequiredKeys(spec.requiredKeys, params, ruleTypeAdditionalRequiredParamsKeys[ruleTypeID])
		unexpectedKeys := unexpectedKeysPresent(spec.allowedKeys, params, ruleTypeAdditionalAllowedParamsKeys[ruleTypeID])
		if len(missingKeys) == 0 && len(unexpectedKeys) == 0 {
			return nil
		}

		candidateErrs := make([]string, 0, 2)
		if len(missingKeys) > 0 {
			candidateErrs = append(candidateErrs, fmt.Sprintf("missing required params keys: %s", strings.Join(missingKeys, ", ")))
		}
		if len(unexpectedKeys) > 0 {
			candidateErrs = append(candidateErrs, fmt.Sprintf("unexpected params keys: %s", strings.Join(unexpectedKeys, ", ")))
		}
		best.consider(true, candidateErrs)
	}

	return best.errs
}

type validationCandidate struct {
	hasValue bool
	decoded  bool
	errs     []string
}

func (c *validationCandidate) consider(decoded bool, errs []string) {
	if !c.hasValue || betterValidationCandidate(decoded, errs, c.decoded, c.errs) {
		c.hasValue = true
		c.decoded = decoded
		c.errs = errs
	}
}

func betterValidationCandidate(decoded bool, errs []string, currentDecoded bool, currentErrs []string) bool {
	if decoded != currentDecoded {
		// Prefer candidates that decoded successfully so key-level diagnostics
		// are surfaced over generic schema-mismatch decode errors.
		return decoded
	}
	if len(errs) != len(currentErrs) {
		return len(errs) < len(currentErrs)
	}
	// Keep stable variant order for deterministic tie-breaking.
	return false
}

func missingRequiredKeys(requiredKeys map[string]struct{}, params map[string]interface{}, additionalRequiredKeys []string) []string {
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

func unexpectedKeysPresent(allowedKeys map[string]struct{}, params map[string]interface{}, additionalAllowedKeys []string) []string {
	allowed := make(map[string]struct{}, len(allowedKeys)+len(additionalAllowedKeys))
	for key := range allowedKeys {
		allowed[key] = struct{}{}
	}
	for _, key := range additionalAllowedKeys {
		allowed[key] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil
	}

	unexpected := make([]string, 0)
	for key := range params {
		if _, ok := allowed[key]; !ok {
			unexpected = append(unexpected, key)
		}
	}

	slices.Sort(unexpected)
	return unexpected
}

func paramsSchemaKeys(target interface{}) (requiredKeys map[string]struct{}, allowedKeys map[string]struct{}) {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, nil
	}

	requiredKeys = make(map[string]struct{}, t.NumField())
	allowedKeys = make(map[string]struct{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported/synthetic fields (for example union backing fields).
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		jsonName, jsonHasOmitEmpty := parseJSONTag(jsonTag)
		if jsonName == "" || jsonName == "-" {
			continue
		}

		allowedKeys[jsonName] = struct{}{}

		// Required keys are represented as non-pointer fields without omitempty.
		if !jsonHasOmitEmpty && field.Type.Kind() != reflect.Pointer {
			requiredKeys[jsonName] = struct{}{}
		}
	}

	return requiredKeys, allowedKeys
}

func parseJSONTag(tag string) (name string, hasOmitEmpty bool) {
	if tag == "" {
		return "", false
	}

	parts := strings.Split(tag, ",")
	name = parts[0]
	if slices.Contains(parts[1:], "omitempty") {
		hasOmitEmpty = true
	}
	return name, hasOmitEmpty
}

func formatParamsValidationErrors(errs []string) string {
	return strings.Join(errs, "\n")
}
