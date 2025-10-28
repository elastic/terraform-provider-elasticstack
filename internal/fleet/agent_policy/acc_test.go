package agent_policy_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agent_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionAgentPolicy = version.Must(version.NewVersion("8.6.0"))

func TestAccResourceAgentPolicyFromSDK(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:                   testAccResourceAgentPolicyCreate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
				),
			},
		},
	})
}

func TestAccResourceAgentPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	policyNameGlobalDataTags := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	var originalPolicyId string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAgentPolicyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttrWith("elasticstack_fleet_agent_policy.test_policy", "policy_id", func(value string) error {
						originalPolicyId = value

						if len(value) == 0 {
							return errors.New("expected policy_id to be non empty")
						}

						return nil
					}),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttrWith("elasticstack_fleet_agent_policy.test_policy", "policy_id", func(value string) error {
						originalPolicyId = value

						if len(value) == 0 {
							return errors.New("expected policy_id to be non empty")
						}

						return nil
					}),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyUpdate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Updated Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "This policy was updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttrWith("elasticstack_fleet_agent_policy.test_policy", "policy_id", func(value string) error {
						if value != originalPolicyId {
							return fmt.Errorf("expected policy_id to not change between test steps. Was [%s], now [%s]", originalPolicyId, value)
						}

						return nil
					}),
				),
			},
			{
				SkipFunc:                versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:                  testAccResourceAgentPolicyUpdate(policyName, false),
				ResourceName:            "elasticstack_fleet_agent_policy.test_policy",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionInactivityTimeout),
				Config:   testAccResourceAgentPolicyCreateWithInactivityTimeout(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy with Inactivity Timeout"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "inactivity_timeout", "2m"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionUnenrollmentTimeout),
				Config:   testAccResourceAgentPolicyCreateWithUnenrollmentTimeout(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy with Unenrollment Timeout"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "unenrollment_timeout", "300s"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionUnenrollmentTimeout),
				Config:   testAccResourceAgentPolicyUpdateWithTimeouts(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Updated Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy with Both Timeouts"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "inactivity_timeout", "120s"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "unenrollment_timeout", "900s"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				Config:   testAccResourceAgentPolicyCreateWithGlobalDataTags(policyNameGlobalDataTags, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyNameGlobalDataTags)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "global_data_tags.tag1.string_value", "value1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "global_data_tags.tag2.number_value", "1.1"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				Config:   testAccResourceAgentPolicyUpdateWithGlobalDataTags(policyNameGlobalDataTags, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Updated Policy %s", policyNameGlobalDataTags)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "This policy was updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "global_data_tags.tag1.string_value", "value1a"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_policy.test_policy", "global_data_tags.tag2"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				Config:   testAccResourceAgentPolicyUpdateWithNoGlobalDataTags(policyNameGlobalDataTags, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Updated Policy %s", policyNameGlobalDataTags)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "This policy was updated without global data tags"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "global_data_tags.#", "0"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfNotServerless(),
				Config:   testAccResourceAgentPolicyUpdateWithSupportsAgentless(policyNameGlobalDataTags, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Updated Policy %s", policyNameGlobalDataTags)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "This policy was updated with supports_agentless"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "supports_agentless", "true"),
				),
			},
		},
	})
}

func TestAccResourceAgentPolicySkipDestroy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAgentPolicySkipDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "true"),
				),
			},
		},
	})
}

func TestAccResourceAgentPolicyWithBadGlobalDataTags(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				Config:      testAccResourceAgentPolicyCreateWithBadGlobalDataTags(policyName, true),
				ExpectError: regexp.MustCompile(".*Error: Invalid Attribute Combination.*"),
			},
		},
	})
}

func TestAccResourceAgentPolicyWithSpaceIds(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAgentPolicyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				Config:   testAccResourceAgentPolicyCreateWithSpaceIds(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy with Space IDs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.0", "default"),
				),
			},
		},
	})
}

