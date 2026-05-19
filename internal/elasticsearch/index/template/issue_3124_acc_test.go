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

package template_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3124 reproduces the "Provider produced inconsistent result after apply"
// error for elasticstack_elasticsearch_index_template when template.settings contains
// index.search.slowlog.include.user.
//
// Root cause: go-elasticsearch v8 SlowlogSettings type does not have an Include field,
// so when the provider reads back the settings from Elasticsearch after apply,
// the include sub-object is silently dropped. The resulting state is missing
// index.search.slowlog.include.user, which does not match the planned state,
// triggering Terraform's post-apply consistency check error.
func TestAccReproduceIssue3124(t *testing.T) {
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccReproduceIssue3124Config(templateName),
				ExpectError:              regexp.MustCompile(`Provider produced inconsistent result after apply`),
			},
		},
	})
}

func testAccReproduceIssue3124Config(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = %q
  index_patterns = ["%s-*"]

  template {
    settings = jsonencode({
      index = {
        number_of_shards   = "1"
        number_of_replicas = "0"
        search = {
          slowlog = {
            include = {
              user = "true"
            }
          }
        }
      }
    })
  }
}
`, name, name)
}
