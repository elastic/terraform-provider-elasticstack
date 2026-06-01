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
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue864 reproduces the bug reported in
// https://github.com/elastic/terraform-provider-elasticstack/issues/864:
// modifying elasticstack_fleet_server_host.fleet_host when host_id is not
// explicitly set in the config returns 404.
//
// The root cause: host_id is Optional+Computed but lacks UseStateForUnknown().
// During an update, when host_id is absent from the config, the plan carries
// null (not the prior state value), so Update() calls the Fleet API with an
// empty ID → PUT /api/fleet/fleet_server_hosts/ → 404.
func TestAccReproduceIssue864(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionFleetServerHost, versionutils.FlavorAny)

	suffix := sdkacctest.RandString(12)
	name := fmt.Sprintf("fleet-864-%s", suffix)
	vars := config.Variables{
		"name": config.StringVariable(name),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetServerHostDestroy,
		Steps: []resource.TestStep{
			// Step 1: create a fleet server host without an explicit host_id.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_fleet_server_host.fleet_host", "host_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.fleet_host", "hosts.0", "https://fleet-server-issue-864-a.example:8220"),
				),
			},
			// Step 2: update hosts while still omitting host_id.
			// Without UseStateForUnknown() on host_id, the plan carries a null
			// host_id, Update() calls the Fleet API with an empty ID:
			//   PUT /api/fleet/fleet_server_hosts/  →  HTTP 404
			// The ExpectError confirms the bug is present.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile(`404`),
			},
		},
	})
}
