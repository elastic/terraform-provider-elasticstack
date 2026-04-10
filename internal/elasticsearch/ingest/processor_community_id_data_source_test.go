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

func TestAccDataSourceIngestProcessorCommunityID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "seed", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "json", expectedJSONCommunityID),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorCommunityIDCoreNetwork(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "source_ip", "source.address"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "source_port", "12345"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "destination_ip", "destination.address"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "destination_port", "443"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "target_field", "network.community_id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "seed", "123"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_failure", "true"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "json", expectedJSONCommunityIDCoreNetwork),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorCommunityIDICMP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "transport", "icmp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "icmp_type", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "icmp_code", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "seed", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "json", expectedJSONCommunityIDICMP),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorCommunityIDMetadata(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "description", "Compute the community ID"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "if", "ctx.network != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "tag", "community-id-tag"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "seed", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "json", expectedJSONCommunityIDMetadata),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorCommunityIDOnFailure(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_community_id.test_on_failure", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_community_id.test_on_failure", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_community_id.test_on_failure", "on_failure.0", `{"set":{"field":"error.message","value":"community id failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_community_id.test_on_failure", "json", expectedJSONCommunityIDOnFailure),
				),
			},
		},
	})
}

const expectedJSONCommunityID = `{
	"community_id": {
		"seed": 0,
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONCommunityIDCoreNetwork = `{
	"community_id": {
		"ignore_failure": true,
		"source_ip": "source.address",
		"source_port": 12345,
		"destination_ip": "destination.address",
		"destination_port": 443,
		"target_field": "network.community_id",
		"seed": 123,
		"ignore_missing": true
	}
}`

const expectedJSONCommunityIDICMP = `{
	"community_id": {
		"ignore_failure": false,
		"icmp_type": 3,
		"icmp_code": 1,
		"transport": "icmp",
		"seed": 0,
		"ignore_missing": false
	}
}`

const expectedJSONCommunityIDMetadata = `{
	"community_id": {
		"description": "Compute the community ID",
		"if": "ctx.network != null",
		"ignore_failure": false,
		"tag": "community-id-tag",
		"seed": 0,
		"ignore_missing": false
	}
}`

const expectedJSONCommunityIDOnFailure = `{
	"community_id": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "community id failed"
				}
			}
		],
		"seed": 0,
		"ignore_missing": false
	}
}`
