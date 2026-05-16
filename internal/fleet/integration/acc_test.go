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
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var (
	minVersionIntegration       = version.Must(version.NewVersion("8.6.0"))
	minVersionIntegrationPolicy = version.Must(version.NewVersion("8.10.0"))
	// minVersionInstallationInfo is the first Fleet release that populates
	// PackageInfo.InstallationInfo on GET /epm/packages/{name}/{version}.
	// Tests that rely on the actually-installed version (as opposed to the
	// version echoed back from the request path) require this floor.
	minVersionInstallationInfo = version.Must(version.NewVersion("8.9.0"))
	minVersionSpaceIDReadback  = version.Must(version.NewVersion("8.10.0"))
)

//go:embed testdata/TestAccResourceIntegrationFromSDK/main.tf
var testAccResourceIntegrationFromSDKConfig string

//go:embed testdata/TestAccResourceIntegrationFrom0_13_1/sdk/main.tf
var testAccResourceIntegrationFrom013SDKConfig string

func TestAccResourceIntegrationFromSDK(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegration, versionutils.FlavorAny)

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
				Config: testAccResourceIntegrationFromSDKConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
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
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSpaceIDReadback),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ResourceName:             "elasticstack_fleet_integration.test_integration",
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ImportState:              true,
				ImportStateVerify:        true,
				ExpectError:              regexp.MustCompile("Resource Import Not Implemented"),
			},
		},
	})
}

// TestAccResourceIntegration_kibanaConnection exercises the kibana_connection block
// (scoped Kibana client via r.Client). Import is not implemented for this resource.
func TestAccResourceIntegration_kibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegration, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          acctest.KibanaConnectionVariables(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "kibana_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration.test_integration", "kibana_connection.0.endpoints.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          acctest.KibanaConnectionVariables(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.17.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "kibana_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration.test_integration", "kibana_connection.0.endpoints.0"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationWithPolicy(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegrationPolicy, versionutils.FlavorAny)

	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("v1_16_0"),
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("v1_17_0"),
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
				ResourceName:             "elasticstack_fleet_integration.test_integration",
				ConfigDirectory:          acctest.NamedTestCaseDirectory("v1_17_0"),
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
	versionutils.SkipIfUnsupported(t, minVersionIntegration, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "sysmon_linux"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.7.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				// Force uninstall the integration
				PreConfig: func() {
					fleetClient, err := testAccFleetClient()
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

// TestAccResourceIntegration_ExternalChange asserts that out-of-band version
// changes to an installed Fleet integration package are detected on the next
// refresh, and that terraform plan surfaces the drift.
//
// Regression test for https://github.com/elastic/terraform-provider-elasticstack/issues/1585:
// Fleet's GET /epm/packages/{name}/{version} returns status "installed"
// whenever the package is installed at *any* version, so the provider
// previously did not notice out-of-band upgrades. Read now consults
// InstallationInfo.Version and records the actually-installed version in
// state, so plan sees a diff between state and config.
func TestAccResourceIntegration_ExternalChange(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionInstallationInfo, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: apply tcp@1.16.0 via terraform.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			// Step 2: upgrade tcp to 1.17.0 via the Fleet API (out-of-band).
			// The next refresh must record the *actually installed* version
			// (1.17.0) in state, and the resulting plan must be non-empty
			// because the configured version (1.16.0) no longer matches
			// state. Relies on PackageInfo.InstallationInfo being populated
			// by Fleet, which is why the step is gated on
			// minVersionInstallationInfo rather than the broader
			// minVersionIntegration — on pre-8.9 servers the Read fallback
			// returns the requested (state) version and no drift is
			// observable through the integration GET alone.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				PreConfig: func() {
					notSupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionIntegration)()
					require.NoError(t, err)
					if notSupported {
						return
					}

					fleetClient, err := testAccFleetClient()
					require.NoError(t, err)

					diags := fleet.InstallPackage(t.Context(), fleetClient, "tcp", "1.17.0", fleet.InstallPackageOptions{
						Force: true,
					})
					require.Empty(t, diags)
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.17.0"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationWithPrerelease(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegration, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
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
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_params_step1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "prerelease", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "skip_destroy", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "ignore_constraints", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration.test_integration_all_params", "version"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(integration.MinVersionIgnoreMappingUpdateErrors),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_params_step2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "prerelease", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "skip_destroy", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "ignore_mapping_update_errors", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "skip_data_stream_rollover", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "ignore_constraints", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration.test_integration_all_params", "version"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_params_step1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "name", "tcp"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "ignore_mapping_update_errors"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration.test_integration_all_params", "skip_data_stream_rollover"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationFrom0_13_1(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegration, versionutils.FlavorAny)

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
				Config: testAccResourceIntegrationFrom013SDKConfig,
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_upgrade", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration_upgrade", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
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

// testAccCheckIntegrationInstalled queries the Fleet API to verify that the
// given package version is installed globally (no space scoping).
func testAccCheckIntegrationInstalled(pkgName, pkgVersion string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		fleetClient, err := testAccFleetClient()
		if err != nil {
			return err
		}
		pkg, diags := fleet.GetPackage(context.Background(), fleetClient, pkgName, pkgVersion, "")
		if diags.HasError() {
			return fmt.Errorf("failed to get package: %v", diags)
		}
		if pkg == nil {
			return fmt.Errorf("package %s/%s not installed", pkgName, pkgVersion)
		}
		installed := false
		if pkg.InstallationInfo != nil {
			installed = pkg.InstallationInfo.InstallStatus == kbapi.PackageInfoInstallationInfoInstallStatusInstalled
		}
		if !installed && pkg.Status != nil && strings.EqualFold(*pkg.Status, "installed") {
			installed = true
		}
		if !installed {
			return fmt.Errorf("package %s/%s is not installed", pkgName, pkgVersion)
		}
		return nil
	}
}

// TestAccResourceIntegrationSkipDestroy verifies that when skip_destroy = true,
// Terraform removes the resource from state on destroy but leaves the integration
// package installed in Fleet.
func TestAccResourceIntegrationSkipDestroy(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegration, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: Install tcp@1.16.0 with skip_destroy = true; assert attributes.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_skip_destroy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_skip_destroy", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_skip_destroy", "version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_skip_destroy", "skip_destroy", "true"),
				),
			},
			// Step 2: Remove the resource from config. Terraform calls Delete, but
			// skip_destroy = true means no actual uninstall occurs. Check that the
			// package is still installed via the Fleet API.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_config"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationInstalled("tcp", "1.16.0"),
				),
			},
			// Step 3: Reinstall without skip_destroy so the test framework's automatic
			// final destroy properly uninstalls the package on cleanup.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_skip_destroy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_skip_destroy", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_skip_destroy", "version", "1.16.0"),
				),
			},
		},
	})
}

