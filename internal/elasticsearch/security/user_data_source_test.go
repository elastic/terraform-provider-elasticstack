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
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSecurityUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_security_user.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "username", "elastic"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_user.test", "roles.*", "superuser"),
				),
			},
		},
	})
}

func TestAccDataSourceSecurityUserCustom(t *testing.T) {
	enabledUsername := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	disabledUsername := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("enabled"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(enabledUsername),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_security_user.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "username", enabledUsername),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "full_name", "Test Custom User"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "email", "custom@example.com"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "roles.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_admin"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_user.test", "roles.*", "viewer"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "metadata", `{"env":"test"}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("disabled"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(disabledUsername),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_security_user.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "username", disabledUsername),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "full_name", "Disabled Test User"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "email", "disabled@example.com"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "enabled", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "roles.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_user.test", "roles.*", "viewer"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "metadata", "{}"),
				),
			},
		},
	})
}

func TestAccDataSourceSecurityUserNotFound(t *testing.T) {
	username := fmt.Sprintf("nonexistent-%s", sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlphaNum))
	const ds = "data.elasticstack_elasticsearch_security_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "id", ""),
					resource.TestCheckResourceAttr(ds, "username", username),
					resource.TestCheckNoResourceAttr(ds, "full_name"),
					resource.TestCheckNoResourceAttr(ds, "email"),
					resource.TestCheckNoResourceAttr(ds, "metadata"),
					resource.TestCheckNoResourceAttr(ds, "enabled"),
					resource.TestCheckNoResourceAttr(ds, "roles.#"),
				),
			},
		},
	})
}

func TestAccDataSourceSecurityUserOptionalNames(t *testing.T) {
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	const ds = "data.elasticstack_elasticsearch_security_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ds, "id"),
					resource.TestCheckResourceAttr(ds, "username", username),
					resource.TestCheckResourceAttr(ds, "full_name", ""),
					resource.TestCheckResourceAttr(ds, "email", ""),
					resource.TestCheckResourceAttr(ds, "roles.#", "1"),
					resource.TestCheckTypeSetElemAttr(ds, "roles.*", "viewer"),
					resource.TestCheckResourceAttr(ds, "enabled", "true"),
				),
			},
		},
	})
}
