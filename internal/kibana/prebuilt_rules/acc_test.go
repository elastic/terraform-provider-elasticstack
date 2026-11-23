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

package prebuiltrules_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var (
	minVersionPrebuiltRules = version.Must(version.NewVersion("8.0.0"))
	// 9.4.x and 9.5.x install-all in a new custom space 400s on a deprecated rule stub
	// (elastic/kibana#285497). go-version constraints are AND-only, so the supported range
	// is expressed as alternatives (OR): < 9.4.0 or >= 9.6.0 (including 9.6.0 prereleases).
	prebuiltRulesInSpaceConstraints = []version.Constraints{
		version.MustConstraints(version.NewConstraint("< 9.4.0")),
		version.MustConstraints(version.NewConstraint(">= 9.6.0-0")),
	}
)

func TestPrebuiltRulesInSpaceConstraints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		allowed bool
	}{
		{name: "8.19.21", version: "8.19.21", allowed: true},
		{name: "9.3.9", version: "9.3.9", allowed: true},
		{name: "9.4.0", version: "9.4.0", allowed: false},
		{name: "9.4.6", version: "9.4.6", allowed: false},
		{name: "9.4.8", version: "9.4.8", allowed: false},
		{name: "9.5.0", version: "9.5.0", allowed: false},
		{name: "9.5.3", version: "9.5.3", allowed: false},
		{name: "9.5.5", version: "9.5.5", allowed: false},
		{name: "9.6.0-SNAPSHOT", version: "9.6.0-SNAPSHOT", allowed: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v := version.Must(version.NewVersion(tc.version))
			allowed := false
			for _, c := range prebuiltRulesInSpaceConstraints {
				allowed = allowed || c.Check(v)
			}
			require.Equal(t, tc.allowed, allowed)
		})
	}
}

func TestAccResourcePrebuiltRules(t *testing.T) {
	testCases := []struct {
		name        string
		spaceID     string
		constraints []version.Constraints
	}{
		{
			name:    "default",
			spaceID: "default",
		},
		{
			name:        "in_space",
			spaceID:     "security_rules" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum),
			constraints: prebuiltRulesInSpaceConstraints,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.constraints) > 0 {
				versionutils.SkipIfUnsupportedAnyConstraints(t, versionutils.FlavorAny, tc.constraints...)
			}
			testAccResourcePrebuiltRules(t, tc.spaceID)
		})
	}
}

