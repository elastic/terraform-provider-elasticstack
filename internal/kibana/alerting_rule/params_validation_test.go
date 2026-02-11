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
	if !strings.Contains(strings.Join(errs, "; "), "json: unknown field \"extraParam\"") {
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

func TestValidateRuleParamsEsQueryRequiresSize(t *testing.T) {
	params := map[string]interface{}{
		"searchType":          "searchSource",
		"threshold":           []interface{}{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	errs := validateRuleParams(".es-query", params)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for missing es-query size")
	}
	if !strings.Contains(strings.Join(errs, "; "), "missing required params keys:") || !strings.Contains(strings.Join(errs, "; "), "size") {
		t.Fatalf("expected missing size error, got: %v", errs)
	}
}

func TestValidateRuleParamsSloBurnRateStillRejectsUnknownExtraKeys(t *testing.T) {
	params := map[string]interface{}{
		"sloId":        "o11y_managed_o11y-search-success-rat",
		"dependencies": []interface{}{},
		"windows":      []interface{}{},
		"unexpected":   true,
	}

	errs := validateRuleParams("slo.rules.burnRate", params)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for unexpected slo burn rate key")
	}
	if !strings.Contains(strings.Join(errs, "; "), "json: unknown field \"unexpected\"") {
		t.Fatalf("expected unexpected key error, got: %v", errs)
	}
}

func TestValidateRuleParamsIndexThresholdRejectsSourceFields(t *testing.T) {
	params := map[string]interface{}{
		"index":               []interface{}{"logs-*"},
		"threshold":           []interface{}{1.0},
		"thresholdComparator": ">",
		"timeField":           "@timestamp",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
		"sourceFields": []interface{}{
			map[string]interface{}{
				"label":      "cluster_id",
				"searchPath": "cluster_id",
			},
		},
	}

	errs := validateRuleParams(".index-threshold", params)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for sourceFields on non-es-query rule")
	}
	if !strings.Contains(strings.Join(errs, "; "), "json: unknown field \"sourceFields\"") {
		t.Fatalf("expected sourceFields unexpected key error, got: %v", errs)
	}
}

func TestValidationCandidatePrefersDecodedOverDecodeFailure(t *testing.T) {
	candidate := validationCandidate{}
	candidate.consider(false, []string{"params do not match expected generated schema: bad type"})
	candidate.consider(true, []string{"missing required params keys: query"})

	if len(candidate.errs) != 1 {
		t.Fatalf("expected single decoded error to win, got: %v", candidate.errs)
	}
	if candidate.errs[0] != "missing required params keys: query" {
		t.Fatalf("expected decoded candidate to be selected, got: %v", candidate.errs)
	}
}

func TestValidationCandidateKeepsStableOrderOnTie(t *testing.T) {
	candidate := validationCandidate{}
	candidate.consider(true, []string{"missing required params keys: a"})
	candidate.consider(true, []string{"missing required params keys: b"})

	if len(candidate.errs) != 1 {
		t.Fatalf("expected one error, got: %v", candidate.errs)
	}
	if candidate.errs[0] != "missing required params keys: a" {
		t.Fatalf("expected first candidate to win tie, got: %v", candidate.errs)
	}
}

func TestFormatParamsValidationErrorsMultiline(t *testing.T) {
	formatted := formatParamsValidationErrors([]string{
		"missing required params keys: threshold",
		"unexpected params keys: extraParam",
	})

	expected := "missing required params keys: threshold\nunexpected params keys: extraParam"
	if formatted != expected {
		t.Fatalf("expected %q, got %q", expected, formatted)
	}
}

