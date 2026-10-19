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

package agentpolicyds_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentpolicy"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func randomPolicyName() string {
	return fmt.Sprintf("Policy %s", sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum))
}

var minVersionAgentPolicy = version.Must(version.NewVersion("8.6.0"))
var minVersionTamperProtection = version.Must(version.NewVersion("8.14.0"))

var (
	minVersionAgentPolicySpaceID  = agentpolicy.MinVersionSpaceIDs
	minVersionRequiredVersions    = agentpolicy.MinVersionRequiredVersions
	minVersionGlobalDataTags      = agentpolicy.MinVersionGlobalDataTags
	minVersionAdvancedSettings    = agentpolicy.MinVersionAdvancedSettings
	minVersionAgentInternal       = agentpolicy.MinVersionAgentInternal
	minVersionAdvancedMonitoring  = agentpolicy.MinVersionAdvancedMonitoring
	minVersionSupportsAgentless   = agentpolicy.MinSupportsAgentlessVersion
	minVersionInactivityTimeout   = agentpolicy.MinVersionInactivityTimeout
	minVersionHostNameFormat      = agentpolicy.MinVersionAgentFeatures
	minVersionUnenrollmentTimeout = agentpolicy.MinVersionUnenrollmentTimeout
	minVersionFeaturesDisableAcks = agentpolicy.MinVersionFeaturesDisablePolicyChangeAcksEnabled
	minVersionFeaturesIncludeTags = agentpolicy.MinVersionFeaturesIncludeTagsInEventsEnabled
)

// skipUnlessMonitoringRuntimeExperimentalSupported skips the step unless the live
// Elasticsearch server satisfies agentpolicy.MonitoringRuntimeExperimentalSupported
// (8.19.x or 9.1.0+), a non-linear version check that versionutils.SkipIfUnsupported
// cannot express.
func skipUnlessMonitoringRuntimeExperimentalSupported() (bool, error) {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return false, err
	}
	serverVersion, _, diags := clients.AcceptanceServerInfo(context.Background(), client)
	if diags.HasError() {
		return false, fmt.Errorf("failed to determine the elasticsearch server version: %v", diags)
	}
	return !agentpolicy.MonitoringRuntimeExperimentalSupported(serverVersion), nil
}

const agentInternalYAML = "agent:\n  monitoring:\n    http:\n      enabled: true\n    # trailing comment\n  logging:\n    level: info\n"

// Terraform wraps the ExactlyOneOf detail between "is" and "required", and the
// two attribute diagnostics are emitted in either order.
func exactlyOneOfPattern(detailPrefix string) *regexp.Regexp {
	policyIDThenName := regexp.QuoteMeta(detailPrefix+"[policy_id.<.name] is") + `\s+required`
	nameThenPolicyID := regexp.QuoteMeta(detailPrefix+"[name.<.policy_id] is") + `\s+required`
	return regexp.MustCompile(`(?s)(?:` +
		`Invalid Attribute Combination.*` + policyIDThenName + `.*Invalid Attribute Combination.*` + nameThenPolicyID +
		`|` +
		`Invalid Attribute Combination.*` + nameThenPolicyID + `.*Invalid Attribute Combination.*` + policyIDThenName +
		`)`)
}

var exactlyOneOfNeitherSelector = exactlyOneOfPattern("No attribute specified when one (and only one) of ")
var exactlyOneOfBothSelectors = exactlyOneOfPattern("2 attributes specified when one (and only one) of ")

func TestAccDataSourceAgentPolicyByID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.elasticstack_fleet_agent_policy.test", "space_id"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyByName(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.test", "policy_id", "elasticstack_fleet_agent_policy.test_policy"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.test", "id", "elasticstack_fleet_agent_policy.test_policy"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyOmittedOptionals(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyNeitherSelector(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan"),
				PlanOnly:                 true,
				ExpectError:              exactlyOneOfNeitherSelector,
			},
		},
	})
}

func TestAccDataSourceAgentPolicyBothSelectors(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan"),
				PlanOnly:                 true,
				ExpectError:              exactlyOneOfBothSelectors,
			},
		},
	})
}

func TestAccDataSourceAgentPolicyNameNotFound(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	policyName := "does-not-exist-" + sdkacctest.RandStringFromCharSet(16, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ExpectError: regexp.MustCompile(`(?s)` +
					regexp.QuoteMeta(fmt.Sprintf("no agent policy found matching name %q in", policyName)) +
					`\s+` +
					regexp.QuoteMeta(`space "default"`)),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyByIDNotFound(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ExpectError:              regexp.MustCompile(`(?s)fleet_agent_policy "does-not-exist" was\s+not found`),
			},
		},
	})
}