func testAccResourcePrebuiltRules(t *testing.T, spaceID string) {
	versionutils.SkipIfUnsupported(t, minVersionPrebuiltRules, versionutils.FlavorAny)

	// Captured from the initial create step so the post-reinstall step can
	// assert the dataset returned to its original, deterministic shape
	// instead of only checking that the counters are set.
	var initialRulesInstalled, initialTimelinesInstalled string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "space_id", spaceID),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_updated"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_updated"),
					resource.TestCheckResourceAttrWith("elasticstack_kibana_install_prebuilt_rules.test", "rules_installed", func(value string) error {
						initialRulesInstalled = value
						return nil
					}),
					resource.TestCheckResourceAttrWith("elasticstack_kibana_install_prebuilt_rules.test", "timelines_installed", func(value string) error {
						initialTimelinesInstalled = value
						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				PreConfig: func() {
					deleteSingleDetectionRule(t, spaceID)
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectKnownValue("elasticstack_kibana_install_prebuilt_rules.test", tfjsonpath.New("rules_not_installed"), knownvalue.Int64Exact(0)),
						plancheck.ExpectKnownValue("elasticstack_kibana_install_prebuilt_rules.test", tfjsonpath.New("rules_not_updated"), knownvalue.Int64Exact(0)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_installed", "0"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_installed", "0"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_updated"),
					resource.TestCheckResourceAttrWith("elasticstack_kibana_install_prebuilt_rules.test", "rules_installed", func(value string) error {
						if value != initialRulesInstalled {
							return fmt.Errorf("expected rules_installed to return to its pre-delete value %q after reinstall, got %q", initialRulesInstalled, value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("elasticstack_kibana_install_prebuilt_rules.test", "timelines_installed", func(value string) error {
						if value != initialTimelinesInstalled {
							return fmt.Errorf("expected timelines_installed to remain unchanged at %q after reinstall, got %q", initialTimelinesInstalled, value)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccResourcePrebuiltRulesDefaultSpaceID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionPrebuiltRules, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "space_id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_updated"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_updated"),
				),
			},
		},
	})
}

// TestAccResourcePrebuiltRulesSpaceIDForcesReplace verifies that changing
// space_id on an existing prebuilt rules resource actually triggers the
// RequiresReplace plan modifier, not just that the resource defaults to and
// reflects a given space_id.
func TestAccResourcePrebuiltRulesSpaceIDForcesReplace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionPrebuiltRules, versionutils.FlavorAny)
	versionutils.SkipIfUnsupportedAnyConstraints(t, versionutils.FlavorAny, prebuiltRulesInSpaceConstraints...)

	spaceAID := "security_rules_a_" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	spaceBID := "security_rules_b_" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_install_prebuilt_rules.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_replace"),
				ConfigVariables: config.Variables{
					"space_a_id":      config.StringVariable(spaceAID),
					"space_b_id":      config.StringVariable(spaceBID),
					"active_space_id": config.StringVariable(spaceAID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceID, "id"),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceAID),
					resource.TestCheckResourceAttrSet(resourceID, "rules_installed"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_replace"),
				ConfigVariables: config.Variables{
					"space_a_id":      config.StringVariable(spaceAID),
					"space_b_id":      config.StringVariable(spaceBID),
					"active_space_id": config.StringVariable(spaceBID),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceID, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceID, "id"),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceBID),
					resource.TestCheckResourceAttrSet(resourceID, "rules_installed"),
				),
			},
		},
	})
}

// TestAccResourcePrebuiltRulesKibanaConnection exercises the kibana_connection
// override block, which is injected by the shared Kibana resource envelope but
// had no dedicated fixture for this resource.
func TestAccResourcePrebuiltRulesKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionPrebuiltRules, versionutils.FlavorAny)

	resourceID := "elasticstack_kibana_install_prebuilt_rules.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("kibana_connection"),
				ConfigVariables:          acctest.KibanaConnectionVariables(),
				Check:                    prebuiltRulesKibanaConnectionChecks(resourceID),
			},
		},
	})
}

func prebuiltRulesKibanaConnectionChecks(resourceID string) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceID, "id"),
		resource.TestCheckResourceAttr(resourceID, "space_id", "default"),
		resource.TestCheckResourceAttrSet(resourceID, "rules_installed"),
		resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr(resourceID, "kibana_connection.0.endpoints.0", acctest.KibanaConnectionEndpoint()),
		resource.TestCheckResourceAttr(resourceID, "kibana_connection.0.insecure", "false"),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks(resourceID)...)

	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func deleteSingleDetectionRule(t *testing.T, spaceID string) {
	unsupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionPrebuiltRules)()
	require.NoError(t, err)

	if unsupported {
		return
	}

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient := client.GetKibanaOapiClient()

	resp, err := oapiClient.API.FindRulesWithResponse(t.Context(), &kbapi.FindRulesParams{}, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())

	ruleBytes, err := resp.JSON200.Data[0].MarshalJSON()
	require.NoError(t, err)

	var ruleMap map[string]any
	err = json.Unmarshal(ruleBytes, &ruleMap)
	require.NoError(t, err)

	id, ok := ruleMap["id"].(string)
	require.True(t, ok, "rule ID not found or not a string")

	idUUID, err := uuid.Parse(id)
	require.NoError(t, err)

	deleteResp, err := oapiClient.API.DeleteRuleWithResponse(t.Context(), spaceID, &kbapi.DeleteRuleParams{Id: &idUUID})
	require.NoError(t, err)
	require.Equal(t, 200, deleteResp.StatusCode())
}
