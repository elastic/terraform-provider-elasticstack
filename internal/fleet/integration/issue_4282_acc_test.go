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
		PreCheck: func() { acctest.PreCheck(t) },
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
					testAccCheckIntegrationInstalledInSpace("tcp", "1.16.0", spaceID),
					testAccCheckFleetGetPackageDefaultSpaceForbidden(username, password),
				),
			},
		},
	})
}

// testAccCheckFleetGetPackageDefaultSpaceForbidden verifies that credentials
// scoped only to a custom space cannot read packages from the default space.
func testAccCheckFleetGetPackageDefaultSpaceForbidden(username, password string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		endpoint := strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))
		if endpoint == "" {
			return fmt.Errorf("KIBANA_ENDPOINT is not set")
		}

		fleetClient, err := fleet.NewClient(fleet.Config{
			URL:      endpoint,
			Username: username,
			Password: password,
		})
		if err != nil {
			return fmt.Errorf("failed to create Fleet client: %w", err)
		}

		_, diags := fleet.GetPackage(context.Background(), fleetClient, "tcp", "1.16.0", "")
		if !diags.HasError() {
			return fmt.Errorf("expected default-space GetPackage to fail with restricted credentials, but succeeded")
		}

		for _, d := range diags {
			summary := strings.ToLower(d.Summary())
			detail := strings.ToLower(d.Detail())
			if strings.Contains(summary, "http 403") || strings.Contains(detail, "forbidden") {
				return nil
			}
		}

		return fmt.Errorf("expected HTTP 403/forbidden for default-space GetPackage, got: %v", diags)
	}
}
