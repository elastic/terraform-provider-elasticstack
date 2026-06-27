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
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kibanacustomtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRuleParamsIndexThreshold(t *testing.T) {
	valid := map[string]any{
		"index":               []any{"logs-*"},
		"threshold":           []any{1.0},
		"thresholdComparator": ">",
		"timeField":           "@timestamp",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	if errs := validateRuleParams(".index-threshold", valid); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}

	invalid := map[string]any{
		"index":               "logs-*",
		"threshold":           []any{1.0},
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
	validKQL := map[string]any{
		"searchType":          "searchSource",
		"size":                0.0,
		"threshold":           []any{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	if errs := validateRuleParams(".es-query", validKQL); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}

	invalid := map[string]any{
		"searchType":          "searchSource",
		"size":                "not-a-number",
		"threshold":           []any{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	if errs := validateRuleParams(".es-query", invalid); len(errs) == 0 {
		t.Fatalf("expected validation errors for invalid es-query params")
	}
}

func TestValidateRuleParamsUnknownRuleTypeIsAllowed(t *testing.T) {
	params := map[string]any{
		"anything": "goes",
	}

	if errs := validateRuleParams("custom.rule.type", params); len(errs) > 0 {
		t.Fatalf("expected unknown rule type to skip validation, got: %v", errs)
	}
}

func TestValidateRuleParamsSyntheticsMonitorStatusRequiredFields(t *testing.T) {
	invalid := map[string]any{
		"numTimes": 1.0,
	}

	if errs := validateRuleParams("xpack.uptime.alerts.monitorStatus", invalid); len(errs) == 0 {
		t.Fatalf("expected missing field errors for synthetics monitor status")
	}

	valid := map[string]any{
		"numTimes":                1.0,
		"shouldCheckStatus":       true,
		"shouldCheckAvailability": false,
	}

	if errs := validateRuleParams("xpack.uptime.alerts.monitorStatus", valid); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}
}

func TestValidateRuleParamsApmAnomalyRequiredKeys(t *testing.T) {
	valid := map[string]any{
		"windowSize":          5.0,
		"windowUnit":          "m",
		"environment":         "production",
		"anomalySeverityType": "critical",
	}

	if errs := validateRuleParams("apm.anomaly", valid); len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}

	invalid := map[string]any{
		"windowSize":  5.0,
		"windowUnit":  "m",
		"environment": "production",
	}

	if errs := validateRuleParams("apm.anomaly", invalid); len(errs) == 0 {
		t.Fatalf("expected required field validation errors for apm anomaly params")
	}
}

func TestValidateRuleParamsApmTransactionRulesAllowSearchConfiguration(t *testing.T) {
	testCases := []struct {
		name     string
		ruleType string
		params   map[string]any
	}{
		{
			name:     "transaction error rate",
			ruleType: "apm.transaction_error_rate",
			params: map[string]any{
				"environment": "ENVIRONMENT_ALL",
				"searchConfiguration": map[string]any{
					"query": map[string]any{
						"language": "kuery",
						"query":    "test*",
					},
				},
				"threshold":    30.0,
				"useKqlFilter": true,
				"windowSize":   5.0,
				"windowUnit":   "m",
			},
		},
		{
			name:     "transaction duration",
			ruleType: "apm.transaction_duration",
			params: map[string]any{
				"aggregationType": "avg",
				"environment":     "ENVIRONMENT_ALL",
				"searchConfiguration": map[string]any{
					"query": map[string]any{
						"language": "kuery",
						"query":    "test*",
					},
				},
				"threshold":    1500.0,
				"useKqlFilter": true,
				"windowSize":   5.0,
				"windowUnit":   "m",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if errs := validateRuleParams(tc.ruleType, tc.params); len(errs) > 0 {
				t.Fatalf("expected no validation errors, got: %v", errs)
			}
		})
	}
}

func TestValidateRuleParamsRejectsUnexpectedKeys(t *testing.T) {
	params := map[string]any{
		"windowSize":          5.0,
		"windowUnit":          "m",
		"environment":         "production",
		"anomalySeverityType": "critical",
		"extraParam":          true,
	}

	errs := validateRuleParams("apm.anomaly", params)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for unexpected params key")
	}
	if !strings.Contains(strings.Join(errs, "; "), "json: unknown field \"extraParam\"") {
		t.Fatalf("expected unexpected key error, got: %v", errs)
	}
}

func TestValidateRuleParamsSloBurnRateAllowsWindows(t *testing.T) {
	params := map[string]any{
		"sloId":        "o11y_managed_o11y-search-success-rat",
		"dependencies": []any{},
		"windows": []any{
			map[string]any{
				"id":                   "ede70e84-ff91-4f69-9f1e-558e45737998",
				"burnRateThreshold":    14.4,
				"maxBurnRateThreshold": 168.0,
				"longWindow": map[string]any{
					"value": 1.0,
					"unit":  "h",
				},
				"shortWindow": map[string]any{
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
	params := map[string]any{
		"searchType":          "searchSource",
		"size":                0.0,
		"threshold":           []any{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
		"sourceFields": []any{
			map[string]any{
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
	params := map[string]any{
		"searchType":          "searchSource",
		"threshold":           []any{1.0},
		"thresholdComparator": ">",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
	}

	errs := validateRuleParams(".es-query", params)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for missing es-query size")
	}
	if !strings.Contains(strings.Join(errs, "; "), "missing required params keys") || !strings.Contains(strings.Join(errs, "; "), "size") {
		t.Fatalf("expected missing size error, got: %v", errs)
	}
}

func TestValidateRuleParamsSloBurnRateStillRejectsUnknownExtraKeys(t *testing.T) {
	params := map[string]any{
		"sloId":        "o11y_managed_o11y-search-success-rat",
		"dependencies": []any{},
		"windows":      []any{},
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
	params := map[string]any{
		"index":               []any{"logs-*"},
		"threshold":           []any{1.0},
		"thresholdComparator": ">",
		"timeField":           "@timestamp",
		"timeWindowSize":      5.0,
		"timeWindowUnit":      "m",
		"sourceFields": []any{
			map[string]any{
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

func TestAllParamsOverridesInitialize(t *testing.T) {
	for ruleType, specs := range ruleTypeParamsOverrides {
		for _, spec := range specs {
			if spec.requiredKeys == nil {
				t.Errorf("spec %q for rule type %q has nil requiredKeys", spec.name, ruleType)
			}
		}
	}
}

func TestAllowedKeyOverridesAreNotInSchema(t *testing.T) {
	for ruleType, allowlistedKeys := range ruleTypeAdditionalAllowedParamsKeys {
		specs, ok := ruleTypeParamsOverrides[ruleType]
		if !ok || len(specs) == 0 {
			t.Fatalf("rule type %q has allowlisted keys but no params overrides", ruleType)
		}

		for _, key := range allowlistedKeys {
			if paramsSchemaAcceptsKey(specs, key) {
				t.Errorf("rule type %q allowlists key %q, but the generated params schema now accepts it; remove it from ruleTypeAdditionalAllowedParamsKeys", ruleType, key)
			}
		}
	}

	// Also verify per-spec additionalAllowedKeys are not in the spec's own schema.
	for ruleType, specs := range ruleTypeParamsOverrides {
		for _, spec := range specs {
			for _, key := range spec.additionalAllowedKeys {
				if paramsSchemaAcceptsKey([]paramsSchemaSpec{spec}, key) {
					t.Errorf("rule type %q spec %s has per-spec allowed key %q, but the generated schema now accepts it; remove it from additionalAllowedKeys", ruleType, spec.name, key)
				}
			}
		}
	}
}

func paramsSchemaAcceptsKey(specs []paramsSchemaSpec, key string) bool {
	raw, err := json.Marshal(map[string]any{key: nil})
	if err != nil {
		// This should never fail for a simple map + nil value.
		return true
	}

	for _, spec := range specs {
		target := spec.newTarget()
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(target)
		if err == nil {
			return true
		}

		// If the schema does not support this key, DisallowUnknownFields yields
		// a stable error message containing `unknown field "<key>"`.
		if strings.Contains(err.Error(), fmt.Sprintf("unknown field %q", key)) {
			continue
		}

		// Any other error implies the key was recognized (e.g. type mismatch
		// because we used `null`), so this key is part of the generated schema.
		return true
	}

	return false
}

func TestValidationCandidatePrefersDecodedOverDecodeFailure(t *testing.T) {
	candidate := validationCandidate{}
	candidate.consider(false, "params do not match expected generated schema: bad type")
	candidate.consider(true, "missing required params keys: query")

	if candidate.err == "" {
		t.Fatalf("expected a selected error to win, got empty")
	}
	if candidate.err != "missing required params keys: query" {
		t.Fatalf("expected decoded candidate to be selected, got: %v", candidate.err)
	}
}

func TestValidationCandidateKeepsStableOrderOnTie(t *testing.T) {
	candidate := validationCandidate{}
	candidate.consider(true, "missing required params keys: a")
	candidate.consider(true, "missing required params keys: b")

	if candidate.err == "" {
		t.Fatalf("expected one error, got empty")
	}
	if candidate.err != "missing required params keys: a" {
		t.Fatalf("expected first candidate to win tie, got: %v", candidate.err)
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
		params    map[string]any
		expectErr string
	}{
		{
			name:     "es-query high disk watermark valid fixture",
			ruleType: ".es-query",
			params: map[string]any{
				"aggType":                    "count",
				"groupBy":                    "top",
				"termSize":                   10.0,
				"termField":                  "docker.container.labels.co.elastic.cloud.allocator.deployment_id",
				"timeWindowSize":             30.0,
				"timeWindowUnit":             "m",
				"threshold":                  []any{30.0},
				"thresholdComparator":        ">",
				"index":                      []any{"cluster-elasticsearch-*"},
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
			params: map[string]any{
				"aggType":                    "count",
				"esQuery":                    "{\"query\":{\"bool\":{\"filter\":[]}}}",
				"excludeHitsFromPreviousRun": false,
				"groupBy":                    "top",
				"index":                      []any{"logging-*:service-constructor-*"},
				"searchType":                 "esQuery",
				"size":                       1.0,
				"sourceFields": []any{
					map[string]any{"label": "cluster_id", "searchPath": "cluster_id"},
				},
				"termField":           "cluster_id",
				"termSize":            100.0,
				"threshold":           []any{1.0},
				"thresholdComparator": ">",
				"timeField":           "@timestamp",
				"timeWindowSize":      5.0,
				"timeWindowUnit":      "m",
			},
		},
		{
			name:     "es-query failed rule evaluations valid fixture",
			ruleType: ".es-query",
			params: map[string]any{
				"aggType":                    "count",
				"groupBy":                    "top",
				"termSize":                   10.0,
				"termField":                  "rule.id",
				"size":                       100.0,
				"timeWindowSize":             6.0,
				"timeWindowUnit":             "h",
				"threshold":                  []any{3.0},
				"thresholdComparator":        ">",
				"index":                      []any{".ds-.kibana-event-log*"},
				"timeField":                  "@timestamp",
				"searchType":                 "esQuery",
				"esQuery":                    "{\"query\":{\"bool\":{\"must\":[]}}}",
				"excludeHitsFromPreviousRun": false,
			},
		},
		{
			name:     "es-query flood stage invalid fixture catches unknown key",
			ruleType: ".es-query",
			params: map[string]any{
				"aggType":                    "count",
				"groupBy":                    "top",
				"termSize":                   10.0,
				"termField":                  "docker.container.labels.co.elastic.cloud.allocator.deployment_id",
				"timeWindowSize":             10.0,
				"timeWindowUnit":             "m",
				"threshold":                  []any{10.0},
				"thresholdComparator":        ">",
				"index":                      []any{"cluster-elasticsearch-*"},
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
			name:     "es-query esql with unknown key shows variant-specific error",
			ruleType: ".es-query",
			params: map[string]any{
				"searchType":          "esqlQuery",
				"timeWindowSize":      5.0,
				"timeWindowUnit":      "m",
				"threshold":           []any{0.0},
				"thresholdComparator": ">",
				"size":                100.0,
				"esqlQuery": map[string]any{
					"esql": "FROM logs-* | STATS count = COUNT(*)",
				},
				"aggType":      "count",
				"groupBy":      "all",
				"timeField":    "@timestamp",
				"sourceFields": []any{},
				"bogusKey":     "should-fail",
			},
			expectErr: "bogusKey",
		},
		{
			name:     "slo burn rate valid fixture with dependencies",
			ruleType: "slo.rules.burnRate",
			params: map[string]any{
				"sloId": "abc123",
				"windows": []any{
					map[string]any{
						"id":                   "0c59b724-200b-462f-928c-d975e69b1eef",
						"burnRateThreshold":    3.36,
						"maxBurnRateThreshold": 168.0,
						"longWindow":           map[string]any{"value": 1.0, "unit": "h"},
						"shortWindow":          map[string]any{"value": 5.0, "unit": "m"},
						"actionGroup":          "slo.burnRate.alert",
					},
					map[string]any{
						"id":                   "62770ca9-c0f9-4a9a-bb1c-3a9666f54cf7",
						"burnRateThreshold":    1.4,
						"maxBurnRateThreshold": 28.0,
						"longWindow":           map[string]any{"value": 6.0, "unit": "h"},
						"shortWindow":          map[string]any{"value": 30.0, "unit": "m"},
						"actionGroup":          "slo.burnRate.high",
					},
				},
				"dependencies": []any{},
			},
		},
		{
			name:     "uptime monitor status valid fixture",
			ruleType: "xpack.uptime.alerts.monitorStatus",
			params: map[string]any{
				"search":                  "",
				"numTimes":                8.0,
				"timerangeUnit":           "m",
				"timerangeCount":          10.0,
				"shouldCheckStatus":       true,
				"shouldCheckAvailability": false,
				"availability": map[string]any{
					"range":     30.0,
					"rangeUnit": "d",
					"threshold": "99",
				},
				"filters": map[string]any{
					"tags": []any{"o11y"},
				},
			},
		},
		{
			name:     "custom threshold valid fixture",
			ruleType: "observability.rules.custom_threshold",
			params: map[string]any{
				"criteria": []any{
					map[string]any{
						"comparator": ">",
						"metrics":    []any{},
						"threshold":  []any{1.0},
						"timeSize":   5.0,
						"timeUnit":   "m",
					},
				},
				"searchConfiguration": map[string]any{
					"index": "logs-*",
					"query": map[string]any{
						"language": "kuery",
						"query":    "",
					},
				},
			},
		},
		{
			name:     "transform_health valid fixture",
			ruleType: "transform_health",
			params: map[string]any{
				"includeTransforms": []any{"foo-transform"},
			},
		},
		{
			name:     "apm anomaly valid fixture",
			ruleType: "apm.anomaly",
			params: map[string]any{
				"windowSize":          5.0,
				"windowUnit":          "m",
				"environment":         "production",
				"anomalySeverityType": "critical",
			},
		},
		{
			name:     "monitoring alert cluster health valid fixture",
			ruleType: "monitoring_alert_cluster_health",
			params: map[string]any{
				"duration": "1d",
			},
		},
		{
			name:     "custom threshold accepts unknown key via additionalProperties",
			ruleType: "observability.rules.custom_threshold",
			params: map[string]any{
				"criteria":      []any{},
				"alertOnNoData": true,
				"searchConfiguration": map[string]any{
					"index": "logs-*",
					"query": map[string]any{
						"language": "kuery",
						"query":    "",
					},
				},
				"bogusKey": true,
			},
		},
		{
			name:      "apm anomaly rejects unknown key",
			ruleType:  "apm.anomaly",
			expectErr: "json: unknown field \"bogusKey\"",
			params: map[string]any{
				"windowSize":          5.0,
				"windowUnit":          "m",
				"environment":         "production",
				"anomalySeverityType": "critical",
				"bogusKey":            true,
			},
		},
		{
			name:      "transform health rejects unknown key",
			ruleType:  "transform_health",
			expectErr: "json: unknown field \"bogusKey\"",
			params: map[string]any{
				"includeTransforms": []any{"foo-transform"},
				"bogusKey":          true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateRuleParams(tc.ruleType, tc.params)
			if len(errs) > 0 {
				t.Logf("validation errors: %v", errs)
			}
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
		params    map[string]any
		expectErr string
	}{
		{
			name:     "metrics alert threshold k8s node disk pressure",
			ruleType: "metrics.alert.threshold",
			params: map[string]any{
				"criteria": []any{
					map[string]any{
						"aggType":    "count",
						"comparator": ">",
						"threshold":  []any{30.0},
						"timeSize":   15.0,
						"timeUnit":   "m",
					},
				},
				"sourceId":              "default",
				"alertOnNoData":         false,
				"alertOnGroupDisappear": false,
				"filterQueryText":       "kubernetes.node.status.disk_pressure: true and orchestrator.platform.type: mki",
				"filterQuery":           "{\"bool\":{\"filter\":[]}}",
				"groupBy":               []any{"orchestrator.cluster.name", "kubernetes.node.name"},
			},
		},
		{
			name:     "logs document count project api 500",
			ruleType: "logs.alert.document.count",
			params: map[string]any{
				"timeSize": 10.0,
				"timeUnit": "m",
				"logView": map[string]any{
					"type":      "log-view-reference",
					"logViewId": "default",
				},
				"count": map[string]any{
					"value":      1.0,
					"comparator": "more than",
				},
				"criteria": []any{
					map[string]any{"field": "kubernetes.container.name", "comparator": "equals", "value": "project-api"},
					map[string]any{"field": "http.response.status_code", "comparator": "equals", "value": 500.0},
					map[string]any{"field": "message", "comparator": "matches", "value": "\"HTTP request\""},
				},
			},
		},
		{
			name:     "apm error rate cosmos throttling",
			ruleType: "apm.error_rate",
			params: map[string]any{
				"environment": "production",
				"searchConfiguration": map[string]any{
					"query": map[string]any{
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
			params: map[string]any{
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
			params: map[string]any{
				"aggType":             "count",
				"filterKuery":         "data_stream.dataset: \"proxy.log\"",
				"groupBy":             "all",
				"index":               []any{"proxy-logs-*"},
				"termSize":            5.0,
				"threshold":           []any{0.0},
				"thresholdComparator": "<=",
				"timeField":           "@timestamp",
				"timeWindowSize":      10.0,
				"timeWindowUnit":      "m",
			},
		},
		{
			name:     "uptime monitor status with stackVersion",
			ruleType: "xpack.uptime.alerts.monitorStatus",
			params: map[string]any{
				"search":                  "monitor.name: \"Production Backstage Monitor\"",
				"numTimes":                5.0,
				"timerangeUnit":           "m",
				"timerangeCount":          5.0,
				"shouldCheckStatus":       true,
				"shouldCheckAvailability": false,
				"availability": map[string]any{
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

func TestValidateRuleParamsUptimeMonitorStatusWithAvailabilityAndFilters(t *testing.T) {
	params := map[string]any{
		"numTimes":                5.0,
		"shouldCheckStatus":       true,
		"shouldCheckAvailability": false,
		"availability": map[string]any{
			"range":     30.0,
			"rangeUnit": "d",
			"threshold": "99",
		},
		"filters": map[string]any{
			"tags":              []any{"o11y"},
			"monitor.type":      []any{"http"},
			"observer.geo.name": []any{"us-east"},
			"url.port":          []any{"443"},
		},
	}

	if errs := validateRuleParams("xpack.uptime.alerts.monitorStatus", params); len(errs) > 0 {
		t.Fatalf("expected no validation errors for uptime monitor status with availability and all filter sub-fields, got: %v", errs)
	}
}

func TestValidateRuleParamsUptimeMonitorStatusRejectsSyntheticsFields(t *testing.T) {
	params := map[string]any{
		"numTimes":                5.0,
		"shouldCheckStatus":       true,
		"shouldCheckAvailability": false,
		"condition": map[string]any{
			"downMonitors": map[string]any{
				"count": 1,
			},
		},
	}

	errs := validateRuleParams("xpack.uptime.alerts.monitorStatus", params)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for synthetics-only field 'condition'")
	}
	if !strings.Contains(strings.Join(errs, "; "), "unknown field \"condition\"") {
		t.Fatalf("expected unknown field error for 'condition', got: %v", errs)
	}

	paramsWithMonitorIDs := map[string]any{
		"numTimes":                5.0,
		"shouldCheckStatus":       true,
		"shouldCheckAvailability": false,
		"monitorIds":              []any{"monitor-1"},
	}

	errs = validateRuleParams("xpack.uptime.alerts.monitorStatus", paramsWithMonitorIDs)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors for synthetics-only field 'monitorIds'")
	}
	if !strings.Contains(strings.Join(errs, "; "), "unknown field \"monitorIds\"") {
		t.Fatalf("expected unknown field error for 'monitorIds', got: %v", errs)
	}
}

func TestDiscriminatorValidationCoversAllKbapiRuleTypes(t *testing.T) {
	// Parse generated/kbapi/kibana.gen.go to extract all ValueByDiscriminator cases.
	src, err := os.ReadFile("../../../generated/kbapi/kibana.gen.go")
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	content := string(src)
	funcStart := strings.Index(content, "func (t AlertingRuleAPIBody) ValueByDiscriminator()")
	if funcStart == -1 {
		t.Fatalf("ValueByDiscriminator function not found in generated file")
	}

	switchStart := strings.Index(content[funcStart:], "switch discriminator {")
	if switchStart == -1 {
		t.Fatalf("switch statement not found in ValueByDiscriminator")
	}
	switchStart += funcStart

	switchEnd := strings.Index(content[switchStart:], "default:")
	if switchEnd == -1 {
		t.Fatalf("default case not found in switch statement")
	}
	switchEnd += switchStart

	switchBody := content[switchStart:switchEnd]
	re := regexp.MustCompile(`case "([^"]+)":`)
	matches := re.FindAllStringSubmatch(switchBody, -1)

	cases := make([]string, 0, len(matches))
	casesSet := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		cases = append(cases, m[1])
		casesSet[m[1]] = struct{}{}
	}

	if len(cases) == 0 {
		t.Fatalf("no discriminator cases found in generated file")
	}

	// Every case value must either dispatch successfully through the default
	// path (no "unknown discriminator" error) or be listed in overrides.
	for _, ruleTypeID := range cases {
		t.Run(ruleTypeID, func(t *testing.T) {
			errs := validateParamsViaDiscriminator(ruleTypeID, map[string]any{})
			if len(errs) > 0 {
				lower := strings.ToLower(errs[0])
				if strings.Contains(lower, "unknown discriminator value") {
					t.Fatalf("rule type %q not handled by discriminator validation: %v", ruleTypeID, errs)
				}
				if strings.Contains(lower, "has no params field") {
					t.Fatalf("rule type %q discriminated type lacks Params field: %v", ruleTypeID, errs)
				}
				// Other errors (missing required keys, etc.) are fine — they mean
				// the dispatcher recognized the rule type.
			}
		})
	}

	// Override table must not contain unknown discriminators.
	for ruleTypeID := range ruleTypeParamsOverrides {
		if _, ok := casesSet[ruleTypeID]; !ok {
			t.Errorf("override table contains rule type %q that is not in ValueByDiscriminator()", ruleTypeID)
		}
	}
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_notifyWhenAndFrequency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.True(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_throttleAndFrequency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		Throttle: kibanacustomtypes.NewAlertingDurationValue("5m"),
		Actions:  testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.True(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_frequencyOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringNull(),
		Throttle:   kibanacustomtypes.NewAlertingDurationNull(),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_ruleLevelOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Actions:    testActionsList(ctx, t, false),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_noFalsePositiveWhenFrequencyAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Throttle:   kibanacustomtypes.NewAlertingDurationValue("5m"),
		Actions:    testActionsList(ctx, t, false),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_prefersNotifyWhenDiagnostic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("onActiveAlert"),
		Throttle:   kibanacustomtypes.NewAlertingDurationValue("5m"),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	require.True(t, diags.HasError())
	if assert.Len(t, diags, 1) {
		assert.Contains(t, diags[0].Summary(), "notify_when")
	}
}

func TestValidateNotifyWhenThrottleFrequencyExclusivity_emptyRuleNotifyWhenIgnored(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	data := alertingRuleModel{
		NotifyWhen: types.StringValue("   "),
		Throttle:   kibanacustomtypes.NewAlertingDurationNull(),
		Actions:    testActionsList(ctx, t, true),
	}
	validateNotifyWhenThrottleFrequencyExclusivity(ctx, &data, &diags)
	assert.False(t, diags.HasError())
}

func TestConfigActionsIncludeKnownFrequencyBlock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	assert.True(t, configActionsIncludeKnownFrequencyBlock(ctx, testActionsList(ctx, t, true), &diags))
	assert.False(t, diags.HasError())

	diags = nil
	assert.False(t, configActionsIncludeKnownFrequencyBlock(ctx, testActionsList(ctx, t, false), &diags))
	assert.False(t, diags.HasError())

	diags = nil
	assert.False(t, configActionsIncludeKnownFrequencyBlock(ctx, types.ListUnknown(types.ObjectType{AttrTypes: getActionsAttrTypes()}), &diags))
}
