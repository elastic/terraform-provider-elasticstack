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
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var (
	minVersionFleetServerHost       = version.Must(version.NewVersion("8.6.0"))
	minVersionFleetServerHostSpaces = version.Must(version.NewVersion("9.1.0"))
	minVersionFleetServerHostProxy  = version.Must(version.NewVersion("8.7.1"))
	minVersionFleetServerHostSSL    = version.Must(version.NewVersion("9.1.0"))
	// minVersionFleetServerHostSSLDeferred is the first stack whose server host ssl schema accepts the
	// agent_* and es_* fields that the deferred-fields test seeds directly (9.1.10 and 9.2.8 reject them).
	minVersionFleetServerHostSSLDeferred = version.Must(version.NewVersion("9.3.0"))
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

func TestAccResourceFleetServerHost_ProxyID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostProxy, versionutils.FlavorAny)

	random := sdkacctest.RandString(8)
	vars := config.Variables{"suffix": config.StringVariable(random)}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_proxy"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test", "proxy_id", fmt.Sprintf("server-host-proxy-%s", random)),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_server_host.test", "proxy_id", "elasticstack_fleet_proxy.test", "proxy_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_proxy"),
				ConfigVariables:          vars,
				ResourceName:             "elasticstack_fleet_server_host.test",
				ImportState:              true,
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("other_proxy"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test", "proxy_id", fmt.Sprintf("server-host-other-proxy-%s", random)),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_server_host.test", "proxy_id", "elasticstack_fleet_proxy.other", "proxy_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_proxy"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_fleet_server_host.test", "proxy_id"),
					checkServerHostProxyClearedOnServer("elasticstack_fleet_server_host.test"),
				),
			},
		},
	})
}

func TestAccResourceFleetServerHost_emptyProxyID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_proxy"),
				ExpectError:              regexp.MustCompile(`(?s)proxy_id.*(at least 1|length)`),
			},
		},
	})
}

func checkServerHostProxyClearedOnServer(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %s not found in state", name)
		}
		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		host, diags := fleet.GetFleetServerHost(context.Background(), client.GetFleetClient(), rs.Primary.ID, rs.Primary.Attributes["space_ids.0"])
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if host == nil {
			return fmt.Errorf("fleet server host id=%s not found", rs.Primary.ID)
		}
		if host.ProxyId != nil && *host.ProxyId != "" {
			return fmt.Errorf("expected proxy_id to be cleared on the server, got %q", *host.ProxyId)
		}
		return nil
	}
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

const sslTestResourceName = "elasticstack_fleet_server_host.test"

func sslTestVariables(suffix string) config.Variables {
	return config.Variables{"suffix": config.StringVariable(suffix)}
}

func getServerHostFromState(s *terraform.State) (*kbapi.ServerHost, error) {
	rs, ok := s.RootModule().Resources[sslTestResourceName]
	if !ok {
		return nil, fmt.Errorf("resource %s not found in state", sslTestResourceName)
	}
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, err
	}
	host, diags := fleet.GetFleetServerHost(context.Background(), client.GetFleetClient(), rs.Primary.ID, rs.Primary.Attributes["space_ids.0"])
	if diags.HasError() {
		return nil, diagutil.FwDiagsAsError(diags)
	}
	if host == nil {
		return nil, fmt.Errorf("fleet server host id=%s not found", rs.Primary.ID)
	}
	return host, nil
}

func checkServerHostSSLOnServer(check func(host *kbapi.ServerHost) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		host, err := getServerHostFromState(s)
		if err != nil {
			return err
		}
		return check(host)
	}
}

func checkSSLClearedOnServer(host *kbapi.ServerHost) error {
	if ssl := host.Ssl; ssl != nil && (ssl.Certificate != nil || ssl.Key != nil || ssl.CertificateAuthorities != nil || ssl.ClientAuth != nil) {
		return fmt.Errorf("expected ssl to be cleared on the server, got %+v", *ssl)
	}
	return nil
}

