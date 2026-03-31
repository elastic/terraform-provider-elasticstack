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

package defaultdataview_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))

func TestAccResourceDefaultDataView(t *testing.T) {
	indexName1 := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	indexName2 := "my-other-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("unset"),
				ConfigVariables: config.Variables{
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewWithSkipDelete(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("skip_delete"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewWithCustomSpace(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	spaceID := "test-space-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("custom_space"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
					"space_id":   config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", spaceID),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", spaceID),
				),
			},
		},
	})
}
