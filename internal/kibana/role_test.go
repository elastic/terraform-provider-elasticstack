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

package kibana_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceKibanaSecurityRole(t *testing.T) {
	// generate a random role name
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	roleNameRemoteIndices := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	minSupportedDescriptionVersion := version.Must(version.NewVersion("8.15.0"))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityRoleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.base.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.0.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.2.privileges", []string{"minimal_read", "store_search_session", "url_create"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedDescriptionVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_description"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "description", "Role description"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remote_indices_create"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleNameRemoteIndices)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.base.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.0.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.2.privileges", []string{"minimal_read", "store_search_session", "url_create"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.clusters", []string{"test-cluster"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.field_security.0.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.privileges", []string{"create", "read", "write"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remote_indices_update"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleNameRemoteIndices)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.clusters", []string{"test-cluster2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.field_security.0.grant", []string{"sample2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.names", []string{"sample2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.privileges", []string{"create", "read", "write"}),
				),
			},
		},
	})
}

func checkResourceSecurityRoleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_security_role" {
			continue
		}
		compID := rs.Primary.ID

		kibanaClient, err := client.GetKibanaClient()
		if err != nil {
			return err
		}
		res, err := kibanaClient.KibanaRoleManagement.Get(compID)
		if err != nil || res != nil {
			return fmt.Errorf("Role (%s) still exists", compID)
		}
	}
	return nil
}
