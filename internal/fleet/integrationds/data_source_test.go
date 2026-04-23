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

package integrationds_test

import (
	"context"
	"errors"
	"fmt"
	"os"
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

var minVersionIntegrationDataSource = version.Must(version.NewVersion("8.6.0"))
var errFleetPackageNotFound = errors.New("fleet package not found")

const integrationDataSourceResourceName = "data.elasticstack_fleet_integration.test"

func TestAccDataSourceIntegration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationDataSourceResourceName, "id"),
					resource.TestCheckResourceAttr(integrationDataSourceResourceName, "name", "tcp"),
					checkFleetPackageVersion("tcp", false, ""),
				),
			},
		},
	})
}

func TestAccDataSourceIntegrationWithSpaceID(t *testing.T) {
	spaceID := strings.ToLower("test-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_name": config.StringVariable(spaceID),
					"space_id":   config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationDataSourceResourceName, "id"),
					resource.TestCheckResourceAttr(integrationDataSourceResourceName, "name", "tcp"),
					resource.TestCheckResourceAttr(integrationDataSourceResourceName, "space_id", spaceID),
					checkFleetPackageVersion("tcp", false, spaceID),
				),
			},
		},
	})
}

func TestAccDataSourceIntegrationAlternateName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationDataSourceResourceName, "id"),
					resource.TestCheckResourceAttr(integrationDataSourceResourceName, "name", "system"),
					checkFleetPackageVersion("system", false, ""),
				),
			},
		},
	})
}

func TestAccDataSourceIntegrationWithPrerelease(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationDataSourceResourceName, "id"),
					resource.TestCheckResourceAttr(integrationDataSourceResourceName, "name", "apm"),
					resource.TestCheckResourceAttr(integrationDataSourceResourceName, "prerelease", "true"),
					checkFleetPackageVersion("apm", true, ""),
					checkFleetPackageVersionChangesWhenPrereleaseEnabled("apm", ""),
				),
			},
		},
	})
}

func TestAccDataSourceIntegrationKibanaConnection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          testAccIntegrationKibanaConnectionVariables(),
				Check: resource.ComposeTestCheckFunc(
					append([]resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(integrationDataSourceResourceName, "id"),
						resource.TestCheckResourceAttr(integrationDataSourceResourceName, "name", "tcp"),
						checkFleetPackageVersion("tcp", false, ""),
						resource.TestCheckResourceAttr(integrationDataSourceResourceName, "kibana_connection.#", "1"),
						resource.TestCheckResourceAttr(integrationDataSourceResourceName, "kibana_connection.0.endpoints.#", "1"),
						resource.TestCheckResourceAttr(integrationDataSourceResourceName, "kibana_connection.0.endpoints.0", strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))),
						resource.TestCheckResourceAttr(integrationDataSourceResourceName, "kibana_connection.0.insecure", "false"),
					}, testAccIntegrationKibanaConnectionAuthChecks()...)...,
				),
			},
		},
	})
}

func checkFleetPackageVersion(packageName string, prerelease bool, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedVersion, err := fleetPackageVersion(packageName, prerelease, spaceID)
		if err != nil {
			return err
		}

		return resource.TestCheckResourceAttr(integrationDataSourceResourceName, "version", expectedVersion)(s)
	}
}

func checkFleetPackageVersionChangesWhenPrereleaseEnabled(packageName, spaceID string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		prereleaseVersion, err := fleetPackageVersion(packageName, true, spaceID)
		if err != nil {
			return err
		}

		gaVersion, err := fleetPackageVersion(packageName, false, spaceID)
		if err != nil {
			if errors.Is(err, errFleetPackageNotFound) {
				return nil
			}
			return err
		}

		if gaVersion == prereleaseVersion {
			return fmt.Errorf("expected package %q to resolve differently when prerelease is enabled, but both versions were %q", packageName, gaVersion)
		}

		return nil
	}
}

func fleetPackageVersion(packageName string, prerelease bool, spaceID string) (string, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return "", err
	}

	fleetClient, err := client.GetFleetClient()
	if err != nil {
		return "", err
	}

	packages, diags := fleet.GetPackages(context.Background(), fleetClient, prerelease, spaceID)
	if diags.HasError() {
		return "", diagutil.FwDiagsAsError(diags)
	}

	for _, pkg := range packages {
		if pkg.Name == packageName {
			return pkg.Version, nil
		}
	}

	return "", fmt.Errorf("%w: %q (prerelease=%t, space_id=%q)", errFleetPackageNotFound, packageName, prerelease, spaceID)
}

func testAccIntegrationKibanaConnectionVariables() config.Variables {
	apiKey := os.Getenv("KIBANA_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ELASTICSEARCH_API_KEY")
	}

	username := os.Getenv("KIBANA_USERNAME")
	if username == "" {
		username = os.Getenv("ELASTICSEARCH_USERNAME")
	}

	password := os.Getenv("KIBANA_PASSWORD")
	if password == "" {
		password = os.Getenv("ELASTICSEARCH_PASSWORD")
	}

	return config.Variables{
		"kibana_endpoints": config.ListVariable(config.StringVariable(strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT")))),
		"api_key":          config.StringVariable(apiKey),
		"username":         config.StringVariable(username),
		"password":         config.StringVariable(password),
	}
}

func testAccIntegrationKibanaConnectionAuthChecks() []resource.TestCheckFunc {
	apiKey := os.Getenv("KIBANA_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ELASTICSEARCH_API_KEY")
	}

	if apiKey != "" {
		return []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(integrationDataSourceResourceName, "kibana_connection.0.api_key", apiKey),
		}
	}

	username := os.Getenv("KIBANA_USERNAME")
	if username == "" {
		username = os.Getenv("ELASTICSEARCH_USERNAME")
	}

	password := os.Getenv("KIBANA_PASSWORD")
	if password == "" {
		password = os.Getenv("ELASTICSEARCH_PASSWORD")
	}

	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(integrationDataSourceResourceName, "kibana_connection.0.username", username),
		resource.TestCheckResourceAttr(integrationDataSourceResourceName, "kibana_connection.0.password", password),
	}
}
