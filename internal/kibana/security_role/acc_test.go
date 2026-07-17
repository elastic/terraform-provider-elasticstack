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

package security_role_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceKibanaSecurityRole(t *testing.T) {
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
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.run_as.#"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.cluster.*", "create_snapshot"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "write"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.0.names", []string{"sample"}),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.field_security.grant.*", "sample"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.field_security.except.*", "restricted"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_kibana_security_role.test", "elasticsearch.indices.*", map[string]string{
						"query": `{"match_all":{}}`,
					}),
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
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.0.field_security.%"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.0.query"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.run_as", []string{"elastic", "kibana"}),
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
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.0.field_security.%"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "description", "Role description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "metadata", `{"acc_resource":"meta_v1"}`),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.cluster.*", "create_snapshot"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "write"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedDescriptionVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unset_description"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "description"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.cluster.*", "create_snapshot"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
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
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.run_as.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.0.field_security.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.2.privileges", []string{"minimal_read", "store_search_session", "url_create"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.clusters", []string{"test-cluster"}),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.field_security.grant.*", "sample"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.field_security.except.*", "restricted"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*", map[string]string{
						"query": `{"match_all":{}}`,
					}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.privileges", []string{"create", "read", "write"}),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.allow_restricted_indices", "true"),
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
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.indices.0.field_security.%"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.query"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.clusters", []string{"test-cluster2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.field_security.grant", []string{"sample2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.names", []string{"sample2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.privileges", []string{"create", "read", "write"}),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.allow_restricted_indices", "false"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccKibanaSecurityRoleResourceFromSDK/create/main.tf
var kibanaSecurityRoleFromSDKCreateConfig string

func TestAccResourceKibanaSecurityRoleDynamicFeature(t *testing.T) {
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityRoleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.dynamic_feature", "name", roleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.dynamic_feature", "kibana.0.feature.#", "1"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.dynamic_feature", "kibana.0.base.#"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_kibana_security_role.dynamic_feature", "kibana.0.feature.*", map[string]string{
						"name": "discover",
					}),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_role.dynamic_feature", "kibana.0.feature.*.privileges.*", "read"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.dynamic_feature", "kibana.0.spaces", []string{"*"}),
				),
			},
		},
	})
}

func TestAccResourceKibanaSecurityRoleDynamicFeatureEmptyForEach(t *testing.T) {
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"role_name": config.StringVariable(roleName)},
				ExpectError: regexp.MustCompile(
					`Either one of the ` + "`feature`" + ` or ` + "`base`" + ` privileges must be set for kibana role!`,
				),
			},
		},
	})
}

func TestAccKibanaSecurityRoleResourceFromSDK(t *testing.T) {
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityRoleDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.15.1",
					},
				},
				Config: kibanaSecurityRoleFromSDKCreateConfig,
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.upgrade", "name", roleName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

var checkResourceSecurityRoleDestroy = checks.KibanaResourceDestroyCheck(
	"elasticstack_kibana_security_role",
	func(ctx context.Context, client *kibanaoapi.Client, id string) (bool, error) {
		role, diags := kibanaoapi.GetSecurityRole(ctx, client, id)
		if diags.HasError() {
			return false, fmt.Errorf("failed to get security role %s: %v", id, diags)
		}
		return role != nil, nil
	},
)

func TestAccDataSourceKibanaSecurityRole(t *testing.T) {
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	minSupportedDescriptionVersion := version.Must(version.NewVersion("8.15.0"))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "data_source_test"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.0.field_security.%"),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.cluster.*", "create_snapshot"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.names.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "write"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "data_source_test2"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.0.field_security.%"),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.clusters", []string{"test-cluster"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.field_security.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.privileges", []string{"create", "read", "write"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.cluster.*", "create_snapshot"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.names.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "write"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("feature_privileges"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "ds_test_feature_privs"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "kibana.0.feature.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.elasticstack_kibana_security_role.test", "kibana.0.feature.*", map[string]string{
						"name": "actions",
					}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "kibana.0.feature.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.names.*", "test-index"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "read"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_field_security"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "ds_test_idx_field_sec"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.cluster.*", "monitor"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.names.*", "sample-index"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.field_security.grant.*", "field1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.field_security.grant.*", "field2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*.field_security.except.*", "restricted"),
					resource.TestCheckTypeSetElemNestedAttrs("data.elasticstack_kibana_security_role.test", "elasticsearch.indices.*", map[string]string{
						"query": `{"match_all":{}}`,
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remote_indices_extended"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "ds_test_remote_idx_ext"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.clusters.*", "test-cluster"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.names.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.privileges.*", "write"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.field_security.grant.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*.field_security.except.*", "restricted"),
					resource.TestCheckTypeSetElemNestedAttrs("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.*", map[string]string{
						"query": `{"match_all":{}}`,
					}),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.remote_indices.0.allow_restricted_indices", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedDescriptionVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("description_metadata"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "ds_test_desc_metadata"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "description", "Test role description"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_role.test", "metadata"),
				),
			},
		},
	})
}