func TestAccDataSourceAgentPolicySpaceID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicySpaceID, versionutils.FlavorAny)

	policyName := randomPolicyName()
	spaceID := "test-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceName := "Test Space " + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("spaces"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"space_id":    config.StringVariable(spaceID),
					"space_name":  config.StringVariable(spaceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.elasticstack_fleet_agent_policy.omitted", "space_id"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.omitted", "policy_id", "elasticstack_fleet_agent_policy.default_space"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.omitted", "id", "elasticstack_fleet_agent_policy.default_space"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.omitted", "description", "Default space agent policy"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.omitted"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.explicit_default", "space_id", "default"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.explicit_default", "policy_id", "elasticstack_fleet_agent_policy.default_space"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.explicit_default", "id", "elasticstack_fleet_agent_policy.default_space"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.explicit_default", "description", "Default space agent policy"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.explicit_default"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.other", "space_id", spaceID),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.other", "policy_id", "elasticstack_fleet_agent_policy.other_space"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.other", "id", "elasticstack_fleet_agent_policy.other_space"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.other", "description", "Non-default space agent policy"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.other"),
					checkResourceIDsDiffer("elasticstack_fleet_agent_policy.default_space", "elasticstack_fleet_agent_policy.other_space"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wrong_space"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"space_id":    config.StringVariable(spaceID),
					"space_name":  config.StringVariable(spaceName),
				},
				ExpectError: regexp.MustCompile(`(?s)fleet_agent_policy "[^"]+" in\s+space "default"\s+was\s+not found`),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyMonitoringUpdate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("enabled"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "monitor_metrics", "true"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("disabled"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "monitor_logs", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "monitor_metrics", "false"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicySelectorSwitch(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicy, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_id"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.test", "policy_id", "elasticstack_fleet_agent_policy.test_policy"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.test", "id", "elasticstack_fleet_agent_policy.test_policy"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "name", policyName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_name"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.test", "policy_id", "elasticstack_fleet_agent_policy.test_policy"),
					checkAttrEqualsResourceID("data.elasticstack_fleet_agent_policy.test", "id", "elasticstack_fleet_agent_policy.test_policy"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "name", policyName),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyGlobalDataTags(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionGlobalDataTags, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.%", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.tag1.string_value", "value1"),
					checkFloat32Attr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.tag2.number_value", 1.1),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// A third entry proves multi-element handling beyond the original two tags.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.%", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.tag1.string_value", "value1"),
					checkFloat32Attr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.tag2.number_value", 1.1),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.tag3.string_value", "value3"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// Removing all tags should surface as an empty collection, not a stale value.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "global_data_tags.%", "0"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyAdvancedSettings(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAdvancedSettings, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.logging_level", "debug"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.logging_to_files", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.logging_files_interval", "45s"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.logging_files_keepfiles", "10"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.logging_files_rotateeverybytes", "20971520"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.logging_metrics_period", "1m"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.go_max_procs", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.download_timeout", "3h"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.download_target_directory", "/tmp/agent-downloads"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyAdvancedSettingsMonitoringRuntimeExperimental(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAdvancedSettings, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipUnlessMonitoringRuntimeExperimentalSupported,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.monitoring_runtime_experimental", "process"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyAdvancedSettingsFeaturesDisableAcks(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFeaturesDisableAcks, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.features_disable_policy_change_acks_enabled", "true"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyAdvancedSettingsFeaturesIncludeTags(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFeaturesIncludeTags, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.features_include_tags_in_events_enabled", "true"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyAdvancedSettingsInternal(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentInternal, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_settings.internal", agentInternalYAML),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyAdvancedMonitoring(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAdvancedMonitoring, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.http_monitoring_endpoint.enabled", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.http_monitoring_endpoint.host", "0.0.0.0"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.http_monitoring_endpoint.port", "8080"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.http_monitoring_endpoint.buffer_enabled", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.http_monitoring_endpoint.pprof_enabled", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.diagnostics.rate_limits.interval", "2m"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.diagnostics.rate_limits.burst", "5"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.diagnostics.file_uploader.init_duration", "2s"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.diagnostics.file_uploader.backoff_duration", "2m"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "advanced_monitoring_options.diagnostics.file_uploader.max_retries", "15"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicySupportsAgentless(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionSupportsAgentless, versionutils.FlavorServerless)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "supports_agentless", "true"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyInactivityTimeout(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionInactivityTimeout, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "inactivity_timeout", "2m0s"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// Changing the duration on the backing resource proves the data source
				// tracks updates rather than a value baked in at creation.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "inactivity_timeout", "5m0s"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyHostNameFormat(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionHostNameFormat, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "host_name_format", "fqdn"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// host_name_format omitted on the backing resource should surface as the
				// default "hostname" branch, which no other test asserts explicitly.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "host_name_format", "hostname"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyUnenrollmentTimeout(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionUnenrollmentTimeout, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "unenrollment_timeout", "5m0s"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// Changing the duration on the backing resource proves the data source
				// tracks updates rather than a value baked in at creation.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "unenrollment_timeout", "10m0s"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyRequiredVersions(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionRequiredVersions, versionutils.FlavorAny)

	policyName := randomPolicyName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "required_versions.%", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "required_versions.8.15.0", "100"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// A second entry proves multi-element handling of the map.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "required_versions.%", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "required_versions.8.15.0", "50"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "required_versions.8.16.0", "50"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// Removing all entries should surface as an empty collection, not a stale value.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "required_versions.%", "0"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicySpaceIDs(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionAgentPolicySpaceID, versionutils.FlavorAny)

	policyName := randomPolicyName()
	spaceID := "test-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceName := "Test Space " + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	secondSpaceID := "test-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	secondSpaceName := "Test Space " + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"space_id":    config.StringVariable(spaceID),
					"space_name":  config.StringVariable(spaceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_fleet_agent_policy.test", "space_ids.*", spaceID),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				// A second space proves multi-element handling of the set.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi"),
				ConfigVariables: config.Variables{
					"policy_name":       config.StringVariable(policyName),
					"space_id":          config.StringVariable(spaceID),
					"space_name":        config.StringVariable(spaceName),
					"second_space_id":   config.StringVariable(secondSpaceID),
					"second_space_name": config.StringVariable(secondSpaceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "space_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_fleet_agent_policy.test", "space_ids.*", spaceID),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_fleet_agent_policy.test", "space_ids.*", secondSpaceID),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAgentPolicyIsProtected(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionTamperProtection, versionutils.FlavorAny)

	policyName := randomPolicyName()
	vars := config.Variables{
		"policy_name": config.StringVariable(policyName),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("disabled"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "is_protected", "false"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defend"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "is_protected", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_elastic_defend_integration_policy.test", "id"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("enabled"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_agent_policy.test", "is_protected", "true"),
					checkAgentPolicyDataSourceMatchesAPI("data.elasticstack_fleet_agent_policy.test"),
				),
			},
		},
	})
}

func checkResourceIDsDiffer(leftName, rightName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		left, ok := state.RootModule().Resources[leftName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", leftName)
		}
		right, ok := state.RootModule().Resources[rightName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", rightName)
		}
		if left.Primary.ID == "" || right.Primary.ID == "" {
			return fmt.Errorf("expected both agent policies to have ids, got %q and %q", left.Primary.ID, right.Primary.ID)
		}
		if left.Primary.ID == right.Primary.ID {
			return fmt.Errorf("expected agent policies in different spaces to have different ids, both were %q", left.Primary.ID)
		}
		return nil
	}
}

func checkAttrEqualsResourceID(dataSourceName, attrName, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		dataSource, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found in state", dataSourceName)
		}
		fixture, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		got := dataSource.Primary.Attributes[attrName]
		if got == "" {
			return fmt.Errorf("data source %s attribute %s is empty", dataSourceName, attrName)
		}
		if got != fixture.Primary.ID {
			return fmt.Errorf("data source %s attribute %s = %q, fixture resource ID = %q", dataSourceName, attrName, got, fixture.Primary.ID)
		}
		return nil
	}
}

