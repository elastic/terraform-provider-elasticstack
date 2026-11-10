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

package rule_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Every case plans the module in testdata/<test>/<case> without a live stack
// and expects the validation diagnostic before any API call. Terraform wraps
// diagnostic text, so a space in a pattern matches any whitespace run.
func TestUnitResourceAlertingV2Rule_validation(t *testing.T) {
	cases := map[string]string{
		"grouping_fields_too_many":              `grouping\.fields.*at most 16`,
		"grouping_fields_empty":                 `fields = \[\].*at least one item.*omit the attribute`,
		"metadata_name_too_long":                `metadata\.name.*between 1 and 256`,
		"pending_count_too_large":               `state_transition\.pending\.count.*between 0 and 1000`,
		"tags_empty":                            `tags = \[\].*at least one item.*omit the attribute`,
		"artifacts_empty":                       `artifacts = \[\].*at least one item.*omit the attribute`,
		"rule_id_invalid":                       `rule_id.*1-150 characters of letters, digits, underscores, or hyphens`,
		"kind_unknown":                          `kind.*\["alert" "signal"\]`,
		"schedule_every_malformed":              `schedule\.every.*unsigned integer followed by one of ms, s, m, h, d, or w`,
		"schedule_every_years":                  `schedule\.every.*unsigned integer followed by one of ms, s, m, h, d, or w`,
		"schedule_every_below_minimum":          `schedule\.every.*at least 5s`,
		"no_data_strategy_alert":                `no_data\.strategy.*\["ignore" "keep_last" "resolve"\]`,
		"artifact_ids_duplicate":                `Duplicate artifact id.*"rb" is used more than once`,
		"description_too_long":                  `metadata\.description.*between 1 and 1024`,
		"tags_too_many":                         `metadata\.tags.*at most 20`,
		"tag_too_long":                          `metadata\.tags\[0\].*between 1 and 128`,
		"routing_tags_empty":                    `routing_tags = \[\].*at least one item.*omit the attribute`,
		"routing_tags_too_many":                 `metadata\.routing_tags.*at most 20`,
		"routing_tag_too_long":                  `metadata\.routing_tags\[0\].*between 1 and 128`,
		"query_base_too_long":                   `query\.base.*between 1 and 10000`,
		"breach_segment_too_long":               `query\.breach\.segment.*between 1 and 10000`,
		"recovery_segment_too_long":             `recovery\.segment.*between 1 and 10000`,
		"recovery_query_too_long":               `recovery\.query.*between 1 and 10000`,
		"no_data_query_too_long":                `no_data\.query.*between 1 and 10000`,
		"pending_operator_invalid":              `state_transition\.pending\.operator.*\["and" "or"\]`,
		"grouping_field_too_long":               `grouping\.fields\[0\].*between 1 and 256`,
		"artifacts_too_many":                    `artifacts.*at most 100`,
		"artifact_id_too_long":                  `artifacts\[0\]\.id.*between 1 and 150`,
		"artifact_type_too_long":                `artifacts\[0\]\.type.*between 1 and 128`,
		"time_field_too_long":                   `time_field.*between 1 and 256`,
		"schedule_lookback_too_long":            `schedule\.lookback.*at most 365d`,
		"pending_timeframe_too_long":            `state_transition\.pending\.timeframe.*at most 365d`,
		"artifact_data_array":                   `expected value to be a JSON object.*data = jsonencode\(\["a"\]\)`,
		"artifact_data_not_json":                `data = "not json".*not valid JSON`,
		"signal_with_recovery":                  "signal rules cannot set `recovery` or `no_data`",
		"alert_without_no_data":                 "`no_data` is required for alert rules",
		"condition_recovery_without_breach":     "`recovery.strategy = \"condition\"` requires `query.breach`",
		"resolve_without_breach_or_query":       "a no-data strategy other than `ignore` requires `query.breach` or `no_data.query`",
		"manual_recovery_with_recovering_phase": "`state_transition.recovering` is not allowed with `recovery.strategy = \"manual\"`",
		"operator_without_both_thresholds":      "`operator` requires both `count` and `timeframe`",
		"signal_with_routing_tags":              "`metadata.routing_tags` is only allowed when `kind` is \"alert\"",
		"signal_with_state_transition":          "`state_transition` is only allowed when `kind` is \"alert\"",
		"recovery_segment_without_condition":    "`segment` is only allowed with `strategy = \"condition\"`",
		"recovery_query_strategy_without_query": "`query` is required for `strategy = \"query\"`",
		"no_data_ignore_with_query":             "`query` is not allowed with `strategy = \"ignore\"`",
	}

	for name, expectError := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resource.UnitTest(t, resource.TestCase{
				Steps: []resource.TestStep{
					{
						ProtoV6ProviderFactories: acctest.Providers,
						ConfigDirectory:          config.TestNameDirectory(),
						ExpectError:              wrappedDiagnostic(expectError),
					},
				},
			})
		})
	}
}

func wrappedDiagnostic(pattern string) *regexp.Regexp {
	return regexp.MustCompile(`(?s)` + strings.ReplaceAll(pattern, " ", `\s+`))
}
