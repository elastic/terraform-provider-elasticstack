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

import "testing"

func TestClassifyLensDashboardAppConfigFromRoot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		root map[string]any
		want lensConfigClass
	}{
		{
			name: "by_value chart with ref_id and time_range still chart",
			root: map[string]any{
				"type":   "xyChartESQL",
				"ref_id": "panel_0",
				"time_range": map[string]any{
					"from": "now-7d", "to": "now",
				},
			},
			want: lensConfigClassByValueChart,
		},
		{
			name: "by_reference without top-level chart type",
			root: map[string]any{
				"ref_id": "panel_0",
				"time_range": map[string]any{
					"from": "now-7d", "to": "now",
				},
			},
			want: lensConfigClassByReference,
		},
		{
			name: "ambiguous incomplete",
			root: map[string]any{"ref_id": "x"},
			want: lensConfigClassAmbiguous,
		},
		{
			name: "by_value empty type string not chart",
			root: map[string]any{"type": ""},
			want: lensConfigClassAmbiguous,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := classifyLensDashboardAppConfigFromRoot(tc.root)
			if got != tc.want {
				t.Fatalf("classify: got %v want %v", got, tc.want)
			}
		})
	}
}

func TestLensDashboardAppByValueToAPI_UnknownConfigJSON(t *testing.T) {
	t.Parallel()
	_, diags := lensDashboardAppByValueToAPI(
		lensDashboardAppByValueModel{},
		lensDashboardAPIGrid{},
		nil,
	)
	if !diags.HasError() {
		t.Fatal("expected error for unknown by_value.config_json")
	}
}
