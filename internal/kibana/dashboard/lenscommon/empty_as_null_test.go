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

package lenscommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// emptyAsNullUnsupportedFieldOps are field-metric operations whose Kibana metric schema
// does not define empty_as_null (shared between the predicate and populator tests).
var emptyAsNullUnsupportedFieldOps = []string{
	"percentile", "percentile_rank",
	dashboardValueAvg, "median", "standard_deviation",
	"last_value",
}

func TestOperationSupportsEmptyAsNull(t *testing.T) {
	t.Parallel()

	for _, op := range []string{operationCount, operationSum, operationUniqueCount} {
		assert.Truef(t, operationSupportsEmptyAsNull(op), "expected %q to support empty_as_null", op)
	}

	for _, op := range emptyAsNullUnsupportedFieldOps {
		assert.Falsef(t, operationSupportsEmptyAsNull(op), "expected %q to NOT support empty_as_null", op)
	}

	// Pipeline and breakdown operations also lack empty_as_null support.
	for _, op := range []string{"moving_average", "cumulative_sum", "differences", "counter_rate", OperationTerms, "", "unknown"} {
		assert.Falsef(t, operationSupportsEmptyAsNull(op), "expected %q to NOT support empty_as_null", op)
	}
}

// emptyAsNullPopulators enumerates the populate functions that inject the
// empty_as_null default so the gating is asserted uniformly across chart families.
var emptyAsNullPopulators = map[string]func(map[string]any) map[string]any{
	"PopulateLensMetricDefaults":         PopulateLensMetricDefaults,
	"PopulateMetricChartMetricDefaults":  PopulateMetricChartMetricDefaults,
	"PopulateGaugeMetricDefaults":        PopulateGaugeMetricDefaults,
	"PopulatePieChartMetricDefaults":     PopulatePieChartMetricDefaults,
	"PopulateLegacyMetricMetricDefaults": PopulateLegacyMetricMetricDefaults,
	"PopulateTagcloudMetricDefaults":     PopulateTagcloudMetricDefaults,
	"PopulateRegionMapMetricDefaults":    PopulateRegionMapMetricDefaults,
}

func TestPopulators_injectEmptyAsNullForSupportedOperations(t *testing.T) {
	t.Parallel()

	for _, op := range []string{operationCount, operationSum, operationUniqueCount} {
		for name, populate := range emptyAsNullPopulators {
			model := populate(map[string]any{"operation": op, "field": "f"})
			v, exists := model["empty_as_null"]
			assert.Truef(t, exists, "%s: expected empty_as_null to be injected for operation %q", name, op)
			assert.Equalf(t, false, v, "%s: expected injected empty_as_null=false for operation %q", name, op)
		}
	}
}

func TestPopulators_doNotInjectEmptyAsNullForUnsupportedOperations(t *testing.T) {
	t.Parallel()

	for _, op := range emptyAsNullUnsupportedFieldOps {
		for name, populate := range emptyAsNullPopulators {
			model := populate(map[string]any{"operation": op, "field": "f"})
			_, exists := model["empty_as_null"]
			assert.Falsef(t, exists, "%s: expected empty_as_null NOT to be injected for operation %q", name, op)
		}
	}
}

func TestPopulateLensMetricDefaults_preservesExplicitEmptyAsNull(t *testing.T) {
	t.Parallel()

	// An explicit empty_as_null is never overwritten, regardless of operation support.
	supported := PopulateLensMetricDefaults(map[string]any{"operation": operationCount, "empty_as_null": true})
	assert.Equal(t, true, supported["empty_as_null"])

	// For an unsupported operation the gate never adds the key, and an explicit value is
	// left untouched (the practitioner owns it, even though Kibana would reject it).
	explicit := PopulateLensMetricDefaults(map[string]any{"operation": emptyAsNullUnsupportedFieldOps[0], "empty_as_null": true})
	assert.Equal(t, true, explicit["empty_as_null"])
}
