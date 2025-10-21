package private_location_test

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
	resourceId := "elasticstack_kibana_synthetics_private_location.test"
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
					resource.TestCheckResourceAttr(resourceId, "label", fmt.Sprintf("pl-test-label-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceId, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceId, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceId, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceId, "tags.1", "b"),
					resource.TestCheckResourceAttr(resourceId, "geo.lat", "42.42"),
					resource.TestCheckResourceAttr(resourceId, "geo.lon", "-42.42"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ResourceName:      resourceId,
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
					resource.TestCheckResourceAttr(resourceId, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceId, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceId, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceId, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceId, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceId, "tags.2", "e"),
					resource.TestCheckResourceAttr(resourceId, "geo.lat", "-33.21"),
					resource.TestCheckResourceAttr(resourceId, "geo.lon", "-33.21"),
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
					resource.TestCheckResourceAttr(resourceId, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceId, "agent_policy_id"),
					resource.TestCheckNoResourceAttr(resourceId, "tags"),
					resource.TestCheckNoResourceAttr(resourceId, "geo"),
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
					resource.TestCheckResourceAttr(resourceId, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceId, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceId, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceId, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceId, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceId, "tags.2", "e"),
					resource.TestCheckNoResourceAttr(resourceId, "geo"),
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
					resource.TestCheckResourceAttr(resourceId, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceId, "agent_policy_id"),
					resource.TestCheckNoResourceAttr(resourceId, "tags"),
					resource.TestCheckResourceAttr(resourceId, "geo.lat", "-33.21"),
					resource.TestCheckResourceAttr(resourceId, "geo.lon", "-33.21"),
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
