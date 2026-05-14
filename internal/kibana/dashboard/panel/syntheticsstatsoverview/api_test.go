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

package syntheticsstatsoverview_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/syntheticsstatsoverview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
)

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, syntheticsstatsoverview.Handler{}, contracttest.Config{
		// Baseline fixture exposes only sparse config fields (here `title`). Terraform nests optional drilldowns and structured
		// filters requiring inner list elements (`label`, `value`) when those branches are exercised; fixture.config does not
		// share that nested segment layout, so the harness cannot prove required Terraform leaves exist literal-for-literal under
		// the raw fixture path.
		OmitRequiredLeafPresence: true,
		FullAPIResponse: `{
			"type": "synthetics_stats_overview",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 8},
			"id": "syn-stats-contract",
			"config": {
				"title": "Stats overview"
			}
		}`,
		SkipFields: []string{
			"config.drilldowns", "drilldowns",
			"config.filters", "filters",
		},
	})
}
