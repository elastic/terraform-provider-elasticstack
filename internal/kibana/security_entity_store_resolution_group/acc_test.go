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

package security_entity_store_resolution_group_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

var minVersionEntityStoreResolution = version.Must(version.NewVersion("9.4.0"))

func TestAccDataSourceSecurityEntityStoreResolutionGroup(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEntityStoreResolution, versionutils.FlavorAny)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_resolution_group.test", "entity_id", "generic:acc-test-target"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_resolution_group.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_resolution_group.test", "id", spaceID+"/generic:acc-test-target"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_resolution_group.test", "resolution_group_json"),
					resource.TestMatchResourceAttr("data.elasticstack_kibana_security_entity_store_resolution_group.test", "resolution_group_json", regexp.MustCompile(`generic:acc-test-target`)),
				),
			},
		},
	})
}
