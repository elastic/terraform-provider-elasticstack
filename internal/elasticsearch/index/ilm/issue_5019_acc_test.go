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

package ilm_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue5019 reproduces https://github.com/elastic/terraform-provider-elasticstack/issues/5019.
// Explicitly configuring warm.allocate.include/exclude as jsonencode({}) plans "{}" but
// flatten reads the empty object back as null, so Terraform reports an inconsistent result after apply.
func TestAccReproduceIssue5019(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config: `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "issue_5019" {
  name = "` + policyName + `"

  hot {
    min_age = "0ms"
    rollover {
      max_age = "90d"
    }
  }

  warm {
    min_age = "0ms"
    allocate {
      number_of_replicas    = 1
      total_shards_per_node = -1
      include               = jsonencode({})
      exclude               = jsonencode({})
    }
  }
}
`,
				ExpectError: regexp.MustCompile(`(?s)inconsistent result after apply`),
			},
		},
	})
}
