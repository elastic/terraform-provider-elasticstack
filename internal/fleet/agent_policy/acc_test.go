package agent_policy_test

import (
	"context"
	_ "embed"
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
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionAgentPolicy = version.Must(version.NewVersion("8.6.0"))

//go:embed testdata/TestAccResourceAgentPolicyFromSDK/main.tf
var sdkCreateTestConfig string

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
				Config:   sdkCreateTestConfig,
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(false),
				},
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(false),
				},
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(false),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(false),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Updated Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(false),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Updated Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(false),
				},
				ResourceName:            "elasticstack_fleet_agent_policy.test_policy",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionInactivityTimeout),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_inactivity_timeout"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionUnenrollmentTimeout),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_unenrollment_timeout"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionUnenrollmentTimeout),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_with_timeouts"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Updated Policy %s", policyName)),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_with_global_data_tags"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Policy %s", policyNameGlobalDataTags)),
					"skip_destroy": config.BoolVariable(false),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_with_global_data_tags"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Updated Policy %s", policyNameGlobalDataTags)),
					"skip_destroy": config.BoolVariable(false),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_with_no_global_data_tags"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Updated Policy %s", policyNameGlobalDataTags)),
					"skip_destroy": config.BoolVariable(false),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfNotServerless(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_with_supports_agentless"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Updated Policy %s", policyNameGlobalDataTags)),
					"skip_destroy": config.BoolVariable(false),
				},
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicySkipDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(true),
				},
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
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionGlobalDataTags),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_with_bad_tags"),
				ConfigVariables: config.Variables{
					"policy_name":  config.StringVariable(fmt.Sprintf("Updated Policy %s", policyName)),
					"skip_destroy": config.BoolVariable(true),
				},
				ExpectError: regexp.MustCompile(".*Error: Invalid Attribute Combination.*"),
			},
		},
	})
}

func TestAccResourceAgentPolicyWithSpaceIds(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_with_space_ids"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy with Space IDs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.*", "default"),
				),
			},
		},
	})
}

// TestAccResourceAgentPolicySpaceReordering validates that space_ids as a Set works correctly:
// With Sets, order doesn't matter - Terraform handles set comparison automatically.
//
// This test validates:
// Step 1: Create policy with space_ids = ["default"]
// Step 2: Add a new space ["space-test-a", "default"] - proves stable operational space
// Step 3: Same spaces in different order ["default", "space-test-a"] - no drift (Sets are unordered)
//
// With Sets: No drift from reordering, policy_id remains constant across all steps
func TestAccResourceAgentPolicySpaceReordering(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	var originalPolicyId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				// Step 1: Create with space_ids = ["default"]
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				ConfigDirectory: acctest.NamedTestCaseDirectory("step1"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.*", "default"),
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
				ProtoV6ProviderFactories: acctest.Providers,
				// Step 2: Add new space ["space-test-a", "default"]
				// With Sets + GetOperationalSpaceFromState: reads from STATE, finds resource, updates in-place
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				ConfigDirectory: acctest.NamedTestCaseDirectory("step2"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test space reordering - step 2: prepend new space"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.*", "space-test-a"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.*", "default"),
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
				ProtoV6ProviderFactories: acctest.Providers,
				// Step 3: Same spaces, different order ["default", "space-test-a"]
				// With Sets: No drift because order doesn't matter - Terraform sees identical sets
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionSpaceIds),
				ConfigDirectory: acctest.NamedTestCaseDirectory("step3"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test space reordering - step 3: reorder spaces"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.*", "default"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_policy.test_policy", "space_ids.*", "space-test-a"),
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
		policy, diags := fleet.GetAgentPolicy(context.Background(), fleetClient, rs.Primary.ID, "")
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
		policy, diags := fleet.GetAgentPolicy(context.Background(), fleetClient, rs.Primary.ID, "")
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if policy == nil {
			return fmt.Errorf("agent policy id=%v does not exist, but should still exist when skip_destroy is true", rs.Primary.ID)
		}

		if diags = fleet.DeleteAgentPolicy(context.Background(), fleetClient, rs.Primary.ID, ""); diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
	}
	return nil
}

func TestAccResourceAgentPolicyWithHostNameFormat(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: Create with host_name_format = "fqdn"
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionAgentFeatures),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_with_fqdn"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy with FQDN host name format"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "host_name_format", "fqdn"),
				),
			},
			{
				// Step 2: Remove host_name_format from config - should use default "hostname"
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionAgentFeatures),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_host_name_format"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy without host_name_format"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "host_name_format", "hostname"),
				),
			},
			{
				// Step 3: Explicitly set host_name_format = "hostname"
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionAgentFeatures),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_to_hostname"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy with hostname format"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "host_name_format", "hostname"),
				),
			},
		},
	})
}

func TestAccResourceAgentPolicyWithRequiredVersions(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionRequiredVersions),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.%", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.8.15.0", "100"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionRequiredVersions),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_percentage"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.%", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.8.15.0", "50"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionRequiredVersions),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("add_version"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.8.15.0", "50"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.8.16.0", "50"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionRequiredVersions),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unset_versions"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.8.15.0", "50"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.8.16.0", "50"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(agent_policy.MinVersionRequiredVersions),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_versions"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "required_versions.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceAgentPolicyWithAdvancedSettings(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with logging settings
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_with_logging"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.logging_level", "debug"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.logging_to_files", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.go_max_procs", "2"),
				),
			},
			// Step 2: Update settings
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_settings"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.logging_level", "info"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.logging_to_files", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.logging_files_keepfiles", "7"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.logging_files_rotateeverybytes", "10485760"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.go_max_procs", "4"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "advanced_settings.download_target_directory", "/tmp/elastic-agent"),
				),
			},
			// Step 3: Remove settings
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_settings"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(fmt.Sprintf("Policy %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
				),
			},
		},
	})
}