func checkAgentPolicyDataSourceMatchesAPI(dataSourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found in state", dataSourceName)
		}

		policyID := rs.Primary.Attributes["policy_id"]
		if policyID == "" {
			return fmt.Errorf("data source %s has an empty policy_id", dataSourceName)
		}

		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}

		policy, diags := fleet.GetAgentPolicy(context.Background(), client.GetFleetClient(), policyID, rs.Primary.Attributes["space_id"])
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if policy == nil {
			return fmt.Errorf("agent policy %q was not returned by the Fleet API", policyID)
		}

		return compareDataSourceToAPI(rs.Primary.Attributes, policy)
	}
}

func compareDataSourceToAPI(attrs map[string]string, policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) error {
	var mismatches attrMismatches
	expectAttr(&mismatches, attrs, "id", policy.Id)
	expectAttr(&mismatches, attrs, "policy_id", policy.Id)
	expectAttr(&mismatches, attrs, "name", policy.Name)
	expectAttr(&mismatches, attrs, "namespace", policy.Namespace)
	expectStringPtr(&mismatches, attrs, "description", policy.Description)
	expectStringPtr(&mismatches, attrs, "data_output_id", policy.DataOutputId)
	expectStringPtr(&mismatches, attrs, "monitoring_output_id", policy.MonitoringOutputId)
	expectStringPtr(&mismatches, attrs, "fleet_server_host_id", policy.FleetServerHostId)
	expectStringPtr(&mismatches, attrs, "download_source_id", policy.DownloadSourceId)
	expectAttr(&mismatches, attrs, "monitor_logs", strconv.FormatBool(monitoringEnabled(policy, kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabledLogs)))
	expectAttr(&mismatches, attrs, "monitor_metrics", strconv.FormatBool(monitoringEnabled(policy, kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabledMetrics)))
	expectAttr(&mismatches, attrs, "is_protected", strconv.FormatBool(policy.IsProtected))
	expectAttr(&mismatches, attrs, "host_name_format", hostNameFormatFromAPI(policy))
	expectBoolPtr(&mismatches, attrs, "supports_agentless", policy.SupportsAgentless)
	expectDurationSeconds(&mismatches, attrs, "inactivity_timeout", policy.InactivityTimeout)
	expectDurationSeconds(&mismatches, attrs, "unenrollment_timeout", policy.UnenrollTimeout)
	expectGlobalDataTags(&mismatches, attrs, policy.GlobalDataTags)
	expectSpaceIDs(&mismatches, attrs, policy.SpaceIds)
	expectRequiredVersions(&mismatches, attrs, policy.RequiredVersions)
	expectAdvancedSettings(&mismatches, attrs, policy.AdvancedSettings)
	expectAdvancedMonitoring(&mismatches, attrs, policy)
	return mismatches.err()
}