func TestAccResourceFleetServerHost_SSL(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectSensitiveValue(sslTestResourceName, tfjsonpath.New("ssl").AtMapKey("key")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectSensitiveValue(sslTestResourceName, tfjsonpath.New("ssl").AtMapKey("key")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate", "PEM-CERT"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.key", "PEM-KEY"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.#", "2"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.0", "PEM-CA-1"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.1", "PEM-CA-2"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.client_auth", "required"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl_updated"),
				ConfigVariables:          sslTestVariables(suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(sslTestResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate", "PEM-CERT-2"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.key", "PEM-KEY-2"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.#", "3"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.2", "PEM-CA-5"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.client_auth", "optional"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl_shrunk"),
				ConfigVariables:          sslTestVariables(suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(sslTestResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.#", "1"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.0", "PEM-CA-3"),
					checkServerHostSSLOnServer(func(host *kbapi.ServerHost) error {
						if host.Ssl == nil || host.Ssl.CertificateAuthorities == nil || len(*host.Ssl.CertificateAuthorities) != 1 {
							return fmt.Errorf("expected a single certificate authority on the server, got %+v", host.Ssl)
						}
						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl_updated"),
				ConfigVariables:          sslTestVariables(suffix),
				ResourceName:             sslTestResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"space_ids"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[sslTestResourceName].Primary.Attributes["host_id"], nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("clear_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(sslTestResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl"),
					checkServerHostSSLOnServer(checkSSLClearedOnServer),
				),
			},
		},
	})
}

func TestAccResourceFleetServerHost_SSLChildRemoval(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)

	fullStep := resource.TestStep{
		ProtoV6ProviderFactories: acctest.Providers,
		ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
		ConfigVariables:          sslTestVariables(suffix),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate", "PEM-CERT"),
			resource.TestCheckResourceAttr(sslTestResourceName, "ssl.key", "PEM-KEY"),
			resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.#", "2"),
			resource.TestCheckResourceAttr(sslTestResourceName, "ssl.client_auth", "required"),
		),
	}

	dropStep := func(dir, attr string, serverCheck func(host *kbapi.ServerHost) error) resource.TestStep {
		return resource.TestStep{
			ProtoV6ProviderFactories: acctest.Providers,
			ConfigDirectory:          acctest.NamedTestCaseDirectory(dir),
			ConfigVariables:          sslTestVariables(suffix),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(sslTestResourceName, plancheck.ResourceActionUpdate),
				},
			},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl."+attr),
				checkServerHostSSLOnServer(serverCheck),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			fullStep,
			dropStep("drop_key", "key", func(host *kbapi.ServerHost) error {
				if host.Ssl == nil || host.Ssl.Key != nil {
					return fmt.Errorf("expected key to be cleared on the server, got %+v", host.Ssl)
				}
				return nil
			}),
			fullStep,
			dropStep("drop_cas", "certificate_authorities", func(host *kbapi.ServerHost) error {
				if host.Ssl == nil || host.Ssl.CertificateAuthorities != nil {
					return fmt.Errorf("expected certificate_authorities to be cleared on the server, got %+v", host.Ssl)
				}
				return nil
			}),
			fullStep,
			dropStep("drop_cert", "certificate", func(host *kbapi.ServerHost) error {
				if host.Ssl == nil || host.Ssl.Certificate != nil {
					return fmt.Errorf("expected certificate to be cleared on the server, got %+v", host.Ssl)
				}
				return nil
			}),
			fullStep,
			dropStep("drop_client_auth", "client_auth", func(host *kbapi.ServerHost) error {
				if host.Ssl == nil || host.Ssl.ClientAuth != nil {
					return fmt.Errorf("expected client_auth to be cleared on the server, got %+v", host.Ssl)
				}
				return nil
			}),
		},
	})
}

func TestAccResourceFleetServerHost_SSLValidation(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)

	rejected := func(dir string, pattern string) resource.TestStep {
		return resource.TestStep{
			ProtoV6ProviderFactories: acctest.Providers,
			ConfigDirectory:          acctest.NamedTestCaseDirectory(dir),
			ConfigVariables:          sslTestVariables(suffix),
			ExpectError:              regexp.MustCompile(pattern),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			rejected("ssl_empty_object", `(?s)Empty ssl configuration.*At least one of certificate, key, certificate_authorities, or client_auth`),
			rejected("ssl_empty_certificate", `(?is)certificate.*string length must be at least 1`),
			rejected("ssl_empty_key", `(?is)key.*string length must be at least 1`),
			rejected("ssl_empty_ca_list", `(?is)certificate_authorities.*must contain at least 1`),
			rejected("ssl_empty_ca_element", `(?is)certificate_authorities.*string length must be at least 1`),
			rejected("ssl_invalid_client_auth", `(?is)client_auth.*value must be one of`),
			rejected("ssl_ca_11", `(?is)certificate_authorities.*at most 10`),
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_ca_10"),
				ConfigVariables:          sslTestVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.#", "10"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate_authorities.9", "PEM-CA-10"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_certificate_only"),
				ConfigVariables:          sslTestVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate", "PEM-CERT"),
					resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl.key"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ssl_key_only"),
				ConfigVariables:          sslTestVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.key", "PEM-KEY"),
					resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl.certificate"),
				),
			},
		},
	})
}

func TestAccResourceFleetServerHost_SSLClientAuthOmitted(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_client_auth"),
				ConfigVariables:          sslTestVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate", "PEM-CERT"),
					resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl.client_auth"),
					checkServerHostSSLOnServer(func(host *kbapi.ServerHost) error {
						if host.Ssl == nil || host.Ssl.ClientAuth != nil {
							return fmt.Errorf("expected Kibana to report no client_auth default, got %+v", host.Ssl)
						}
						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_client_auth"),
				ConfigVariables:          sslTestVariables(suffix),
				PlanOnly:                 true,
			},
		},
	})
}

