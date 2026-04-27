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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.cluster.*", "create_snapshot"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.names.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "write"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "data_source_test2"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.clusters", []string{"test-cluster"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.field_security.0.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.privileges", []string{"create", "read", "write"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.cluster.*", "create_snapshot"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.names.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "write"),
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
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.names.*", "test-index"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "read"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_field_security"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "ds_test_idx_field_sec"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.cluster.*", "monitor"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.names.*", "sample-index"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.field_security.0.grant.*", "field1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.field_security.0.grant.*", "field2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*.field_security.0.except.*", "restricted"),
					resource.TestCheckTypeSetElemNestedAttrs("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.*", map[string]string{
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
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*.clusters.*", "test-cluster"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*.names.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*.privileges.*", "create"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*.privileges.*", "read"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*.privileges.*", "write"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*.field_security.0.grant.*", "sample"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*.field_security.0.except.*", "restricted"),
					resource.TestCheckTypeSetElemNestedAttrs("data.elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.*", map[string]string{
						"query": `{"match_all":{}}`,
					}),
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