// TestAccResourceAgentPolicySpaceReordering validates the CRITICAL bug fix:
// Reordering space_ids should NOT cause Terraform to recreate resources.
//
// This test validates the complete fix by:
// Step 1: Create policy with space_ids = ["default"]
// Step 2: Prepend a new space ["space-test-a", "default"] - proves stable operational space
// Step 3: Reorder spaces ["default", "space-test-a"] - proves position-independent lookups
//
// Without the fix: Steps 2 and 3 would cause resource recreation (policy_id changes)
// With the fix: policy_id remains constant across all steps (operational space is stable)
func TestAccResourceAgentPolicySpaceReordering(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	var originalPolicyId string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAgentPolicyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				// Step 1: Create with space_ids = ["default"]
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				Config:   testAccResourceAgentPolicySpaceReorderingStep1(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.0", "default"),
					// Capture the policy ID - it should NOT change in subsequent steps
					resource.TestCheckResourceAttrWith("elasticstack_fleet_agent_policy.test_policy", "policy_id", func(value string) error {
						originalPolicyId = value
						if len(value) == 0 {
							return errors.New("expected policy_id to be non-empty")
						}
						return nil
					}),
				),
			},
			{
				// Step 2: Prepend new space ["space-test-a", "default"]
				// CRITICAL TEST: Without fix, Terraform uses space-test-a for lookup, gets 404, recreates
				// With fix: GetOperationalSpace() returns "default" (stable), finds resource, updates in-place
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				Config:   testAccResourceAgentPolicySpaceReorderingStep2(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test space reordering - step 2: prepend new space"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.0", "space-test-a"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.1", "default"),
					// CRITICAL: policy_id must be UNCHANGED (proves stable operational space)
					resource.TestCheckResourceAttrWith("elasticstack_fleet_agent_policy.test_policy", "policy_id", func(value string) error {
						if value != originalPolicyId {
							return fmt.Errorf("policy_id changed from %s to %s - operational space not stable!", originalPolicyId, value)
						}
						return nil
					}),
				),
			},
			{
				// Step 3: Reorder spaces ["default", "space-test-a"]
				// CRITICAL TEST: Proves operational space is truly position-independent
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				Config:   testAccResourceAgentPolicySpaceReorderingStep3(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test space reordering - step 3: reorder spaces"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.0", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.1", "space-test-a"),
					// CRITICAL: policy_id must STILL be unchanged
					resource.TestCheckResourceAttrWith("elasticstack_fleet_agent_policy.test_policy", "policy_id", func(value string) error {
						if value != originalPolicyId {
							return fmt.Errorf("policy_id changed from %s to %s", originalPolicyId, value)
						}
						return nil
					}),
				),
			},
		},
	})
}

func testAccResourceAgentPolicyCreate(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "Test Agent Policy"
  monitor_logs = true
  monitor_metrics = false
  skip_destroy = %t
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Policy %s", id), skipDestroy)
}

func testAccResourceAgentPolicyCreateWithGlobalDataTags(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "Test Agent Policy"
  monitor_logs = true
  monitor_metrics = false
  skip_destroy = %t
  global_data_tags = {
		tag1 = {
			string_value = "value1"
		},
		tag2 = {
			number_value = 1.1
		}
	}
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Policy %s", id), skipDestroy)
}

func testAccResourceAgentPolicyCreateWithInactivityTimeout(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "Test Agent Policy with Inactivity Timeout"
  monitor_logs = true
  monitor_metrics = false
  skip_destroy = false
  inactivity_timeout = "2m"
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Policy %s", id))
}

func testAccResourceAgentPolicyCreateWithUnenrollmentTimeout(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "Test Agent Policy with Unenrollment Timeout"
  monitor_logs = true
  monitor_metrics = false
  skip_destroy = false
  unenrollment_timeout = "300s"
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Policy %s", id))
}

func testAccResourceAgentPolicyCreateWithBadGlobalDataTags(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "This policy was not created due to bad tags"
  monitor_logs    = false
  monitor_metrics = true
  skip_destroy    = %t
   global_data_tags = {
		tag1 = {
			string_value = "value1a"
			number_value = 1.2
		}
	}
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, fmt.Sprintf("Updated Policy %s", id), skipDestroy)
}

func testAccResourceAgentPolicyUpdateWithGlobalDataTags(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "This policy was updated"
  monitor_logs    = false
  monitor_metrics = true
  skip_destroy    = %t
   global_data_tags = {
		tag1 = {
			string_value = "value1a"
		}
	}
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, fmt.Sprintf("Updated Policy %s", id), skipDestroy)
}

