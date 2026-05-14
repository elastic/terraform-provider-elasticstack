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

package syntheticsmonitors_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/syntheticsmonitors"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
)

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, syntheticsmonitors.Handler{}, contracttest.Config{
		// Baseline fixture only sets config.title while Terraform nests optional filters with required inner `{label,value}` list
		// rows that are absent in the sparse JSON; the navigator cannot treat missing optional subtrees as satisfying nested TF paths.
		OmitRequiredLeafPresence: true,
		FullAPIResponse: `{
			"type": "synthetics_monitors",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 10},
			"id": "syn-mon-contract",
			"config": { "title": "Monitors" }
		}`,
		SkipFields: []string{
			"config.filters", "filters",
		},
	})
}
