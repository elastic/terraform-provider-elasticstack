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

package privatelocation_test

// this test is in synthetics_test package, because of https://github.com/elastic/kibana/issues/190801
// having both tests in same package allows to use mutex in kibana API client and workaround the issue

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	providerConfig = `
provider "elasticstack" {
  	elasticsearch {}
	kibana {}
	fleet{}
}
`
)

var (
	minKibanaPrivateLocationAPIVersion = version.Must(version.NewVersion("8.12.0"))
)

func TestSyntheticPrivateLocationResource(t *testing.T) {
	resourceID := "elasticstack_kibana_synthetics_private_location.test"
	randomSuffix := sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				Config: testConfig("testacc", "test_policy", randomSuffix) + fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "pl-test-label-%s"
	agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
	tags = ["a", "b"]
	geo = {
		lat = 42.42
		lon = -42.42
	}
}
`, randomSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "b"),
					resource.TestCheckResourceAttr(resourceID, "geo.lat", "42.42"),
					resource.TestCheckResourceAttr(resourceID, "geo.lon", "-42.42"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
				Config: testConfig("testacc", "test_policy", randomSuffix) + fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "pl-test-label-%s"
	agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
	tags = ["a", "b"]
	geo = {
		lat = 42.42
		lon = -42.42
	}
}
`, randomSuffix),
			},
			// Update and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				Config: testConfig("default", "test_policy_default", randomSuffix) + fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "pl-test-label-2-%s"
	agent_policy_id = elasticstack_fleet_agent_policy.test_policy_default.policy_id
	tags = ["c", "d", "e"]
	geo = {
		lat = -33.21
		lon = -33.21
	}
}
`, randomSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceID, "tags.2", "e"),
					resource.TestCheckResourceAttr(resourceID, "geo.lat", "-33.21"),
					resource.TestCheckResourceAttr(resourceID, "geo.lon", "-33.21"),
				),
			},
			// Update and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				Config: testConfig("default", "test_policy_default", randomSuffix) + fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "pl-test-label-2-%s"
	agent_policy_id = elasticstack_fleet_agent_policy.test_policy_default.policy_id
}
`, randomSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckNoResourceAttr(resourceID, "tags"),
					resource.TestCheckNoResourceAttr(resourceID, "geo"),
				),
			},
			// Update and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				Config: testConfig("default", "test_policy_default", randomSuffix) + fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "pl-test-label-2-%s"
	agent_policy_id = elasticstack_fleet_agent_policy.test_policy_default.policy_id
	tags = ["c", "d", "e"]
}
`, randomSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceID, "tags.2", "e"),
					resource.TestCheckNoResourceAttr(resourceID, "geo"),
				),
			},
			// Update and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				Config: testConfig("default", "test_policy_default", randomSuffix) + fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "pl-test-label-2-%s"
	agent_policy_id = elasticstack_fleet_agent_policy.test_policy_default.policy_id
	geo = {
		lat = -33.21
		lon = -33.21
	}
}
`, randomSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckNoResourceAttr(resourceID, "tags"),
					resource.TestCheckResourceAttr(resourceID, "geo.lat", "-33.21"),
					resource.TestCheckResourceAttr(resourceID, "geo.lon", "-33.21"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testConfig(namespace, agentPolicy, randomSuffix string) string {
	return providerConfig + fmt.Sprintf(`
resource "elasticstack_fleet_agent_policy" "%s" {
	name            = "Private Location Agent Policy - %s - %s"
	namespace       = "%s"
	description     = "TestPrivateLocationResource Agent Policy"
	monitor_logs    = true
	monitor_metrics = true
	skip_destroy    = false
}
`, agentPolicy, agentPolicy, randomSuffix, namespace)
}
