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

package ingest_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3002 reproduces the drift reported in
// https://github.com/elastic/terraform-provider-elasticstack/issues/3002
// where a rename processor with override=true loses the override field after
// apply because the typed go-elasticsearch client (types.RenameProcessor) does
// not have an Override field and silently drops it on read.
func TestAccReproduceIssue3002(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test" {
  name = %q

  processors = [
    jsonencode({
      rename = {
        field        = "tmp_source_field"
        target_field = "destination_field"
        override     = true
        ignore_missing = true
      }
    })
  ]
}
`, pipelineName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// The typed go-elasticsearch client (types.RenameProcessor) lacks an
			// Override field. When Elasticsearch returns override=true, the field is
			// silently dropped during deserialization, so the post-apply refresh
			// produces a value that differs from the plan. Terraform rejects this with
			// "Provider produced inconsistent result after apply".
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   config,
				ExpectError:              regexp.MustCompile(`(?s)Provider produced inconsistent result after apply.*override`),
			},
		},
	})
}
