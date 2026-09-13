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

package serverhost_test

import (
	"context"
	_ "embed"
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

var (
	minVersionFleetServerHost       = version.Must(version.NewVersion("8.6.0"))
	minVersionFleetServerHostSpaces = version.Must(version.NewVersion("9.1.0"))
)

//go:embed testdata/TestAccResourceFleetServerHostFromSDK/create/main.tf
var testAccResourceFleetServerHostFromSDKConfig string

func TestAccResourceFleetServerHostFromSDK(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)
	hostID := fmt.Sprintf("fleet-server-host-%s", sdkacctest.RandString(12))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				Config: testAccResourceFleetServerHostFromSDKConfig,
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", policyName)),
					"host_id": config.StringVariable(hostID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", hostID),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", policyName)),
					"host_id": config.StringVariable(hostID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", hostID),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
		},
	})
}

func TestAccResourceFleetServerHost(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)
	hostID := fmt.Sprintf("fleet-server-host-%s", sdkacctest.RandString(12))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", policyName)),
					"host_id": config.StringVariable(hostID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", hostID),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("Updated FleetServerHost %s", policyName)),
					"host_id": config.StringVariable(hostID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("Updated FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", hostID),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.1", "https://fleet-server-2:8220"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("Updated FleetServerHost %s", policyName)),
					"host_id": config.StringVariable(hostID),
				},
				ResourceName:      "elasticstack_fleet_server_host.test_host",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceFleetServerHost_computedID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	hostName := sdkacctest.RandString(22)
	var capturedHostID string

	captureHostID := func(value string) error {
		capturedHostID = value
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_computed_id", "name", fmt.Sprintf("FleetServerHost %s", hostName)),
					resource.TestCheckResourceAttrWith("elasticstack_fleet_server_host.test_computed_id", "host_id", captureHostID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Updated FleetServerHost %s", hostName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_computed_id", "name", fmt.Sprintf("Updated FleetServerHost %s", hostName)),
					resource.TestCheckResourceAttrWith("elasticstack_fleet_server_host.test_computed_id", "host_id", func(value string) error {
						if value != capturedHostID {
							return fmt.Errorf("expected host_id to be unchanged from previous step, was [%s], now [%s]", capturedHostID, value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_computed_id", "hosts.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_computed_id", "hosts.0", "https://fleet-server:8220"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_computed_id", "hosts.1", "https://fleet-server-2:8220"),
				),
			},
		},
	})
}

func TestAccResourceFleetServerHost_importFromSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSpaces, versionutils.FlavorAny)

	hostName := sdkacctest.RandString(22)
	spaceName := sdkacctest.RandString(22)
	spaceID := fmt.Sprintf("fleet-server-host-test-%s", spaceName)
	hostID := fmt.Sprintf("fleet-server-host-%s", sdkacctest.RandString(12))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			// Create a server host in a space.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":       config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id":    config.StringVariable(hostID),
					"space_id":   config.StringVariable(spaceID),
					"space_name": config.StringVariable(fmt.Sprintf("Fleet Server Host Test Space %s", spaceName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", hostName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_server_host.test_host", "space_ids.*", spaceID),
				),
			},
			// Scenario 1: composite ID import (<space>/<host_id>) — space_ids is populated.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":       config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id":    config.StringVariable(hostID),
					"space_id":   config.StringVariable(spaceID),
					"space_name": config.StringVariable(fmt.Sprintf("Fleet Server Host Test Space %s", spaceName)),
				},
				ResourceName:            "elasticstack_fleet_server_host.test_host",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"space_ids"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources["elasticstack_fleet_server_host.test_host"]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource elasticstack_fleet_server_host.test_host not found in state")
					}
					return fmt.Sprintf("%s/%s", spaceID, res.Primary.Attributes["host_id"]), nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", hostName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_server_host.test_host", "space_ids.*", spaceID),
				),
			},
			// Scenario 2: plain ID import (no space prefix) — space_ids is NOT set.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":       config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id":    config.StringVariable(hostID),
					"space_id":   config.StringVariable(spaceID),
					"space_name": config.StringVariable(fmt.Sprintf("Fleet Server Host Test Space %s", spaceName)),
				},
				ResourceName:            "elasticstack_fleet_server_host.test_host",
				ImportState:             true,
				ImportStateVerify:       false,
				ImportStateVerifyIgnore: []string{"space_ids"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources["elasticstack_fleet_server_host.test_host"]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource elasticstack_fleet_server_host.test_host not found in state")
					}
					return res.Primary.Attributes["host_id"], nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", hostName)),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_server_host.test_host", "space_ids.#"),
				),
			},
		},
	})
}

