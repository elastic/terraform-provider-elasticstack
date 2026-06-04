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

package security_entity_store_entity_link_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionEntityStoreResolution = version.Must(version.NewVersion("9.4.0"))

func TestAccResourceSecurityEntityStoreEntityLink(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEntityStoreResolution, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "target_id", "generic:acc-test-target"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "id", "default/generic:acc-test-target"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity_link.test", "resolution_group_json"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "target_id", "generic:acc-test-target"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias3"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             "elasticstack_kibana_security_entity_store_entity_link.test",
				ImportState:              true,
				ImportStateVerify:        true,
			},
		},
	})
}