// testAccCheckIntegrationInstalledInSpace queries the Fleet API to verify that
// the given package version has Kibana assets installed in the specified space.
func testAccCheckIntegrationInstalledInSpace(name, version, spaceID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		fleetClient, err := testAccFleetClient()
		if err != nil {
			return err
		}
		pkg, diags := fleet.GetPackage(context.Background(), fleetClient, name, version, spaceID)
		if diags.HasError() {
			return fmt.Errorf("failed to get package: %v", diags)
		}
		if pkg == nil {
			return fmt.Errorf("package %s/%s not installed", name, version)
		}
		globalInstalled := false
		if pkg.InstallationInfo != nil {
			globalInstalled = pkg.InstallationInfo.InstallStatus == kbapi.PackageInfoInstallationInfoInstallStatusInstalled
		}
		if !globalInstalled && pkg.Status != nil && strings.EqualFold(*pkg.Status, "installed") {
			globalInstalled = true
		}
		if !globalInstalled {
			return fmt.Errorf("package %s/%s not globally installed", name, version)
		}
		inSpace := pkg.InstallationInfo.InstalledKibanaSpaceId != nil && *pkg.InstallationInfo.InstalledKibanaSpaceId == spaceID

		if pkg.InstallationInfo.AdditionalSpacesInstalledKibana != nil {
			if _, ok := (*pkg.InstallationInfo.AdditionalSpacesInstalledKibana)[spaceID]; ok {
				inSpace = true
			}
		}
		if !inSpace {
			return fmt.Errorf("package %s/%s not installed in space %s", name, version, spaceID)
		}
		return nil
	}
}

func testAccFleetClient() (*fleet.Client, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, err
	}

	return client.GetFleetClient()
}

func preinstallTCPDefault(t *testing.T) func() {
	return func() {
		fleetClient, err := testAccFleetClient()
		require.NoError(t, err)
		diags := fleet.InstallPackage(t.Context(), fleetClient, "tcp", "1.16.0", fleet.InstallPackageOptions{
			Force: true,
		})
		require.Empty(t, diags)
	}
}

