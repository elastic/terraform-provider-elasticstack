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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue4282 reproduces
// https://github.com/elastic/terraform-provider-elasticstack/issues/4282:
// when space_id is set on elasticstack_fleet_integration, writeIntegration
// correctly scopes the install call to that space, but the post-install
// status-polling call (waitForFleetIntegrationInstalled, invoked from
// internal/fleet/integration/create.go around line 80) hard-codes spaceID =
// "" and spaceAware = false, so the poll always queries the default-space
// endpoint (GET /api/fleet/epm/packages/{name}/{version} with no /s/{space}
// prefix). An API key/user scoped only to the target space has no read
// access to the default space and gets a 403 on every poll, causing resource
// creation to fail even though the install itself succeeded in the correct
// space.
//
// The test provisions:
//   - a custom Kibana space
//   - a Kibana role granting "fleet" feature privileges scoped only to that
//     space (no access to the default space)
//   - a user with that role
//
// and then applies elasticstack_fleet_integration with space_id set to the
// custom space, authenticating as the restricted user. If the bug is
// present, creation fails with an HTTP 403 surfaced through
// waitForFleetIntegrationInstalled's "failed to read package installation
// status" error wrapping.
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
				ExpectError: regexp.MustCompile(`(?s)failed to read package installation status.*HTTP 403|failed to install Fleet integration package`),
			},
		},
	})
}
