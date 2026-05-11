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

package proxy_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minFleetProxyVersion = version.Must(version.NewVersion("8.7.1"))

func TestAccResourceFleetProxy(t *testing.T) {
	proxyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetProxyDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Proxy %s", proxyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "name", fmt.Sprintf("Proxy %s", proxyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "url", "https://proxy.example.com:3128"),
					resource.TestMatchResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_id", regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr("elasticstack_fleet_proxy.test_proxy", "id", regexp.MustCompile(`^default/`)),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "is_preconfigured", "false"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.%"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Proxy Updated %s", proxyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "name", fmt.Sprintf("Proxy Updated %s", proxyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "url", "https://proxy-updated.example.com:3128"),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.X-Custom-Header", "my-value"),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.X-Another", "another-value"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("change_headers"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Proxy Updated %s", proxyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.X-Custom-Header"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.X-Another"),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.X-New-Header", "new-value"),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.X-Extra", "extra-value"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("clear_headers"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Proxy Updated %s", proxyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "url", "https://proxy-updated.example.com:3128"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_headers.%"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("clear_headers"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Proxy Updated %s", proxyName)),
				},
				ResourceName:      "elasticstack_fleet_proxy.test_proxy",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"kibana_connection",
				},
			},
		},
	})
}

// TestAccResourceFleetProxy_TLS exercises the TLS attribute lifecycle:
// create with all three cert fields populated, then clear them all in a
// follow-up apply. Asserting the cleared state catches any regression in
// `toAPIUpdateModel` that would silently leave certs untouched on the server
// when removed from config — Fleet's PUT semantics treat omitted fields as
// "set to null", so the update path must not unconditionally include cert
// values from prior state.
func TestAccResourceFleetProxy_TLS(t *testing.T) {
	proxyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetProxyDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_certs"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("TLS Proxy %s", proxyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "certificate", "PEM-CERT"),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "certificate_authorities", "PEM-CA"),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "certificate_key", "PEM-KEY"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("clear_certs"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("TLS Proxy %s", proxyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_fleet_proxy.test_proxy", "certificate"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_proxy.test_proxy", "certificate_authorities"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_proxy.test_proxy", "certificate_key"),
				),
			},
		},
	})
}

// TestAccResourceFleetProxy_ExplicitProxyID covers two requirements at once:
// (1) the user-supplied `proxy_id` is propagated into both the API request and
// the composite `id`, and (2) changing `proxy_id` triggers RequiresReplace so
// the underlying server-side resource is destroyed and recreated. The ID
// capture/compare helpers assert the replacement actually happened rather
// than just an in-place update on the same record.
func TestAccResourceFleetProxy_ExplicitProxyID(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	firstID := fmt.Sprintf("tf-acc-proxy-%s", suffix)
	secondID := fmt.Sprintf("tf-acc-proxy-%s-renamed", suffix)
	var capturedID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetProxyDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_proxy_id"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(fmt.Sprintf("Explicit ID Proxy %s", suffix)),
					"proxy_id": config.StringVariable(firstID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_id", firstID),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "id", fmt.Sprintf("default/%s", firstID)),
					testCheckFleetProxyCaptureID("elasticstack_fleet_proxy.test_proxy", &capturedID),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_proxy_id"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(fmt.Sprintf("Explicit ID Proxy %s", suffix)),
					"proxy_id": config.StringVariable(secondID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "proxy_id", secondID),
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "id", fmt.Sprintf("default/%s", secondID)),
					testCheckFleetProxyIDChanged("elasticstack_fleet_proxy.test_proxy", &capturedID),
				),
			},
		},
	})
}

// TestAccResourceFleetProxy_NonDefaultSpace verifies the resource correctly
// drives the space-aware path editor: the proxy is created and read inside a
// custom Kibana space rather than `default`. Without space-aware routing the
// create would land in `default` (succeeding) but the destroy check would
// then find a leftover record in the non-default space.
func TestAccResourceFleetProxy_NonDefaultSpace(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	spaceID := fmt.Sprintf("fleet-proxy-%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetProxyDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFleetProxyVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("non_default_space"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(fmt.Sprintf("Space Proxy %s", suffix)),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_proxy.test_proxy", "space_id", spaceID),
					resource.TestMatchResourceAttr("elasticstack_fleet_proxy.test_proxy", "id", regexp.MustCompile(fmt.Sprintf(`^%s/`, regexp.QuoteMeta(spaceID)))),
				),
			},
		},
	})
}

func testCheckFleetProxyCaptureID(resourceName string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		*target = res.Primary.ID
		return nil
	}
}

func testCheckFleetProxyIDChanged(resourceName string, previousID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		if *previousID == "" {
			return fmt.Errorf("previous ID was not captured")
		}
		if res.Primary.ID == *previousID {
			return fmt.Errorf("expected resource ID to change after proxy_id replacement, but remained %q", res.Primary.ID)
		}
		return nil
	}
}

func checkResourceFleetProxyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_proxy" {
			continue
		}

		fc, err := client.GetFleetClient()
		if err != nil {
			return err
		}

		proxyID := rs.Primary.Attributes["proxy_id"]
		spaceID := rs.Primary.Attributes["space_id"]

		proxy, diags := fleetclient.GetProxy(context.Background(), fc, spaceID, proxyID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if proxy != nil {
			return fmt.Errorf("fleet proxy id=%v still exists, but it should have been removed", proxyID)
		}
	}
	return nil
}
