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
		strings.Join(errs, "; "),
	)
}

var ruleTypeParamsTargets = map[string][]func() interface{}{
	"apm.rules.anomaly":                 {func() interface{} { return &kbapi.ParamsPropertyApmAnomaly{} }},
	"apm.error_rate":                    {func() interface{} { return &kbapi.ParamsPropertyApmErrorCount{} }},
	"apm.transaction_duration":          {func() interface{} { return &kbapi.ParamsPropertyApmTransactionDuration{} }},
	"apm.transaction_error_rate":        {func() interface{} { return &kbapi.ParamsPropertyApmTransactionErrorRate{} }},
	".index-threshold":                  {func() interface{} { return &kbapi.ParamsIndexThresholdRule{} }},
	"metrics.alert.inventory.threshold": {func() interface{} { return &kbapi.ParamsPropertyInfraInventory{} }},
	"metrics.alert.threshold":           {func() interface{} { return &kbapi.ParamsPropertyInfraMetricThreshold{} }},
	"slo.rules.burnRate":                {func() interface{} { return &kbapi.ParamsPropertySloBurnRate{} }},
	"xpack.uptime.alerts.tls":           {func() interface{} { return &kbapi.ParamsPropertySyntheticsUptimeTls{} }},
	"xpack.uptime.alerts.monitorStatus": {func() interface{} { return &kbapi.ParamsPropertySyntheticsMonitorStatus{} }},
	".es-query": {
		func() interface{} { return &kbapi.ParamsEsQueryDslRule{} },
		func() interface{} { return &kbapi.ParamsEsQueryEsqlRule{} },
		func() interface{} { return &kbapi.ParamsEsQueryKqlRule{} },
	},
	"logs.alert.document.count": {
		func() interface{} { return &kbapi.ParamsPropertyLogThreshold0{} },
		func() interface{} { return &kbapi.ParamsPropertyLogThreshold1{} },
	},
}

func validateRuleParams(ruleTypeID string, params map[string]interface{}) []string {
	targets, ok := ruleTypeParamsTargets[ruleTypeID]
	if !ok {
		// Backward compatible fallback for custom/unknown rules.
		return nil
	}

	raw, err := json.Marshal(params)
	if err != nil {
		return []string{fmt.Sprintf("failed to marshal params for validation: %v", err)}
	}

	var best []string
	for _, newTarget := range targets {
		target := newTarget()
		if err := json.Unmarshal(raw, target); err != nil {
			candidateErrs := []string{fmt.Sprintf("params do not match expected generated schema: %v", err)}
			if best == nil || len(candidateErrs) < len(best) {
				best = candidateErrs
			}
			continue
		}

		missingKeys := requiredKeysMissing(target, params)
		if len(missingKeys) == 0 {
			return nil
		}

		candidateErrs := []string{fmt.Sprintf("missing required params keys: %s", strings.Join(missingKeys, ", "))}
		if best == nil || len(candidateErrs) < len(best) {
			best = candidateErrs
		}
	}

	return best
}

func requiredKeysMissing(target interface{}, params map[string]interface{}) []string {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	missing := make([]string, 0)
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

		// Required keys are represented as non-pointer fields without omitempty.
		if jsonHasOmitEmpty || field.Type.Kind() == reflect.Pointer {
			continue
		}

		if _, ok := params[jsonName]; !ok {
			missing = append(missing, jsonName)
		}
	}

	slices.Sort(missing)
	return missing
}

func parseJSONTag(tag string) (name string, hasOmitEmpty bool) {
	if tag == "" {
		return "", false
	}

	parts := strings.Split(tag, ",")
	name = parts[0]
	for _, part := range parts[1:] {
		if part == "omitempty" {
			hasOmitEmpty = true
			break
		}
	}
	return name, hasOmitEmpty
}
