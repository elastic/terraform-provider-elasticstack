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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorScript(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorScript,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_script.test", "json", expectedJSONScript),
				),
			},
		},
	})
}

const expectedJSONScript = `{
	"script": {
		"description": "Extract 'tags' from 'env' field",
		"ignore_failure": false,
		"lang": "painless",
		"params": {
			"delimiter": "-",
			"position": 1
		},
		"source": "String[] envSplit = ctx['env'].splitOnToken(params['delimiter']);\nArrayList tags = new ArrayList();\ntags.add(envSplit[params['position']].trim());\nctx['tags'] = tags;\n"
	}
}`

const testAccDataSourceIngestProcessorScript = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_script" "test" {
  description = "Extract 'tags' from 'env' field"
  lang        = "painless"

  source = <<EOF
String[] envSplit = ctx['env'].splitOnToken(params['delimiter']);
ArrayList tags = new ArrayList();
tags.add(envSplit[params['position']].trim());
ctx['tags'] = tags;
EOF

  params = jsonencode({
    delimiter = "-"
    position  = 1
  })

}
`