func TestValidateRuleParamsFixturesFromSreO11yModules(t *testing.T) {
	testCases := []struct {
		name      string
		ruleType  string
		params    map[string]interface{}
		expectErr string
	}{
		{
			name:     "es-query high disk watermark valid fixture",
			ruleType: ".es-query",
			params: map[string]interface{}{
				"aggType":                    "count",
				"groupBy":                    "top",
				"termSize":                   10.0,
				"termField":                  "docker.container.labels.co.elastic.cloud.allocator.deployment_id",
				"timeWindowSize":             30.0,
				"timeWindowUnit":             "m",
				"threshold":                  []interface{}{30.0},
				"thresholdComparator":        ">",
				"index":                      []interface{}{"cluster-elasticsearch-*"},
				"timeField":                  "@timestamp",
				"searchType":                 "esQuery",
				"size":                       10.0,
				"esQuery":                    "{\"query\":{\"bool\":{\"must\":[]}}}",
				"excludeHitsFromPreviousRun": true,
			},
		},
		{
			name:     "es-query autoscaling valid fixture with sourceFields",
			ruleType: ".es-query",
			params: map[string]interface{}{
				"aggType":                    "count",
				"esQuery":                    "{\"query\":{\"bool\":{\"filter\":[]}}}",
				"excludeHitsFromPreviousRun": false,
				"groupBy":                    "top",
				"index":                      []interface{}{"logging-*:service-constructor-*"},
				"searchType":                 "esQuery",
				"size":                       1.0,
				"sourceFields": []interface{}{
					map[string]interface{}{"label": "cluster_id", "searchPath": "cluster_id"},
				},
				"termField":           "cluster_id",
				"termSize":            100.0,
				"threshold":           []interface{}{1.0},
				"thresholdComparator": ">",
				"timeField":           "@timestamp",
				"timeWindowSize":      5.0,
				"timeWindowUnit":      "m",
			},
		},
		{
			name:     "es-query failed rule evaluations valid fixture",
			ruleType: ".es-query",
			params: map[string]interface{}{
				"aggType":                    "count",
				"groupBy":                    "top",
				"termSize":                   10.0,
				"termField":                  "rule.id",
				"size":                       100.0,
				"timeWindowSize":             6.0,
				"timeWindowUnit":             "h",
				"threshold":                  []interface{}{3.0},
				"thresholdComparator":        ">",
				"index":                      []interface{}{".ds-.kibana-event-log*"},
				"timeField":                  "@timestamp",
				"searchType":                 "esQuery",
				"esQuery":                    "{\"query\":{\"bool\":{\"must\":[]}}}",
				"excludeHitsFromPreviousRun": false,
			},
		},
		{
			name:     "es-query flood stage invalid fixture catches unknown key",
			ruleType: ".es-query",
			params: map[string]interface{}{
				"aggType":                    "count",
				"groupBy":                    "top",
				"termSize":                   10.0,
				"termField":                  "docker.container.labels.co.elastic.cloud.allocator.deployment_id",
				"timeWindowSize":             10.0,
				"timeWindowUnit":             "m",
				"threshold":                  []interface{}{10.0},
				"thresholdComparator":        ">",
				"index":                      []interface{}{"cluster-elasticsearch-*"},
				"timeField":                  "@timestamp",
				"searchType":                 "esQuery",
				"size":                       10.0,
				"esQuery":                    "{\"query\":{\"bool\":{\"must\":[]}}}",
				"excludeHitsFromPreviousRun": true,
				"hi":                         "hi",
			},
			expectErr: "json: unknown field \"hi\"",
		},
		{
			name:     "slo burn rate valid fixture with dependencies",
			ruleType: "slo.rules.burnRate",
			params: map[string]interface{}{
				"sloId": "abc123",
				"windows": []interface{}{
					map[string]interface{}{
						"id":                   "0c59b724-200b-462f-928c-d975e69b1eef",
						"burnRateThreshold":    3.36,
						"maxBurnRateThreshold": 168.0,
						"longWindow":           map[string]interface{}{"value": 1.0, "unit": "h"},
						"shortWindow":          map[string]interface{}{"value": 5.0, "unit": "m"},
						"actionGroup":          "slo.burnRate.alert",
					},
					map[string]interface{}{
						"id":                   "62770ca9-c0f9-4a9a-bb1c-3a9666f54cf7",
						"burnRateThreshold":    1.4,
						"maxBurnRateThreshold": 28.0,
						"longWindow":           map[string]interface{}{"value": 6.0, "unit": "h"},
						"shortWindow":          map[string]interface{}{"value": 30.0, "unit": "m"},
						"actionGroup":          "slo.burnRate.high",
					},
				},
				"dependencies": []interface{}{},
			},
		},
		{
			name:     "uptime monitor status valid fixture",
			ruleType: "xpack.uptime.alerts.monitorStatus",
			params: map[string]interface{}{
				"search":                  "",
				"numTimes":                8.0,
				"timerangeUnit":           "m",
				"timerangeCount":          10.0,
				"shouldCheckStatus":       true,
				"shouldCheckAvailability": false,
				"availability": map[string]interface{}{
					"range":     30.0,
					"rangeUnit": "d",
					"threshold": "99",
				},
				"filters": map[string]interface{}{
					"tags": []interface{}{"o11y"},
				},
			},
		},
		{
			name:     "unknown custom threshold from modules remains pass through",
			ruleType: "observability.rules.custom_threshold",
			params: map[string]interface{}{
				"criteria":      []interface{}{},
				"alertOnNoData": true,
			},
		},
		{
			name:     "unknown transform_health from modules remains pass through",
			ruleType: "transform_health",
			params: map[string]interface{}{
				"transforms": []interface{}{"foo-transform"},
				"unhealthy":  true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateRuleParams(tc.ruleType, tc.params)
			if tc.expectErr == "" {
				if len(errs) > 0 {
					t.Fatalf("expected no validation errors, got: %v", errs)
				}
				return
			}

			if len(errs) == 0 {
				t.Fatalf("expected validation errors containing %q", tc.expectErr)
			}
			if !strings.Contains(strings.Join(errs, "; "), tc.expectErr) {
				t.Fatalf("expected error containing %q, got: %v", tc.expectErr, errs)
			}
		})
	}
}

