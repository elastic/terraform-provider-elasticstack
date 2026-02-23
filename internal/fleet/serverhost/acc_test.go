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
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionFleetServerHost = version.Must(version.NewVersion("8.6.0"))

func TestAccResourceFleetServerHostFromSDK(t *testing.T) {
	policyName := sdkacctest.RandString(22)

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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetServerHost),
				Config:   testAccResourceFleetServerHostCreate(policyName),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", "fleet-server-host-id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetServerHost),
				Config:                   testAccResourceFleetServerHostCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", "fleet-server-host-id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
		},
	})
}

func TestAccResourceFleetServerHost(t *testing.T) {
	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceFleetServerHostDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetServerHost),
				Config:   testAccResourceFleetServerHostCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", "fleet-server-host-id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetServerHost),
				Config:   testAccResourceFleetServerHostUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("Updated FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "id", "fleet-server-host-id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minVersionFleetServerHost),
				Config:            testAccResourceFleetServerHostUpdate(policyName),
				ResourceName:      "elasticstack_fleet_server_host.test_host",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceFleetServerHostCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_server_host" "test_host" {
  name               = "%s"
  host_id            = "fleet-server-host-id"
  default            =  false
  hosts              = [
    "https://fleet-server:8220"
  ]
}
`, fmt.Sprintf("FleetServerHost %s", id))
}

func testAccResourceFleetServerHostUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_server_host" "test_host" {
  name               = "%s"
  host_id            = "fleet-server-host-id"
  default            =  false
  hosts              = [
    "https://fleet-server:8220"
  ]
}
`, fmt.Sprintf("Updated FleetServerHost %s", id))
}

func checkResourceFleetServerHostDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
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
		host, diags := fleet.GetFleetServerHost(context.Background(), fleetClient, rs.Primary.ID, "")
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if host != nil {
			return fmt.Errorf("fleet server host id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
