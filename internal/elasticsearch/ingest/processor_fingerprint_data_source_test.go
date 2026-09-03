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

func TestAccDataSourceIngestProcessorFingerprint(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.0", "user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "target_field", "fingerprint"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "method", "SHA-1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "json", expectedJSONFingerprint),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.#", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.0", "user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.1", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.2", "ip"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "target_field", "doc_fingerprint"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "method", "SHA-256"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "salt", "my-secret-salt"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "description", "Fingerprint for dedup"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "if", "ctx.env == 'prod'"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "tag", "fingerprint-docs"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "json", expectedJSONFingerprintAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "on_failure.#", "2"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "on_failure.0", `{"set":{"field":"error.message","value":"fingerprint failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "on_failure.1", `{"set":{"field":"error.type","value":"fingerprint"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "json", expectedJSONFingerprintOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "fields.0", "user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "target_field", "fingerprint"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "method", "SHA-1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "salt"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "tag"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "on_failure.#"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "json", expectedJSONFingerprintDefaults),
				),
			},
		},
	})
}

const expectedJSONFingerprint = `{
  "fingerprint": {
		"fields": [
			"user"
		],
		"ignore_failure": false,
		"ignore_missing": false,
		"method": "SHA-1",
		"target_field": "fingerprint"
	}
}
`

const expectedJSONFingerprintAllAttributes = `{
	"fingerprint": {
		"description": "Fingerprint for dedup",
		"if": "ctx.env == 'prod'",
		"ignore_failure": true,
		"tag": "fingerprint-docs",
		"fields": [
			"user",
			"email",
			"ip"
		],
		"ignore_missing": true,
		"method": "SHA-256",
		"salt": "my-secret-salt",
		"target_field": "doc_fingerprint"
	}
}`

const expectedJSONFingerprintOnFailure = `{
	"fingerprint": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "fingerprint failed"
				}
			},
			{
				"set": {
					"field": "error.type",
					"value": "fingerprint"
				}
			}
		],
		"fields": [
			"user"
		],
		"ignore_missing": false,
		"method": "SHA-1",
		"target_field": "fingerprint"
	}
}`

const expectedJSONFingerprintDefaults = `{
	"fingerprint": {
		"ignore_failure": false,
		"fields": [
			"user"
		],
		"ignore_missing": false,
		"method": "SHA-1",
		"target_field": "fingerprint"
	}
}`
