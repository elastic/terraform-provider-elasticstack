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

package agentpolicy_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitResourceAgentPolicyGlobalDataTagsNeitherValueSet(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "unit-test-policy"
  namespace       = "default"
  description     = "Test Agent Policy"
  monitor_logs    = true
  monitor_metrics = false
  global_data_tags = {
    my_tag = {}
  }
}
`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Combination.*`),
			},
		},
	})
}

func TestUnitResourceAgentPolicyPolicyIDValidation(t *testing.T) {
	tests := []struct {
		name        string
		policyID    string
		expectError *regexp.Regexp
	}{
		{
			name:        "empty policy_id",
			policyID:    "",
			expectError: regexp.MustCompile(`policy_id must be between 1 and 255 characters`),
		},
		{
			name:        "path separator",
			policyID:    "bad/id",
			expectError: regexp.MustCompile(`policy_id must not contain path separators`),
		},
		{
			name:        "traversal sequence",
			policyID:    "my..policy",
			expectError: regexp.MustCompile(`policy_id must not contain traversal sequences`),
		},
		{
			name:        "too long",
			policyID:    strings.Repeat("a", 256),
			expectError: regexp.MustCompile(`policy_id must be between 1 and 255 characters`),
		},
		{
			name:        "reserved substring",
			policyID:    "my-__proto__-policy",
			expectError: regexp.MustCompile(`policy_id must not contain reserved keys \("__proto__"\)`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.Providers,
				Steps: []resource.TestStep{
					{
						Config:      testUnitResourceAgentPolicyPolicyIDConfig(tc.policyID),
						ExpectError: tc.expectError,
					},
				},
			})
		})
	}
}

func testUnitResourceAgentPolicyPolicyIDConfig(policyID string) string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "unit-test-policy"
  namespace       = "default"
  description     = "Test Agent Policy"
  monitor_logs    = true
  monitor_metrics = false
  policy_id       = "` + policyID + `"
}
`
}