// TestAccResourceIntegration_MultiSpaceInstall verifies that the same package
// can be installed in two different Kibana spaces when both resources are
// managed by Terraform. The package is pre-installed in the default space so
// that each space-scoped resource triggers the kibana_assets endpoint rather
// than a full global install.
func TestAccResourceIntegration_MultiSpaceInstall(t *testing.T) {
	versionutils.SkipIfUnsupported(t, integration.MinVersionSpaceAwareIntegration, versionutils.FlavorAny)

	spaceA := "test_a_" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	spaceB := "test_b_" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_a": config.StringVariable(spaceA),
					"space_b": config.StringVariable(spaceB),
				},
				PreConfig: preinstallTCPDefault(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "space_id", spaceA),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_b", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_b", "version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_b", "space_id", spaceB),
					testAccCheckIntegrationInstalledInSpace("tcp", "1.16.0", spaceA),
					testAccCheckIntegrationInstalledInSpace("tcp", "1.16.0", spaceB),
				),
			},
		},
	})
}

// TestAccResourceIntegration_MultiSpaceDelete verifies that destroying a
// resource for one space does not remove the package from another space. The
// package is pre-installed in the default space, then installed in space A and
// space B. When the space B resource is removed, the package must remain
// installed in space A.
func TestAccResourceIntegration_MultiSpaceDelete(t *testing.T) {
	versionutils.SkipIfUnsupported(t, integration.MinVersionSpaceAwareIntegration, versionutils.FlavorAny)

	spaceA := "test_a_" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	spaceB := "test_b_" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_a": config.StringVariable(spaceA),
					"space_b": config.StringVariable(spaceB),
				},
				PreConfig: preinstallTCPDefault(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_b", "name", "tcp"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_a_only"),
				ConfigVariables: config.Variables{
					"space_a": config.StringVariable(spaceA),
					"space_b": config.StringVariable(spaceB),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "space_id", spaceA),
					testAccCheckIntegrationInstalledInSpace("tcp", "1.16.0", spaceA),
				),
			},
		},
	})
}

// TestAccResourceIntegration_SpaceAwareDrift verifies that if Kibana assets
// for a package are manually removed from a space via the API, Terraform
// detects the drift on the next plan and wants to re-create the resource. The
// package is pre-installed in the default space so that removing assets from
// the target space does not remove the global installation record.
func TestAccResourceIntegration_SpaceAwareDrift(t *testing.T) {
	versionutils.SkipIfUnsupported(t, integration.MinVersionSpaceAwareIntegration, versionutils.FlavorAny)

	spaceA := "test_a_" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_a": config.StringVariable(spaceA),
				},
				PreConfig: preinstallTCPDefault(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_a", "space_id", spaceA),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_a": config.StringVariable(spaceA),
				},
				PreConfig: func() {
					fleetClient, err := testAccFleetClient()
					require.NoError(t, err)
					diags := fleet.DeleteKibanaAssets(t.Context(), fleetClient, "tcp", "1.16.0", spaceA, true)
					require.Empty(t, diags)
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccResourceIntegration_destroyWithILMCrossDependency validates that
// destroying an ILM policy succeeds even when a Fleet-managed backing index
// still references it. Regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/1999.
func TestAccResourceIntegration_destroyWithILMCrossDependency(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegration, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: Install the system integration, create an ILM policy and
			// attach it to the Fleet-managed index template.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "system"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.18.0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_lifecycle.test", "name", "test-fleet-ilm-policy"),
				),
			},
			// Step 2: Create a data stream so Elasticsearch creates a backing
			// index that references the ILM policy via the Fleet-managed template.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				PreConfig: func() {
					notSupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionIntegration)()
					require.NoError(t, err)
					if notSupported {
						return
					}

					client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
					require.NoError(t, err)
					ctx := context.Background()

					diags := esclient.PutDataStream(ctx, client, "logs-system.syslog-default")
					require.Empty(t, diags)

					indices, fwDiags := esclient.GetIndicesWithILMPolicy(ctx, client, "test-fleet-ilm-policy")
					require.False(t, fwDiags.HasError(), "unexpected error getting indices with ILM policy: %v", fwDiags.Errors())
					require.NotEmpty(t, indices, "expected at least one backing index with index.lifecycle.name = test-fleet-ilm-policy")
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "system"),
				),
			},
			// Step 3: Destroy the ILM policy (force_destroy clears references first).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             "elasticstack_elasticsearch_index_lifecycle.test",
				Destroy:                  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "system"),
				),
			},
			// Step 4: Remove the data stream so the implicit terraform destroy at
			// the end of the test case can clean up the remaining resources.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				PreConfig: func() {
					notSupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionIntegration)()
					require.NoError(t, err)
					if notSupported {
						return
					}

					client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
					require.NoError(t, err)
					ctx := context.Background()

					diags := esclient.DeleteDataStream(ctx, client, "logs-system.syslog-default")
					require.Empty(t, diags)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "system"),
				),
			},
		},
	})
}
