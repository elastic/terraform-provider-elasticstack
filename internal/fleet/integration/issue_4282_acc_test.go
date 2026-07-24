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
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccReproduceIssue4282 is a regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/4282.
// When space_id is set on elasticstack_fleet_integration, the post-install
// status poll must use the same space context as the install call. A caller
// scoped only to the target space must not hit the default-space get-package
// endpoint during the wait.
//
// The test provisions:
//   - a custom Kibana space
//   - a Kibana role granting Fleet feature privileges scoped only to that
//     space (no access to the default space)
//   - a user with that role
//
// and then applies elasticstack_fleet_integration with space_id set to the
// custom space, authenticating as the restricted user.
func TestAccReproduceIssue4282(t *testing.T) {
	versionutils.SkipIfUnsupported(t, integration.MinVersionSpaceAwareIntegration, versionutils.FlavorAny)

	spaceID := "issue4282-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	username := "issue4282-user-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	roleName := "issue4282-role-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	password := "Password123!"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("setup"),
				ConfigVariables: config.Variables{
					"space_id":  config.StringVariable(spaceID),
					"username":  config.StringVariable(username),
					"password":  config.StringVariable(password),
					"role_name": config.StringVariable(roleName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("install"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"space_id":  config.StringVariable(spaceID),
					"username":  config.StringVariable(username),
					"password":  config.StringVariable(password),
					"role_name": config.StringVariable(roleName),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "space_id", spaceID),
					testAccCheckIntegrationInstalledInSpace(spaceID),
					testAccCheckFleetGetPackageTargetSpaceAllowed(username, password, spaceID),
					testAccCheckFleetGetPackageDefaultSpaceForbidden(username, password),
				),
			},
		},
	})
}

func testAccFleetClientForUser(username, password string) (*fleet.Client, error) {
	endpoint := strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))
	if endpoint == "" {
		return nil, fmt.Errorf("KIBANA_ENDPOINT is not set")
	}

	return fleet.NewClient(fleet.Config{
		URL:      endpoint,
		Username: username,
		Password: password,
	})
}

func diagnosticContainsHTTP403(d interface {
	Summary() string
	Detail() string
}) bool {
	summary := strings.ToLower(d.Summary())
	detail := strings.ToLower(d.Detail())
	return strings.Contains(summary, "http 403") || strings.Contains(detail, "http 403")
}

// testAccCheckFleetGetPackageTargetSpaceAllowed verifies restricted credentials
// can read the package from the configured target space and that Fleet reports
// it installed there.
func testAccCheckFleetGetPackageTargetSpaceAllowed(username, password, spaceID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		fleetClient, err := testAccFleetClientForUser(username, password)
		if err != nil {
			return err
		}

		pkg, diags := fleet.GetPackage(context.Background(), fleetClient, testAccTCPIntegrationName, testAccTCPIntegrationVersion, spaceID)
		if diags.HasError() {
			return fmt.Errorf("expected target-space GetPackage to succeed with restricted credentials, got: %v", diags)
		}

		return testAccAssertPackageInstalledInSpace(pkg, testAccTCPIntegrationName, testAccTCPIntegrationVersion, spaceID)
	}
}

// testAccCheckFleetGetPackageDefaultSpaceForbidden verifies that credentials
// scoped only to a custom space cannot read packages from the default space.
func testAccCheckFleetGetPackageDefaultSpaceForbidden(username, password string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		fleetClient, err := testAccFleetClientForUser(username, password)
		if err != nil {
			return err
		}

		_, diags := fleet.GetPackage(context.Background(), fleetClient, testAccTCPIntegrationName, testAccTCPIntegrationVersion, "")
		if !diags.HasError() {
			return fmt.Errorf("expected default-space GetPackage to fail with restricted credentials, but succeeded")
		}

		for _, d := range diags {
			if diagnosticContainsHTTP403(d) {
				return nil
			}
		}

		return fmt.Errorf("expected HTTP 403 for default-space GetPackage, got: %v", diags)
	}
}
