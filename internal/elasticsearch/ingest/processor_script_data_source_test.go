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

const testAccDataSourceIngestProcessorScriptResourceName = "data.elasticstack_elasticsearch_ingest_processor_script.test"

func TestAccDataSourceIngestProcessorScript(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccDataSourceIngestProcessorScriptResourceName, "id"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "description", "Extract 'tags' from 'env' field"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "lang", "painless"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "source", expectedScriptSource),
					CheckResourceJSON(testAccDataSourceIngestProcessorScriptResourceName, "params", `{"delimiter":"-","position":1}`),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "script_id"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "if"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "on_failure.#"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "tag"),
					CheckResourceJSON(testAccDataSourceIngestProcessorScriptResourceName, "json", expectedJSONScript),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccDataSourceIngestProcessorScriptResourceName, "id"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "description", "Annotate tags when env is present"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "if", "ctx.env != null"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "ignore_failure", "true"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "tag", "script-tag"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "lang", "expression"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "source", expectedScriptSourceAllAttributes),
					CheckResourceJSON(testAccDataSourceIngestProcessorScriptResourceName, "params", `{"count":2,"prefix":"prod"}`),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "on_failure.#", "1"),
					CheckResourceJSON(testAccDataSourceIngestProcessorScriptResourceName, "on_failure.0", `{"set":{"field":"error.message","value":"script processor failed"}}`),
					CheckResourceJSON(testAccDataSourceIngestProcessorScriptResourceName, "json", expectedJSONScriptAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("stored_script"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccDataSourceIngestProcessorScriptResourceName, "id"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "description", "Run stored script to derive tags"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "lang", "painless"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "script_id", "stored-script-derive-tags"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "source"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "params"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "if"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "on_failure.#"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "tag"),
					CheckResourceJSON(testAccDataSourceIngestProcessorScriptResourceName, "json", expectedJSONScriptStoredScript),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccDataSourceIngestProcessorScriptResourceName, "id"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "source", "ctx.result = 'ok';"),
					resource.TestCheckResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "description"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "lang"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "script_id"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "params"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "if"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "on_failure.#"),
					resource.TestCheckNoResourceAttr(testAccDataSourceIngestProcessorScriptResourceName, "tag"),
					CheckResourceJSON(testAccDataSourceIngestProcessorScriptResourceName, "json", expectedJSONScriptMinimal),
				),
			},
		},
	})
}

const expectedScriptSource = `String[] envSplit = ctx['env'].splitOnToken(params['delimiter']);
ArrayList tags = new ArrayList();
tags.add(envSplit[params['position']].trim());
ctx['tags'] = tags;
`

const expectedScriptSourceAllAttributes = `ctx['tag_count'] = params['count'];
ctx['tag_prefix'] = params['prefix'];
`

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

const expectedJSONScriptAllAttributes = `{
	"script": {
		"description": "Annotate tags when env is present",
		"if": "ctx.env != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "script processor failed"
				}
			}
		],
		"tag": "script-tag",
		"lang": "expression",
		"params": {
			"count": 2,
			"prefix": "prod"
		},
		"source": "ctx['tag_count'] = params['count'];\nctx['tag_prefix'] = params['prefix'];\n"
	}
}`

const expectedJSONScriptStoredScript = `{
	"script": {
		"description": "Run stored script to derive tags",
		"ignore_failure": false,
		"lang": "painless",
		"id": "stored-script-derive-tags"
	}
}`

const expectedJSONScriptMinimal = `{
	"script": {
		"ignore_failure": false,
		"source": "ctx.result = 'ok';"
	}
}`
