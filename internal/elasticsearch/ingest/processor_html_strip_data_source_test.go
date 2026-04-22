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

func TestAccDataSourceIngestProcessorHTMLStrip(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "field", "foo"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "json", expectedJSONHTMLStrip),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "field", "body.html"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "target_field", "body.plain"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "description", "Strip HTML markup from body content"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "if", "ctx.body?.html != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "tag", "html-strip-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "json", expectedJSONHTMLStripAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "field", "content.html"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "tag"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "on_failure.#"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_html_strip.test", "json", expectedJSONHTMLStripDefaults),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorHTMLStripOnFailure(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_html_strip.test_on_failure", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test_on_failure", "field", "body.html"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_html_strip.test_on_failure", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_html_strip.test_on_failure", "on_failure.0", `{"set":{"field":"error.message","value":"html strip failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_html_strip.test_on_failure", "json", expectedJSONHTMLStripOnFailure),
				),
			},
		},
	})
}

const expectedJSONHTMLStrip = `{
	"html_strip": {
		"field": "foo",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONHTMLStripAllAttributes = `{
	"html_strip": {
		"description": "Strip HTML markup from body content",
		"if": "ctx.body?.html != null",
		"ignore_failure": true,
		"tag": "html-strip-tag",
		"field": "body.html",
		"target_field": "body.plain",
		"ignore_missing": true
	}
}`

const expectedJSONHTMLStripDefaults = `{
	"html_strip": {
		"field": "content.html",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONHTMLStripOnFailure = `{
	"html_strip": {
		"field": "body.html",
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "html strip failed"
				}
			}
		],
		"ignore_missing": false
	}
}`
