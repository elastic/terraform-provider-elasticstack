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

package dashboard

import (
	"testing"
)

// Regression: strings shaped like plan vs. Kibana GET for by-value lens-dashboard-app
// (see acc lens by-value; keeps embed logic honest when stack output changes).
func TestJsonValuePriorEmbedded_terraformAndKibanaSample(t *testing.T) {
	t.Parallel()
	prior := `{"data_source":{"index_pattern":"metrics-*","time_field":"@timestamp","type":"data_view_spec"},` +
		`"filters":[],"metrics":[{"format":{"type":"number"},"operation":"count","type":"primary"}],` +
		`"query":{"expression":"","language":"kql"},"styling":{"icon":{"name":"heart"}},` +
		`"time_range":{"from":"now-15m","to":"now"},"title":"Acc by-value","type":"metric"}`
	api := `{"time_range":{"from":"now-15m","to":"now"},"title":"Acc by-value",` +
		`"data_source":{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"},` +
		`"type":"metric","sampling":1,"ignore_global_filters":false,` +
		`"metrics":[{"type":"primary","operation":"count","empty_as_null":false,` +
		`"format":{"type":"number","decimals":2,"compact":false},"color":{"type":"auto"}}],` +
		`"styling":{"primary":{"position":"bottom","labels":{"alignment":"left"},` +
		`"value":{"sizing":"auto","alignment":"right"}}}}`
	ok, err := jsonValuePriorEmbeddedInExpandedCurrent(prior, api)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatal("expected prior to be a value-subset of api for preservation")
	}
}
