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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3680 reproduces the bug reported in GitHub issue #3680:
// when a kibana block's dynamic "feature" block produces zero entries (because
// the YAML-sourced feature key is absent/null for a given role), the provider
// raises a validation error instead of treating the empty feature set as "no
// Kibana feature privileges."
//
// The simplest reproduction is a kibana block with only `spaces` set and no
// `base` or `feature` — which is the state reached after the workaround
// `for_each = try(each.value.feature, [])` is applied to a role whose YAML
// entry has no feature key.
func TestAccReproduceIssue3680(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config: `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "issue_3680" {
  name = "issue_3680_no_feature_role"

  elasticsearch {
    cluster = ["monitor"]
  }

  kibana {
    spaces = ["default"]
    # No base and no feature blocks – simulates a role whose YAML entry has no
    # "feature" key, causing the dynamic feature block to produce zero entries.
  }
}
`,
				ExpectError: regexp.MustCompile(`Either one of the .feature. or .base. privileges must be set for kibana role`),
			},
		},
	})
}