func seedServerHostSSL(t *testing.T, suffix, sslJSON string) func() {
	t.Helper()

	return func() {
		var body kbapi.PutFleetFleetServerHostsItemidJSONRequestBody
		if err := json.Unmarshal([]byte(`{"ssl":`+sslJSON+`}`), &body); err != nil {
			t.Fatal(err)
		}
		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			t.Fatal(err)
		}
		diags := fleet.PutFleetServerHost(context.Background(), client.GetFleetClient(), "server-host-ssl-"+suffix, "", body)
		if diags.HasError() {
			t.Fatal(diagutil.FwDiagsAsError(diags))
		}
	}
}

// TestAccResourceFleetServerHost_SSLUpgrade verifies an unchanged configuration without ssl plans
// no diff for a host with no server-side ssl, and that pre-existing server-side ssl surfaces as a
// planned removal.
func TestAccResourceFleetServerHost_SSLUpgrade(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				Check:                    resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				PlanOnly:                 true,
			},
			{
				PreConfig:                seedServerHostSSL(t, suffix, `{"certificate":"SEEDED-CERT","key":"SEEDED-KEY"}`),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl"),
					checkServerHostSSLOnServer(checkSSLClearedOnServer),
				),
			},
		},
	})
}

// TestAccResourceFleetServerHost_SSLDeferredFields verifies that an update that would replace ssl is
// rejected, without mutating the host, when the host stores agent_* or es_* ssl fields.
func TestAccResourceFleetServerHost_SSLDeferredFields(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSLDeferred, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)
	checkDeferredFieldsIntact := checkServerHostSSLOnServer(func(host *kbapi.ServerHost) error {
		if host.Ssl == nil || host.Ssl.AgentCertificate == nil || *host.Ssl.AgentCertificate != "SEEDED-AGENT-CERT" {
			return fmt.Errorf("expected agent_certificate to be preserved, got %+v", host.Ssl)
		}
		if host.Ssl.EsKey == nil || *host.Ssl.EsKey != "SEEDED-ES-KEY" {
			return fmt.Errorf("expected es_key to be preserved, got %+v", host.Ssl)
		}
		if host.Ssl.Certificate == nil || *host.Ssl.Certificate != "PEM-CERT" {
			return fmt.Errorf("expected certificate to be unchanged, got %+v", host.Ssl)
		}
		return nil
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
			},
			{
				PreConfig:                seedServerHostSSL(t, suffix, `{"certificate":"PEM-CERT","agent_certificate":"SEEDED-AGENT-CERT","es_key":"SEEDED-ES-KEY"}`),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl_changed"),
				ConfigVariables:          sslTestVariables(suffix),
				ExpectError:              regexp.MustCompile(`(?s)Cannot update ssl without erasing unmanaged fields.*agent_certificate, es_key`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				ExpectError:              regexp.MustCompile(`(?s)Cannot update ssl without erasing unmanaged fields.*agent_certificate, es_key`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl_renamed"),
				ConfigVariables:          sslTestVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "name", "SSL Server Host Renamed "+suffix),
					checkDeferredFieldsIntact,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				Check:                    checkDeferredFieldsIntact,
			},
		},
	})
}

func TestAccResourceFleetServerHost_SSLRejectedBelowMinimumVersion(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	unsupported, err := versionutils.CheckIfVersionIsUnsupported(minVersionFleetServerHostSSL)()
	if err != nil {
		t.Fatalf("failed to check stack version: %v", err)
	}
	if !unsupported {
		t.Skip("stack supports ssl; skipping version gate test")
	}

	suffix := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				ExpectError:              regexp.MustCompile(`(?s)ssl is only supported in Elastic Stack 9\.1\.0 and above`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				Check:                    resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl"),
			},
		},
	})
}

func TestAccResourceFleetServerHost_SSLAddToExisting(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				Check:                    resource.TestCheckNoResourceAttr(sslTestResourceName, "ssl"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_ssl"),
				ConfigVariables:          sslTestVariables(suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(sslTestResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.certificate", "PEM-CERT"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.key", "PEM-KEY"),
					resource.TestCheckResourceAttr(sslTestResourceName, "ssl.client_auth", "required"),
				),
			},
		},
	})
}

func TestAccResourceFleetServerHost_SSLClientAuth(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(8)

	step := func(clientAuth string, planOnly bool) resource.TestStep {
		step := resource.TestStep{
			ProtoV6ProviderFactories: acctest.Providers,
			ConfigDirectory:          acctest.NamedTestCaseDirectory("client_auth"),
			ConfigVariables: config.Variables{
				"suffix":      config.StringVariable(suffix),
				"client_auth": config.StringVariable(clientAuth),
			},
		}
		if planOnly {
			step.PlanOnly = true
			return step
		}
		step.Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(sslTestResourceName, "ssl.client_auth", clientAuth),
			checkServerHostSSLOnServer(func(host *kbapi.ServerHost) error {
				if host.Ssl == nil || host.Ssl.ClientAuth == nil || string(*host.Ssl.ClientAuth) != clientAuth {
					return fmt.Errorf("expected client_auth %q on the server, got %+v", clientAuth, host.Ssl)
				}
				return nil
			}),
		)
		return step
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			step("none", false),
			step("none", true),
			step("required", false),
			step("none", false),
			step("optional", false),
			step("optional", true),
		},
	})
}