func testAccResourceAgentPolicyUpdateWithNoGlobalDataTags(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "This policy was updated without global data tags"
  monitor_logs = false
  monitor_metrics = true
  skip_destroy = %t
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, fmt.Sprintf("Updated Policy %s", id), skipDestroy)
}

func testAccResourceAgentPolicyUpdateWithSupportsAgentless(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "This policy was updated with supports_agentless"
  monitor_logs = false
  monitor_metrics = true
  skip_destroy = %t
  supports_agentless = true
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, fmt.Sprintf("Updated Policy %s", id), skipDestroy)
}

func testAccResourceAgentPolicyUpdate(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "This policy was updated"
  monitor_logs = false
  monitor_metrics = true
  skip_destroy = %t
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, fmt.Sprintf("Updated Policy %s", id), skipDestroy)
}

func checkResourceAgentPolicyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_agent_policy" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
		policy, diags := fleet.GetAgentPolicy(context.Background(), fleetClient, rs.Primary.ID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if policy != nil {
			return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}

func checkResourceAgentPolicySkipDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_agent_policy" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
		policy, diags := fleet.GetAgentPolicy(context.Background(), fleetClient, rs.Primary.ID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if policy == nil {
			return fmt.Errorf("agent policy id=%v does not exist, but should still exist when skip_destroy is true", rs.Primary.ID)
		}

		if diags = fleet.DeleteAgentPolicy(context.Background(), fleetClient, rs.Primary.ID); diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
	}
	return nil
}

func testAccResourceAgentPolicyUpdateWithTimeouts(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "Test Agent Policy with Both Timeouts"
  monitor_logs = false
  monitor_metrics = true
  skip_destroy = false
  inactivity_timeout = "120s"
  unenrollment_timeout = "900s"
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Updated Policy %s", id))
}

func testAccResourceAgentPolicyCreateWithSpaceIds(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name             = "%s"
  namespace        = "default"
  description      = "Test Agent Policy with Space IDs"
  monitor_logs     = true
  monitor_metrics  = false
  skip_destroy     = false
  space_ids        = ["default"]
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Policy %s", id))
}

// Helper functions for space reordering test

func testAccResourceAgentPolicySpaceReorderingStep1(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = "space-test-a"
  name        = "Test Space A"
  description = "Test space for Fleet agent policy space reordering test"
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name             = "Policy %s"
  namespace        = "default"
  description      = "Test space reordering - step 1"
  monitor_logs     = true
  monitor_metrics  = false
  skip_destroy     = false
  space_ids        = ["default"]

  depends_on = [elasticstack_kibana_space.test_space]
}
`, id)
}

func testAccResourceAgentPolicySpaceReorderingStep2(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = "space-test-a"
  name        = "Test Space A"
  description = "Test space for Fleet agent policy space reordering test"
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name             = "Policy %s"
  namespace        = "default"
  description      = "Test space reordering - step 2: prepend new space"
  monitor_logs     = true
  monitor_metrics  = false
  skip_destroy     = false
  # CRITICAL TEST: Prepending "space-test-a" before "default"
  # Without the fix: Terraform queries using space-test-a, gets 404, recreates resource
  # With the fix: Terraform uses "default" (position-independent), finds resource, updates in-place
  space_ids        = ["space-test-a", "default"]

  depends_on = [elasticstack_kibana_space.test_space]
}
`, id)
}

func testAccResourceAgentPolicySpaceReorderingStep3(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = "space-test-a"
  name        = "Test Space A"
  description = "Test space for Fleet agent policy space reordering test"
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name             = "Policy %s"
  namespace        = "default"
  description      = "Test space reordering - step 3: reorder spaces"
  monitor_logs     = true
  monitor_metrics  = false
  skip_destroy     = false
  # CRITICAL TEST: Reordering spaces (default now first)
  # With the fix: Still uses "default", resource found, updates in-place
  space_ids        = ["default", "space-test-a"]

  depends_on = [elasticstack_kibana_space.test_space]
}
`, id)
}
