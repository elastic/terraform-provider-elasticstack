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

package role_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSecurityRole(t *testing.T) {
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("additional"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_security_role.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "name", "data_source_test"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_security_role.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					checks.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.names", []string{"index1", "index2"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.privileges.*", "all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.application", "myapp"),
					checks.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.privileges", []string{"admin", "read"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.resources.*", "*"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "metadata", `{"version":1}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_security_role.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "name", "data_source_test"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					checks.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.names", []string{"index1", "index2"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.privileges.*", "all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.application", "myapp"),
					checks.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.privileges", []string{"admin", "read"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.resources.*", "*"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "metadata", `{"version":1}`),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.*.clusters.*", "test-cluster2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.*.names.*", "sample2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.0.allow_restricted_indices", "true"),
					checks.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.0.privileges", []string{"create", "read", "write"}),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.0.field_security.0.grant.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.0.field_security.0.grant.*", "sample"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.0.field_security.0.except.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(security.MinSupportedDescriptionVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_security_role.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "name", "data_source_test"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					checks.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.names", []string{"index1", "index2"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.privileges.*", "all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.application", "myapp"),
					checks.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.privileges", []string{"admin", "read"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.resources.*", "*"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "metadata", `{"version":1}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "description", `Test data source`),
				),
			},
		},
	})
}

// TestAccDataSourceSecurityRoleNotFound exercises the data source's
// not-found path, where readDataSource nulls every field but the
// configured name.
func TestAccDataSourceSecurityRoleNotFound(t *testing.T) {
	roleName := fmt.Sprintf("nonexistent-%s", sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlphaNum))
	const ds = "data.elasticstack_elasticsearch_security_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "id", ""),
					resource.TestCheckResourceAttr(ds, "name", roleName),
					resource.TestCheckNoResourceAttr(ds, "description"),
					resource.TestCheckNoResourceAttr(ds, "global"),
					resource.TestCheckNoResourceAttr(ds, "metadata"),
					resource.TestCheckNoResourceAttr(ds, "cluster.#"),
					resource.TestCheckNoResourceAttr(ds, "run_as.#"),
					resource.TestCheckNoResourceAttr(ds, "applications.#"),
					resource.TestCheckNoResourceAttr(ds, "indices.#"),
					resource.TestCheckNoResourceAttr(ds, "remote_indices.#"),
				),
			},
		},
	})
}

// TestAccDataSourceSecurityRoleFieldSecurityAndQuery covers indices field
// level security (grant/except), indices/remote_indices query, and global
// privileges, none of which are exercised by TestAccDataSourceSecurityRole.
func TestAccDataSourceSecurityRoleFieldSecurityAndQuery(t *testing.T) {
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	const ds = "data.elasticstack_elasticsearch_security_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "name", roleName),
					resource.TestCheckResourceAttr(ds, "global", `{"application":{},"profile":{"write":{"applications":["*"]}}}`),
					resource.TestCheckResourceAttr(ds, "indices.0.query", `{"term":{"status":"active"}}`),
					resource.TestCheckResourceAttr(ds, "indices.0.field_security.0.grant.#", "2"),
					resource.TestCheckTypeSetElemAttr(ds, "indices.0.field_security.0.grant.*", "field1"),
					resource.TestCheckTypeSetElemAttr(ds, "indices.0.field_security.0.grant.*", "field2"),
					resource.TestCheckResourceAttr(ds, "indices.0.field_security.0.except.#", "1"),
					resource.TestCheckTypeSetElemAttr(ds, "indices.0.field_security.0.except.*", "field2.secret"),
					resource.TestCheckResourceAttr(ds, "remote_indices.0.query", `{"match_all":{}}`),
				),
			},
		},
	})
}

// TestAccDataSourceSecurityRoleMultiEntry covers roles with more than one
// indices/applications entry, which TestAccDataSourceSecurityRole never
// exercises (it only ever configures a single entry of each), and a
// remote_indices entry with allow_restricted_indices set to false.
func TestAccDataSourceSecurityRoleMultiEntry(t *testing.T) {
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	const ds = "data.elasticstack_elasticsearch_security_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "name", roleName),
					resource.TestCheckResourceAttr(ds, "indices.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(ds, "indices.*", map[string]string{
						"allow_restricted_indices": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(ds, "indices.*", map[string]string{
						"allow_restricted_indices": "false",
					}),
					resource.TestCheckTypeSetElemAttr(ds, "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr(ds, "indices.*.names.*", "index2"),
					resource.TestCheckResourceAttr(ds, "applications.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(ds, "applications.*", map[string]string{
						"application": "myapp",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(ds, "applications.*", map[string]string{
						"application": "otherapp",
					}),
					resource.TestCheckTypeSetElemAttr(ds, "applications.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr(ds, "applications.*.privileges.*", "admin"),
					resource.TestCheckResourceAttr(ds, "remote_indices.0.allow_restricted_indices", "false"),
				),
			},
		},
	})
}

// TestAccDataSourceSecurityRoleEmptyCollections covers a role with no
// indices/applications/remote_indices entries and empty cluster/run_as
// sets, asserting the data source nulls the collection attributes rather
// than leaving them populated from a prior read.
func TestAccDataSourceSecurityRoleEmptyCollections(t *testing.T) {
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	const ds = "data.elasticstack_elasticsearch_security_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ds, "id"),
					resource.TestCheckResourceAttr(ds, "name", roleName),
					resource.TestCheckResourceAttr(ds, "cluster.#", "0"),
					resource.TestCheckResourceAttr(ds, "run_as.#", "0"),
					resource.TestCheckResourceAttr(ds, "metadata", "{}"),
					resource.TestCheckNoResourceAttr(ds, "description"),
					resource.TestCheckNoResourceAttr(ds, "global"),
					resource.TestCheckNoResourceAttr(ds, "applications.#"),
					resource.TestCheckNoResourceAttr(ds, "indices.#"),
					resource.TestCheckNoResourceAttr(ds, "remote_indices.#"),
				),
			},
		},
	})
}
