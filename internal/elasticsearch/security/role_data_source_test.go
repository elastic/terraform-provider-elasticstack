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

package security_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSecurityRole(t *testing.T) {
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("additional"),
				Check: resource.ComposeTestCheckFunc(
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
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("read"),
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Check: resource.ComposeTestCheckFunc(
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
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("all_attributes"),
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(security.MinSupportedDescriptionVersion),
				Check: resource.ComposeTestCheckFunc(
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
