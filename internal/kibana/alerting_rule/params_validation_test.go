package alerting_rule

import "testing"

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