type attrMismatches struct {
	msgs []string
}

func (m *attrMismatches) add(format string, args ...any) {
	m.msgs = append(m.msgs, fmt.Sprintf(format, args...))
}

func (m *attrMismatches) err() error {
	if len(m.msgs) == 0 {
		return nil
	}
	return fmt.Errorf("data source attributes differ from the Fleet API response:\n%s", strings.Join(m.msgs, "\n"))
}

func expectAttr(m *attrMismatches, attrs map[string]string, key, want string) {
	got, ok := attrs[key]
	if !ok {
		m.add("%s: API %q, data source is null", key, want)
		return
	}
	if got != want {
		m.add("%s: API %q, data source %q", key, want, got)
	}
}

func expectFloat32Attr(m *attrMismatches, attrs map[string]string, key string, want float32) {
	got, ok := attrs[key]
	if !ok {
		m.add("%s: API %v, data source is null", key, want)
		return
	}
	parsed, err := strconv.ParseFloat(got, 32)
	if err != nil || float32(parsed) != want {
		m.add("%s: API %v, data source %q", key, want, got)
	}
}

func checkFloat32Attr(resourceName, attrName string, want float32) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		got := rs.Primary.Attributes[attrName]
		parsed, err := strconv.ParseFloat(got, 32)
		if err != nil || float32(parsed) != want {
			return fmt.Errorf("%s.%s: got %q, want %v", resourceName, attrName, got, want)
		}
		return nil
	}
}

func expectNull(m *attrMismatches, attrs map[string]string, key string) {
	if got, ok := attrs[key]; ok {
		m.add("%s: API null, data source %q", key, got)
	}
}

func expectStringPtr(m *attrMismatches, attrs map[string]string, key string, api *string) {
	if api == nil {
		expectNull(m, attrs, key)
		return
	}
	expectAttr(m, attrs, key, *api)
}

func expectBoolPtr(m *attrMismatches, attrs map[string]string, key string, api *bool) {
	if api == nil {
		expectNull(m, attrs, key)
		return
	}
	expectAttr(m, attrs, key, strconv.FormatBool(*api))
}

