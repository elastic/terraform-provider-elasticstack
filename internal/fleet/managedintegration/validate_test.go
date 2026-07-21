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

package managedintegration_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitResourceManagedIntegrationGlobalDataTagsNeitherValueSet(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = "unit-test-managed-integration"
  policy_template = "cspm"
  package = {
    name    = "cloud_security_posture"
    version = "3.4.0"
  }
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

func TestUnitResourceManagedIntegrationGlobalDataTagsBothValuesSet(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = "unit-test-managed-integration"
  policy_template = "cspm"
  package = {
    name    = "cloud_security_posture"
    version = "3.4.0"
  }
  global_data_tags = {
    tag1 = {
      string_value = "value1a"
      number_value = 1.2
    }
  }
}
`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Combination.*`),
			},
		},
	})
}
