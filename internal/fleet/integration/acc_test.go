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

package integration_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	minVersionIntegration       = version.Must(version.NewVersion("8.6.0"))
	minVersionIntegrationPolicy = version.Must(version.NewVersion("8.10.0"))
)

func TestAccResourceIntegrationFromSDK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
		},
	})
}

func TestAccResourceIntegration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ResourceName:      "elasticstack_fleet_integration.test_integration",
				ConfigDirectory:   acctest.NamedTestCaseDirectory("create"),
				ImportState:       true,
				ImportStateVerify: true,
				ExpectError:       regexp.MustCompile("Resource Import Not Implemented"),
			},
		},
	})
}

func TestAccResourceIntegrationWithPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory: acctest.NamedTestCaseDirectory("v1_16_0"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory: acctest.NamedTestCaseDirectory("v1_17_0"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.17.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ResourceName:    "elasticstack_fleet_integration.test_integration",
				ConfigDirectory: acctest.NamedTestCaseDirectory("v1_17_0"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ExpectError:       regexp.MustCompile("Resource Import Not Implemented"),
			},
		},
	})
}

func TestAccResourceIntegrationDeleted(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "sysmon_linux"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.7.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				// Force uninstall the integration
				PreConfig: func() {
					client, err := clients.NewAcceptanceTestingClient()
					require.NoError(t, err)

					fleetClient, err := client.GetFleetClient()
					require.NoError(t, err)

					ctx := context.Background()
					diags := fleet.Uninstall(ctx, fleetClient, "sysmon_linux", "1.7.0", "", true)
					require.Empty(t, diags)
				},
				// Expect the plan to want to reinstall
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccResourceIntegration_ExternalChange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				PreConfig: func() {
					notSupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionIntegration)()
					require.NoError(t, err)

					// Skip the pre-config if the version is not supported
					if notSupported {
						return
					}

					client, err := clients.NewAcceptanceTestingClient()
					require.NoError(t, err)

					fleetClient, err := client.GetFleetClient()
					require.NoError(t, err)

					diags := fleet.InstallPackage(t.Context(), fleetClient, "tcp", "1.17.0", fleet.InstallPackageOptions{
						Force: true,
					})
					require.Empty(t, diags)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				PreConfig: func() {
					notSupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionIntegration)()
					require.NoError(t, err)

					// Skip the pre-config if the version is not supported
					if notSupported {
						return
					}

					client, err := clients.NewAcceptanceTestingClient()
					require.NoError(t, err)

					fleetClient, err := client.GetFleetClient()
					require.NoError(t, err)

					diags := fleet.Uninstall(t.Context(), fleetClient, "tcp", "1.16.0", "", true)
					require.Empty(t, diags)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationWithPrerelease(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_prerelease", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_prerelease", "prerelease", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration.test_integration_prerelease", "version"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationWithAllParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory: acctest.NamedTestCaseDirectory("all_params_step1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "prerelease", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "ignore_constraints", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration.test_integration_all_params", "version"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(integration.MinVersionIgnoreMappingUpdateErrors),
				ConfigDirectory: acctest.NamedTestCaseDirectory("all_params_step2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "prerelease", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "ignore_mapping_update_errors", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "skip_data_stream_rollover", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "ignore_constraints", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration.test_integration_all_params", "version"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationFrom0_13_1(t *testing.T) {
	spaceID := "aa_test_space_" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.13.1",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				Config:   testAccResourceIntegrationV0(spaceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_upgrade", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_upgrade", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("upgrade"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_upgrade", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_upgrade", "version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_upgrade", "space_id", spaceID),
				),
			},
		},
	})
}

func testAccResourceIntegrationV0(spaceID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = "%s"
  name     = "Test Space"
}

resource "elasticstack_fleet_integration" "test_integration_upgrade" {
  name         = "tcp"
  version      = "1.16.0"
  space_ids    = [elasticstack_kibana_space.test.space_id, "default"]
  force        = true
  skip_destroy = true
}
`, spaceID)
}
