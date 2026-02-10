package alerting_rule

import (
	"strings"
	"testing"
)

func TestValidateRuleParamsIndexThreshold(t *testing.T) {
	valid := map[string]interface{}{
		"index":               []interface{}{"logs-*"},
		"threshold":           []interface{}{1.0},
		"thresholdComparator": ">",
		"timeField":           "@timestamp",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	if errs := validateRuleParams(".index-threshold", valid); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}

	invalid := map[string]interface{}{
		"index":               "logs-*",
		"threshold":           []interface{}{1.0},
		"thresholdComparator": ">",
		"timeField":           "@timestamp",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	if errs := validateRuleParams(".index-threshold", invalid); len(errs) == 0 {
		t.Fatalf("expected validation errors for invalid index-threshold params")
	}
}

func TestValidateRuleParamsEsQueryUnion(t *testing.T) {
	validKQL := map[string]interface{}{
		"searchType":          "searchSource",
		"size":                0.0,
		"threshold":           []interface{}{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	if errs := validateRuleParams(".es-query", validKQL); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}

	invalid := map[string]interface{}{
		"searchType":          "searchSource",
		"size":                "not-a-number",
		"threshold":           []interface{}{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	if errs := validateRuleParams(".es-query", invalid); len(errs) == 0 {
		t.Fatalf("expected validation errors for invalid es-query params")
	}
}

func TestValidateRuleParamsUnknownRuleTypeIsAllowed(t *testing.T) {
	params := map[string]interface{}{
		"anything": "goes",
	}

	if errs := validateRuleParams("custom.rule.type", params); len(errs) > 0 {
		t.Fatalf("expected unknown rule type to skip validation, got: %v", errs)
	}
}

func TestValidateRuleParamsSyntheticsMonitorStatusRequiredFields(t *testing.T) {
	invalid := map[string]interface{}{
		"numTimes": 1.0,
	}

	if errs := validateRuleParams("xpack.uptime.alerts.monitorStatus", invalid); len(errs) == 0 {
		t.Fatalf("expected missing field errors for synthetics monitor status")
	}

	valid := map[string]interface{}{
		"numTimes":                1.0,
		"shouldCheckStatus":       true,
		"shouldCheckAvailability": false,
	}

	if errs := validateRuleParams("xpack.uptime.alerts.monitorStatus", valid); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}
}

func TestValidateRuleParamsApmAnomalyRequiredKeys(t *testing.T) {
	valid := map[string]interface{}{
		"windowSize":          5.0,
		"windowUnit":          "m",
		"environment":         "production",
		"anomalySeverityType": "critical",
	}

	if errs := validateRuleParams("apm.rules.anomaly", valid); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}

	invalid := map[string]interface{}{
		"windowSize":  5.0,
		"windowUnit":  "m",
		"environment": "production",
	}

	if errs := validateRuleParams("apm.rules.anomaly", invalid); len(errs) == 0 {
		t.Fatalf("expected required field validation errors for apm anomaly params")
	}
}

func TestValidateRuleParamsRejectsUnexpectedKeys(t *testing.T) {
	params := map[string]interface{}{
		"windowSize":          5.0,
		"windowUnit":          "m",
		"environment":         "production",
		"anomalySeverityType": "critical",
		"extraParam":          true,
	}

	errs := validateRuleParams("apm.rules.anomaly", params)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for unexpected params key")
	}
	if !strings.Contains(strings.Join(errs, "; "), "unexpected params keys: extraParam") {
		t.Fatalf("expected unexpected key error, got: %v", errs)
	}
}

func TestValidateRuleParamsSloBurnRateAllowsWindows(t *testing.T) {
	params := map[string]interface{}{
		"sloId":        "o11y_managed_o11y-search-success-rat",
		"dependencies": []interface{}{},
		"windows": []interface{}{
			map[string]interface{}{
				"id":                   "ede70e84-ff91-4f69-9f1e-558e45737998",
				"burnRateThreshold":    14.4,
				"maxBurnRateThreshold": 168.0,
				"longWindow": map[string]interface{}{
					"value": 1.0,
					"unit":  "h",
				},
				"shortWindow": map[string]interface{}{
					"value": 5.0,
					"unit":  "m",
				},
				"actionGroup": "slo.burnRate.alert",
			},
		},
	}

	if errs := validateRuleParams("slo.rules.burnRate", params); len(errs) > 0 {
		t.Fatalf("expected no validation errors for slo burn rate windows payload, got: %v", errs)
	}
}

func TestValidateRuleParamsEsQueryAllowsSourceFields(t *testing.T) {
	params := map[string]interface{}{
		"searchType":          "searchSource",
		"size":                0.0,
		"threshold":           []interface{}{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
		"sourceFields": []interface{}{
			map[string]interface{}{
				"label":      "cluster_id",
				"searchPath": "cluster_id",
			},
		},
	}

	if errs := validateRuleParams(".es-query", params); len(errs) > 0 {
		t.Fatalf("expected no validation errors for es-query sourceFields payload, got: %v", errs)
	}
}
