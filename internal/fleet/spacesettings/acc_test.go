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

package spacesettings_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionFleetSpaceSettings = version.Must(version.NewVersion("9.1.0"))

func TestAccResourceFleetSpaceSettings(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetSpaceSettings, versionutils.FlavorAny)

	spaceID := fmt.Sprintf("tf-acc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha))
	variables := config.Variables{"space_id": config.StringVariable(spaceID)}
	const resourceName = "elasticstack_fleet_space_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetSpaceSettingsReset,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "allowed_namespace_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
					resource.TestCheckNoResourceAttr(resourceName, "managed_by"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "allowed_namespace_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "shared"),
					resource.TestCheckNoResourceAttr(resourceName, "managed_by"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          variables,
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateId:            spaceID,
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "allowed_namespace_prefixes.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceFleetSpaceSettings_importDefaultSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetSpaceSettings, versionutils.FlavorAny)

	const resourceName = "elasticstack_fleet_space_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetSpaceSettingsReset,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "default"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateId:            "default",
				ImportStateVerify:        true,
			},
		},
	})
}

func TestAccResourceFleetSpaceSettings_drift(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetSpaceSettings, versionutils.FlavorAny)

	spaceID := fmt.Sprintf("tf-acc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha))
	variables := config.Variables{"space_id": config.StringVariable(spaceID)}
	const resourceName = "elasticstack_fleet_space_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetSpaceSettingsReset,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          variables,
				Check:                    resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
			},
			{
				PreConfig:                func() { putFleetSpaceSettingsOutOfBand(t, spaceID, []string{"team_b"}) },
				ProtoV6ProviderFactories: acctest.Providers,
				RefreshState:             true,
				ExpectNonEmptyPlan:       true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "allowed_namespace_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_b"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          variables,
				Check:                    resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
			},
		},
	})
}

func TestAccResourceFleetSpaceSettings_destroyResetsPrefixes(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetSpaceSettings, versionutils.FlavorAny)

	spaceID := fmt.Sprintf("tf-acc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha))
	variables := config.Variables{"space_id": config.StringVariable(spaceID)}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          variables,
				Check:                    resource.TestCheckTypeSetElemAttr("elasticstack_fleet_space_settings.test", "allowed_namespace_prefixes.*", "team_a"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_removed"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "space_id", spaceID),
					checkSpaceSettingsHaveNoPrefixes(spaceID),
				),
			},
		},
	})
}

func TestAccResourceFleetSpaceSettings_kibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetSpaceSettings, versionutils.FlavorAny)

	spaceID := fmt.Sprintf("tf-acc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha))
	const resourceName = "elasticstack_fleet_space_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		CheckDestroy: checkFleetSpaceSettingsReset,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"space_id": config.StringVariable(spaceID),
				}),
				Check: resource.ComposeTestCheckFunc(append([]resource.TestCheckFunc{
					resource.TestCheckResourceAttr(resourceName, "id", spaceID),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
					resource.TestCheckResourceAttr(resourceName, "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kibana_connection.0.insecure", "true"),
					resource.TestCheckResourceAttr(resourceName, "kibana_connection.0.endpoints.#", "1"),
				}, acctest.KibanaConnectionAuthChecks(resourceName)...)...),
			},
		},
	})
}

func TestAccResourceFleetSpaceSettings_validation_tooManyPrefixes(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetSpaceSettings, versionutils.FlavorAny)

	spaceID := fmt.Sprintf("tf-acc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha))
	prefixes := make([]string, 0, 11)
	for i := range 11 {
		prefixes = append(prefixes, fmt.Sprintf("prefix_%d", i))
	}
	variables := config.Variables{
		"space_id":                   config.StringVariable(spaceID),
		"allowed_namespace_prefixes": config.SetVariable(toStringVariables(prefixes)...),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid"),
				ConfigVariables:          variables,
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?is)at most 10 elements,\s+got: 11`),
			},
		},
	})
}

func toStringVariables(values []string) []config.Variable {
	variables := make([]config.Variable, 0, len(values))
	for _, value := range values {
		variables = append(variables, config.StringVariable(value))
	}
	return variables
}

func TestAccResourceFleetSpaceSettings_spaceIDForcesReplace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetSpaceSettings, versionutils.FlavorAny)

	spaceID1 := fmt.Sprintf("tf-acc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha))
	spaceID2 := fmt.Sprintf("tf-acc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha))
	variables := config.Variables{
		"space_id_1": config.StringVariable(spaceID1),
		"space_id_2": config.StringVariable(spaceID2),
	}
	const resourceName = "elasticstack_fleet_space_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          variables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID1),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID1),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID2),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID2),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_namespace_prefixes.*", "team_a"),
					checkSpaceSettingsHaveNoPrefixes(spaceID1),
				),
			},
		},
	})
}

func checkSpaceSettingsHaveNoPrefixes(spaceID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}

		settings, diags := fleet.GetSpaceSettings(context.Background(), client.GetFleetClient(), spaceID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if settings == nil {
			return fmt.Errorf("fleet space settings for space %q not found although the space still exists", spaceID)
		}
		if len(settings.AllowedNamespacePrefixes) > 0 {
			return fmt.Errorf("fleet space settings for space %q still restrict namespaces to %v", spaceID, settings.AllowedNamespacePrefixes)
		}
		return nil
	}
}

func putFleetSpaceSettingsOutOfBand(t *testing.T, spaceID string, prefixes []string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatal(err)
	}
	if diags := fleet.PutSpaceSettings(context.Background(), client.GetFleetClient(), spaceID, prefixes); diags.HasError() {
		t.Fatal(diagutil.FwDiagsAsError(diags))
	}
}

func checkFleetSpaceSettingsReset(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_space_settings" {
			continue
		}

		settings, diags := fleet.GetSpaceSettings(context.Background(), client.GetFleetClient(), rs.Primary.ID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if settings != nil && len(settings.AllowedNamespacePrefixes) > 0 {
			return fmt.Errorf("fleet space settings for space %q still restrict namespaces to %v", rs.Primary.ID, settings.AllowedNamespacePrefixes)
		}
	}
	return nil
}
