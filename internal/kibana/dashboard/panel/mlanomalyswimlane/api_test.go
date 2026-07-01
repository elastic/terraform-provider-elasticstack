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

package mlanomalyswimlane_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlanomalyswimlane"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
)

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, mlanomalyswimlane.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "ml_anomaly_swimlane",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 8},
			"id": "ml-swim-contract",
			"config": {
				"swimlane_type": "overall",
				"job_ids": ["job-a"],
				"per_page": 10,
				"title": "Swim Lane",
				"description": "Anomaly swim lane panel",
				"hide_title": true,
				"hide_border": false,
				"time_range": {
					"from": "now-7d",
					"to": "now",
					"mode": "relative"
				}
			}
		}`,
		// job_ids is a required list attribute; contracttest cannot synthesize list zero values yet.
		OmitValidateRequiredZero: true,
	})
}
