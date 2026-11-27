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

package serverhostds_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
	minVersionFleetServerHostProxy  = version.Must(version.NewVersion("8.7.1"))
	minVersionFleetServerHostSpaces = version.Must(version.NewVersion("9.1.0"))
	minVersionFleetServerHostSSL    = version.Must(version.NewVersion("9.1.0"))
)

const (
	dataSourceName = "data.elasticstack_fleet_server_host.test"
	resourceName   = "elasticstack_fleet_server_host.test"
)

func randomSuffix() string {
	return sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
}

func suffixVariables(suffix string) config.Variables {
	return config.Variables{"suffix": config.StringVariable(suffix)}
}

func TestAccDataSourceFleetServerHost(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	suffix := randomSuffix()
	name := "Data Source Server Host " + suffix

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_id"),
				ConfigVariables:          suffixVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "host_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, "host_id"),
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "default", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "hosts.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "hosts.0", "https://fleet-server-1:8220"),
					resource.TestCheckResourceAttr(dataSourceName, "hosts.1", "https://fleet-server-2:8220"),
					resource.TestCheckNoResourceAttr(dataSourceName, "proxy_id"),
					resource.TestCheckNoResourceAttr(dataSourceName, "ssl.%"),
					resource.TestCheckNoResourceAttr(dataSourceName, "space_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_name_updated"),
				ConfigVariables:          suffixVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "host_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, "host_id"),
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "default", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "hosts.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "hosts.0", "https://fleet-server-3:8220"),
				),
			},
		},
	})
}

func TestAccDataSourceFleetServerHostProxyID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostProxy, versionutils.FlavorAny)

	suffix := randomSuffix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          suffixVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, "host_id"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_id", "ds-server-host-proxy-"+suffix),
				),
			},
		},
	})
}

func TestAccDataSourceFleetServerHostSSL(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSSL, versionutils.FlavorAny)

	suffix := randomSuffix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          suffixVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, "host_id"),
					resource.TestCheckResourceAttr(dataSourceName, "ssl.%", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "ssl.certificate", "PEM-CERT"),
					resource.TestCheckResourceAttr(dataSourceName, "ssl.certificate_authorities.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "ssl.certificate_authorities.0", "PEM-CA-1"),
					resource.TestCheckResourceAttr(dataSourceName, "ssl.certificate_authorities.1", "PEM-CA-2"),
					resource.TestCheckResourceAttr(dataSourceName, "ssl.client_auth", "required"),
					resource.TestCheckNoResourceAttr(dataSourceName, "ssl.key"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("certificate_only"),
				ConfigVariables:          suffixVariables(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ssl.certificate", "PEM-CERT"),
					resource.TestCheckNoResourceAttr(dataSourceName, "ssl.certificate_authorities.#"),
					resource.TestCheckNoResourceAttr(dataSourceName, "ssl.client_auth"),
					resource.TestCheckNoResourceAttr(dataSourceName, "ssl.key"),
				),
			},
		},
	})
}

func TestAccDataSourceFleetServerHostSpaceID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHostSpaces, versionutils.FlavorAny)

	suffix := randomSuffix()
	spaceID := "test-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(suffix),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.elasticstack_fleet_server_host.omitted", "space_id"),
					resource.TestCheckResourceAttrPair("data.elasticstack_fleet_server_host.omitted", "host_id", resourceName, "host_id"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_server_host.same_space", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("data.elasticstack_fleet_server_host.same_space", "host_id", resourceName, "host_id"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_server_host.default_space", "space_id", "default"),
					resource.TestCheckResourceAttrPair("data.elasticstack_fleet_server_host.default_space", "host_id", resourceName, "host_id"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_server_host.default_space", "name", "Data Source Space Server Host "+suffix),
				),
			},
		},
	})
}

func TestAccDataSourceFleetServerHostAmbiguousName(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	suffix := randomSuffix()
	hostID := `host_id\s+"ds-duplicate-[ab]-` + suffix + `"`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkFleetServerHostDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          suffixVariables(suffix),
				ExpectError: regexp.MustCompile(`(?s)multiple Fleet server hosts match name.*` +
					hostID + `,\s+` + hostID),
			},
		},
	})
}

func TestAccDataSourceFleetServerHostNotFound(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	suffix := randomSuffix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_id"),
				ConfigVariables:          suffixVariables(suffix),
				ExpectError:              regexp.MustCompile(`(?s)fleet_server_host "does-not-exist-` + suffix + `" was\s+not\s+found`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_name"),
				ConfigVariables:          suffixVariables(suffix),
				ExpectError:              regexp.MustCompile(`(?s)no Fleet server host found matching name\s+"does-not-exist-` + suffix + `"`),
			},
		},
	})
}

func TestAccDataSourceFleetServerHostSelectorValidation(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	steps := []struct {
		directory string
		want      string
	}{
		{"no_selector", "No attribute specified when one (and only one) of [host_id,name,default] is required"},
		{"host_id_and_name", "2 attributes specified when one (and only one) of [host_id,name,default] is required"},
		{"name_and_default", "2 attributes specified when one (and only one) of [host_id,name,default] is required"},
		{"default_false", `Attribute default Value must be "true", got: false`},
		{"empty_name", "Attribute name string length must be at least 1, got: 0"},
	}

	testSteps := make([]resource.TestStep, len(steps))
	for i, step := range steps {
		// Terraform wraps long diagnostics, so any space may be a line break.
		pattern := strings.ReplaceAll(regexp.QuoteMeta(step.want), " ", `\s+`)
		testSteps[i] = resource.TestStep{
			ProtoV6ProviderFactories: acctest.Providers,
			ConfigDirectory:          acctest.NamedTestCaseDirectory(step.directory),
			PlanOnly:                 true,
			ExpectError:              regexp.MustCompile(pattern),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps:    testSteps,
	})
}

func checkFleetServerHostDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_server_host" {
			continue
		}

		host, diags := fleet.GetFleetServerHost(context.Background(), client.GetFleetClient(), rs.Primary.ID, "")
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if host != nil {
			return fmt.Errorf("fleet server host id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
