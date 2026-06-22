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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3680 verifies the dynamic kibana/feature block pattern
// described in GitHub issue #3680. Roles sourced from YAML may or may not
// include Kibana feature privileges; provider should accept the configuration
// when roles without Kibana features simply omit the kibana block.
func TestAccReproduceIssue3680(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          acctest.NamedTestCaseDirectory("count"),
				ProtoV6ProviderFactories: acctest.Providers,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.this.0", "name", "role_with_feature"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.this.1", "name", "role_without_feature"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.this.0", "kibana.0.spaces.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.this.0", "kibana.0.feature.#", "1"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.this.1", "kibana.#"),
				),
			},
		},
	})
}