// TestAccResourceFleetServerHost_defaultOmitted verifies that omitting the
// Optional+Computed `default` attribute from config applies the schema
// default (false), rather than only ever seeing that value when explicitly
// configured.
func TestAccResourceFleetServerHost_defaultOmitted(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	hostName := sdkacctest.RandString(22)
	hostID := fmt.Sprintf("fleet-server-host-%s", sdkacctest.RandString(12))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id": config.StringVariable(hostID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", hostName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
				),
			},
		},
	})
}

// TestAccResourceFleetServerHost_invalidHostID exercises the negative path of
// the custom host_id validator (fleet.IDValidator): an explicit value
// containing a path separator and a traversal sequence must fail at plan
// time rather than reaching the API.
func TestAccResourceFleetServerHost_invalidHostID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", sdkacctest.RandString(22))),
					"host_id": config.StringVariable("../etc/passwd"),
				},
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)must not contain`),
			},
		},
	})
}

// TestAccResourceFleetServerHost_hostIDReplace verifies that changing an
// explicit host_id between two valid values triggers RequiresReplace
// (destroy-before-create), rather than an in-place update.
func TestAccResourceFleetServerHost_hostIDReplace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	hostName := sdkacctest.RandString(22)
	firstHostID := fmt.Sprintf("fleet-server-host-%s", sdkacctest.RandString(12))
	secondHostID := fmt.Sprintf("fleet-server-host-%s", sdkacctest.RandString(12))
	resourceName := "elasticstack_fleet_server_host.test_host"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id": config.StringVariable(firstHostID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", firstHostID),
					resource.TestCheckResourceAttr(resourceName, "host_id", firstHostID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id": config.StringVariable(secondHostID),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", secondHostID),
					resource.TestCheckResourceAttr(resourceName, "host_id", secondHostID),
				),
			},
		},
	})
}

// TestAccResourceFleetServerHost_spaceIDsUpdate exercises an in-place update
// that changes space_ids from a single-element set to a different,
// two-element set that does not include the prior operational space
// ("default"). space_ids has no RequiresReplace modifier, so this must
// succeed as an update, and update.go must resolve the operational space
// from prior state, not from the new plan set.
func TestAccResourceFleetServerHost_spaceIDsUpdate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSpaces, versionutils.FlavorAny)

	hostName := sdkacctest.RandString(22)
	hostID := fmt.Sprintf("fleet-server-host-%s", sdkacctest.RandString(12))
	random := sdkacctest.RandString(8)
	secondSpaceID := fmt.Sprintf("fleet-server-host-space-update-%s", random)
	thirdSpaceID := fmt.Sprintf("fleet-server-host-space-update-b-%s", random)
	resourceName := "elasticstack_fleet_server_host.test_host"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_space"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id": config.StringVariable(hostID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi_space"),
				ConfigVariables: config.Variables{
					"name":              config.StringVariable(fmt.Sprintf("FleetServerHost %s", hostName)),
					"host_id":           config.StringVariable(hostID),
					"second_space_id":   config.StringVariable(secondSpaceID),
					"third_space_id":    config.StringVariable(thirdSpaceID),
					"second_space_name": config.StringVariable(fmt.Sprintf("Fleet Server Host Space Update %s", random)),
					"third_space_name":  config.StringVariable(fmt.Sprintf("Fleet Server Host Space Update B %s", random)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", hostID),
					resource.TestCheckResourceAttr(resourceName, "space_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "space_ids.*", secondSpaceID),
					resource.TestCheckTypeSetElemAttr(resourceName, "space_ids.*", thirdSpaceID),
				),
			},
		},
	})
}

func checkResourceFleetServerHostDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_server_host" {
			continue
		}

		fleetClient := client.GetFleetClient()
		spaceID := rs.Primary.Attributes["space_ids.0"]
		host, diags := fleet.GetFleetServerHost(context.Background(), fleetClient, rs.Primary.ID, spaceID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if host != nil {
			return fmt.Errorf("fleet server host id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