func TestValidateRuleParamsFixturesFromClusterMgmtCustomers(t *testing.T) {
	testCases := []struct {
		name      string
		ruleType  string
		params    map[string]interface{}
		expectErr string
	}{
		{
			name:     "metrics alert threshold k8s node disk pressure",
			ruleType: "metrics.alert.threshold",
			params: map[string]interface{}{
				"criteria": []interface{}{
					map[string]interface{}{
						"aggType":    "count",
						"comparator": ">",
						"threshold":  []interface{}{30.0},
						"timeSize":   15.0,
						"timeUnit":   "m",
					},
				},
				"sourceId":              "default",
				"alertOnNoData":         false,
				"alertOnGroupDisappear": false,
				"filterQueryText":       "kubernetes.node.status.disk_pressure: true and orchestrator.platform.type: mki",
				"filterQuery":           "{\"bool\":{\"filter\":[]}}",
				"groupBy":               []interface{}{"orchestrator.cluster.name", "kubernetes.node.name"},
			},
		},
		{
			name:     "logs document count project api 500",
			ruleType: "logs.alert.document.count",
			params: map[string]interface{}{
				"timeSize": 10.0,
				"timeUnit": "m",
				"logView": map[string]interface{}{
					"type":      "log-view-reference",
					"logViewId": "default",
				},
				"count": map[string]interface{}{
					"value":      1.0,
					"comparator": "more than",
				},
				"criteria": []interface{}{
					map[string]interface{}{"field": "kubernetes.container.name", "comparator": "equals", "value": "project-api"},
					map[string]interface{}{"field": "http.response.status_code", "comparator": "equals", "value": 500.0},
					map[string]interface{}{"field": "message", "comparator": "matches", "value": "\"HTTP request\""},
				},
			},
		},
		{
			name:     "apm error rate cosmos throttling",
			ruleType: "apm.error_rate",
			params: map[string]interface{}{
				"environment": "production",
				"searchConfiguration": map[string]interface{}{
					"query": map[string]interface{}{
						"language": "kuery",
						"query":    "(service.name : \"project-api\") and error.exception.type : \"TooManyRequestsError\"",
					},
				},
				"threshold":    10.0,
				"useKqlFilter": true,
				"windowSize":   5.0,
				"windowUnit":   "m",
			},
		},
		{
			name:     "apm transaction error rate uiam authenticate",
			ruleType: "apm.transaction_error_rate",
			params: map[string]interface{}{
				"environment":     "ENVIRONMENT_ALL",
				"serviceName":     "uiam",
				"transactionType": "request",
				"transactionName": "POST /v1/authentication/_authenticate",
				"threshold":       0.0001,
				"windowSize":      5.0,
				"windowUnit":      "m",
			},
		},
		{
			name:     "index threshold kibana slo no data",
			ruleType: ".index-threshold",
			params: map[string]interface{}{
				"aggType":             "count",
				"filterKuery":         "data_stream.dataset: \"proxy.log\"",
				"groupBy":             "all",
				"index":               []interface{}{"proxy-logs-*"},
				"termSize":            5.0,
				"threshold":           []interface{}{0.0},
				"thresholdComparator": "<=",
				"timeField":           "@timestamp",
				"timeWindowSize":      10.0,
				"timeWindowUnit":      "m",
			},
		},
		{
			name:     "uptime monitor status with stackVersion",
			ruleType: "xpack.uptime.alerts.monitorStatus",
			params: map[string]interface{}{
				"search":                  "monitor.name: \"Production Backstage Monitor\"",
				"numTimes":                5.0,
				"timerangeUnit":           "m",
				"timerangeCount":          5.0,
				"shouldCheckStatus":       true,
				"shouldCheckAvailability": false,
				"availability": map[string]interface{}{
					"range":     30.0,
					"rangeUnit": "d",
					"threshold": "99",
				},
				"stackVersion": "9.2.2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateRuleParams(tc.ruleType, tc.params)
			if tc.expectErr == "" {
				if len(errs) > 0 {
					t.Fatalf("expected no validation errors, got: %v", errs)
				}
				return
			}

			if len(errs) == 0 {
				t.Fatalf("expected validation errors containing %q", tc.expectErr)
			}
			if !strings.Contains(strings.Join(errs, "; "), tc.expectErr) {
				t.Fatalf("expected error containing %q, got: %v", tc.expectErr, errs)
			}
		})
	}
}