func expectDurationSeconds(m *attrMismatches, attrs map[string]string, key string, seconds *float32) {
	if seconds == nil {
		expectNull(m, attrs, key)
		return
	}
	d := time.Duration(int64(*seconds)) * time.Second
	expectAttr(m, attrs, key, d.String())
}

func expectNoPrefix(m *attrMismatches, attrs map[string]string, prefix string) {
	keys := keysWithPrefix(attrs, prefix)
	if len(keys) > 0 {
		m.add("%s: API null, data source has %s", prefix, strings.Join(keys, ", "))
	}
}

func keysWithPrefix(attrs map[string]string, prefix string) []string {
	dot := prefix + "."
	var keys []string
	for key := range attrs {
		if key == prefix || strings.HasPrefix(key, dot) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func monitoringEnabled(policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse, kind kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabled) bool {
	return policy.MonitoringEnabled != nil && slices.Contains(*policy.MonitoringEnabled, kind)
}

func hostNameFormatFromAPI(policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) string {
	if policy.AgentFeatures != nil {
		for _, feature := range *policy.AgentFeatures {
			if feature.Name == "fqdn" && feature.Enabled {
				return "fqdn"
			}
		}
	}
	return "hostname"
}

func expectGlobalDataTags(m *attrMismatches, attrs map[string]string, tags *[]kbapi.AgentPolicyGlobalDataTagsItem) {
	if tags == nil {
		expectNoPrefix(m, attrs, "global_data_tags")
		return
	}
	expectAttr(m, attrs, "global_data_tags.%", strconv.Itoa(len(*tags)))
	for _, tag := range *tags {
		if num, err := tag.Value.AsAgentPolicyGlobalDataTagsItemValue1(); err == nil {
			expectFloat32Attr(m, attrs, "global_data_tags."+tag.Name+".number_value", num)
			expectNull(m, attrs, "global_data_tags."+tag.Name+".string_value")
			continue
		}
		str, err := tag.Value.AsAgentPolicyGlobalDataTagsItemValue0()
		if err != nil {
			m.add("global_data_tags.%s: API value is neither a string nor a number", tag.Name)
			continue
		}
		expectAttr(m, attrs, "global_data_tags."+tag.Name+".string_value", str)
		expectNull(m, attrs, "global_data_tags."+tag.Name+".number_value")
	}
}

func expectSpaceIDs(m *attrMismatches, attrs map[string]string, spaceIDs *[]string) {
	if spaceIDs == nil {
		expectNoPrefix(m, attrs, "space_ids")
		return
	}
	expectAttr(m, attrs, "space_ids.#", strconv.Itoa(len(*spaceIDs)))
	got := indexedValues(attrs, "space_ids")
	want := append([]string(nil), (*spaceIDs)...)
	sort.Strings(got)
	sort.Strings(want)
	if !slices.Equal(got, want) {
		m.add("space_ids: API %q, data source %q", want, got)
	}
}

func expectRequiredVersions(m *attrMismatches, attrs map[string]string, versions *[]struct {
	Percentage float32 `json:"percentage"`
	Version    string  `json:"version"`
}) {
	if versions == nil {
		expectNoPrefix(m, attrs, "required_versions")
		return
	}
	expectAttr(m, attrs, "required_versions.%", strconv.Itoa(len(*versions)))
	for _, rv := range *versions {
		percentage := int32(math.Round(float64(rv.Percentage)))
		expectAttr(m, attrs, "required_versions."+rv.Version, strconv.FormatInt(int64(percentage), 10))
	}
}

func indexedValues(attrs map[string]string, prefix string) []string {
	var values []string
	for key, value := range attrs {
		rest, ok := strings.CutPrefix(key, prefix+".")
		if !ok || rest == "#" || rest == "%" {
			continue
		}
		if _, err := strconv.Atoi(rest); err != nil {
			continue
		}
		values = append(values, value)
	}
	return values
}

func expectAdvancedSettings(m *attrMismatches, attrs map[string]string, settings *struct {
	AgentDownloadTargetDirectory                any `json:"agent_download_target_directory,omitempty"`
	AgentDownloadTimeout                        any `json:"agent_download_timeout,omitempty"`
	AgentFeaturesDisablePolicyChangeAcksEnabled any `json:"agent_features_disable_policy_change_acks_enabled,omitempty"`
	AgentFeaturesIncludeTagsInEventsEnabled     any `json:"agent_features_include_tags_in_events_enabled,omitempty"`
	AgentInternal                               any `json:"agent_internal,omitempty"`
	AgentLimitsGoMaxProcs                       any `json:"agent_limits_go_max_procs,omitempty"`
	AgentLoggingFilesInterval                   any `json:"agent_logging_files_interval,omitempty"`
	AgentLoggingFilesKeepfiles                  any `json:"agent_logging_files_keepfiles,omitempty"`
	AgentLoggingFilesRotateeverybytes           any `json:"agent_logging_files_rotateeverybytes,omitempty"`
	AgentLoggingLevel                           any `json:"agent_logging_level,omitempty"`
	AgentLoggingMetricsPeriod                   any `json:"agent_logging_metrics_period,omitempty"`
	AgentLoggingToFiles                         any `json:"agent_logging_to_files,omitempty"`
	AgentMonitoringRuntimeExperimental          any `json:"agent_monitoring_runtime_experimental,omitempty"`
}) {
	if settings == nil {
		expectNoPrefix(m, attrs, "advanced_settings")
		return
	}
	expectAnyString(m, attrs, "advanced_settings.logging_level", settings.AgentLoggingLevel)
	expectAnyBool(m, attrs, "advanced_settings.logging_to_files", settings.AgentLoggingToFiles)
	expectAnyString(m, attrs, "advanced_settings.logging_files_interval", settings.AgentLoggingFilesInterval)
	expectAnyInt32(m, attrs, "advanced_settings.logging_files_keepfiles", settings.AgentLoggingFilesKeepfiles)
	expectAnyInt64(m, attrs, "advanced_settings.logging_files_rotateeverybytes", settings.AgentLoggingFilesRotateeverybytes)
	expectAnyString(m, attrs, "advanced_settings.logging_metrics_period", settings.AgentLoggingMetricsPeriod)
	expectAnyInt32(m, attrs, "advanced_settings.go_max_procs", settings.AgentLimitsGoMaxProcs)
	expectAnyString(m, attrs, "advanced_settings.download_timeout", settings.AgentDownloadTimeout)
	expectAnyString(m, attrs, "advanced_settings.download_target_directory", settings.AgentDownloadTargetDirectory)
	expectAnyString(m, attrs, "advanced_settings.monitoring_runtime_experimental", settings.AgentMonitoringRuntimeExperimental)
	expectAnyBool(m, attrs, "advanced_settings.features_disable_policy_change_acks_enabled", settings.AgentFeaturesDisablePolicyChangeAcksEnabled)
	expectAnyBool(m, attrs, "advanced_settings.features_include_tags_in_events_enabled", settings.AgentFeaturesIncludeTagsInEventsEnabled)
	expectInternal(m, attrs, settings.AgentInternal)
}

func expectAnyString(m *attrMismatches, attrs map[string]string, key string, api any) {
	str, ok := api.(string)
	if !ok {
		expectNull(m, attrs, key)
		return
	}
	expectAttr(m, attrs, key, str)
}

func expectAnyBool(m *attrMismatches, attrs map[string]string, key string, api any) {
	value, ok := api.(bool)
	if !ok {
		expectNull(m, attrs, key)
		return
	}
	expectAttr(m, attrs, key, strconv.FormatBool(value))
}

func expectAnyInt32(m *attrMismatches, attrs map[string]string, key string, api any) {
	value, ok := api.(float64)
	if !ok {
		expectNull(m, attrs, key)
		return
	}
	expectAttr(m, attrs, key, strconv.FormatInt(int64(int32(value)), 10))
}

func expectAnyInt64(m *attrMismatches, attrs map[string]string, key string, api any) {
	value, ok := api.(float64)
	if !ok {
		expectNull(m, attrs, key)
		return
	}
	expectAttr(m, attrs, key, strconv.FormatInt(int64(value), 10))
}

func expectInternal(m *attrMismatches, attrs map[string]string, api any) {
	const key = "advanced_settings.internal"
	switch internal := api.(type) {
	case nil:
		expectNull(m, attrs, key)
	case string:
		expectAttr(m, attrs, key, internal)
	default:
		encoded, err := json.Marshal(internal)
		if err != nil {
			m.add("%s: unable to encode API value: %s", key, err.Error())
			return
		}
		expectAttr(m, attrs, key, string(encoded))
	}
}

func expectAdvancedMonitoring(m *attrMismatches, attrs map[string]string, policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) {
	hasHTTP := policy.MonitoringHttp != nil || policy.MonitoringPprofEnabled != nil
	hasDiagnostics := policy.MonitoringDiagnostics != nil
	if !hasHTTP && !hasDiagnostics {
		expectNoPrefix(m, attrs, "advanced_monitoring_options")
		return
	}
	expectHTTPMonitoring(m, attrs, policy)
	expectDiagnostics(m, attrs, policy)
}

func expectHTTPMonitoring(m *attrMismatches, attrs map[string]string, policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) {
	const prefix = "advanced_monitoring_options.http_monitoring_endpoint"
	if policy.MonitoringHttp == nil && policy.MonitoringPprofEnabled == nil {
		expectNoPrefix(m, attrs, prefix)
		return
	}
	if policy.MonitoringHttp == nil {
		expectNull(m, attrs, prefix+".enabled")
		expectNull(m, attrs, prefix+".host")
		expectNull(m, attrs, prefix+".port")
		expectNull(m, attrs, prefix+".buffer_enabled")
	} else {
		expectBoolPtr(m, attrs, prefix+".enabled", policy.MonitoringHttp.Enabled)
		expectStringPtr(m, attrs, prefix+".host", policy.MonitoringHttp.Host)
		if policy.MonitoringHttp.Port == nil {
			expectNull(m, attrs, prefix+".port")
		} else {
			expectAttr(m, attrs, prefix+".port", strconv.FormatInt(int64(int32(*policy.MonitoringHttp.Port)), 10))
		}
		if policy.MonitoringHttp.Buffer == nil {
			expectNull(m, attrs, prefix+".buffer_enabled")
		} else {
			expectBoolPtr(m, attrs, prefix+".buffer_enabled", policy.MonitoringHttp.Buffer.Enabled)
		}
	}
	expectBoolPtr(m, attrs, prefix+".pprof_enabled", policy.MonitoringPprofEnabled)
}

func expectDiagnostics(m *attrMismatches, attrs map[string]string, policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) {
	const prefix = "advanced_monitoring_options.diagnostics"
	if policy.MonitoringDiagnostics == nil {
		expectNoPrefix(m, attrs, prefix)
		return
	}
	expectRateLimits(m, attrs, policy)
	expectFileUploader(m, attrs, policy)
}

func expectRateLimits(m *attrMismatches, attrs map[string]string, policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) {
	const prefix = "advanced_monitoring_options.diagnostics.rate_limits"
	if policy.MonitoringDiagnostics == nil || policy.MonitoringDiagnostics.Limit == nil {
		expectNoPrefix(m, attrs, prefix)
		return
	}
	expectStringPtr(m, attrs, prefix+".interval", policy.MonitoringDiagnostics.Limit.Interval)
	if policy.MonitoringDiagnostics.Limit.Burst == nil {
		expectNull(m, attrs, prefix+".burst")
		return
	}
	expectAttr(m, attrs, prefix+".burst", strconv.FormatInt(int64(int32(*policy.MonitoringDiagnostics.Limit.Burst)), 10))
}

func expectFileUploader(m *attrMismatches, attrs map[string]string, policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse) {
	const prefix = "advanced_monitoring_options.diagnostics.file_uploader"
	if policy.MonitoringDiagnostics == nil || policy.MonitoringDiagnostics.Uploader == nil {
		expectNoPrefix(m, attrs, prefix)
		return
	}
	expectStringPtr(m, attrs, prefix+".init_duration", policy.MonitoringDiagnostics.Uploader.InitDur)
	expectStringPtr(m, attrs, prefix+".backoff_duration", policy.MonitoringDiagnostics.Uploader.MaxDur)
	if policy.MonitoringDiagnostics.Uploader.MaxRetries == nil {
		expectNull(m, attrs, prefix+".max_retries")
		return
	}
	expectAttr(m, attrs, prefix+".max_retries", strconv.FormatInt(int64(int32(*policy.MonitoringDiagnostics.Uploader.MaxRetries)), 10))
}

func checkResourceAgentPolicyDestroy(state *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_agent_policy" || rs.Primary == nil {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_ids.0"]
		if spaceID == "" {
			spaceID = rs.Primary.Attributes["space_id"]
		}
		policy, diags := fleet.GetAgentPolicy(context.Background(), client.GetFleetClient(), rs.Primary.ID, spaceID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if policy != nil {
			return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
