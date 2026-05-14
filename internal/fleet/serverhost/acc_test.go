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
					resource.TestCheckResourceAttrSet("elasticstack_fleet_server_host.test_computed_id", "host_id"),
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

func checkResourceFleetServerHostDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_server_host" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
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
