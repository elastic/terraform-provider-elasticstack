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

func TestAccDataSourceIngestProcessorNetworkDirection(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.*", "private"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "json", expectedJSONNetworkDirection),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "source_ip", "source.ip"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "destination_ip", "destination.ip"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "target_field", "network.direction"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.*", "private"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.*", "loopback"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "description", "Infer direction for private and loopback traffic"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "if", "ctx.source?.ip != null && ctx.destination?.ip != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "tag", "network-direction"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_failure", "true"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "json", expectedJSONNetworkDirectionAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("internal_networks_field"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks_field", "network.private_ranges"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.#"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "json", expectedJSONNetworkDirectionInternalNetworksField),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("boolean_variations"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.*", "private"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "ignore_failure", "true"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "json", expectedJSONNetworkDirectionBooleanVariations),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "internal_networks.*", "private"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "on_failure.0", `{"set":{"field":"error.message","value":"network direction failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "json", expectedJSONNetworkDirectionOnFailure),
				),
			},
		},
	})
}

const expectedJSONNetworkDirection = `{
	"network_direction": {
		"ignore_failure": false,
		"ignore_missing": true,
		"internal_networks": [
			"private"
		]
	}
}`

const expectedJSONNetworkDirectionAllAttributes = `{
	"network_direction": {
		"description": "Infer direction for private and loopback traffic",
		"if": "ctx.source?.ip != null && ctx.destination?.ip != null",
		"ignore_failure": true,
		"tag": "network-direction",
		"source_ip": "source.ip",
		"destination_ip": "destination.ip",
		"target_field": "network.direction",
		"internal_networks": [
			"loopback",
			"private"
		],
		"ignore_missing": true
	}
}`

const expectedJSONNetworkDirectionInternalNetworksField = `{
	"network_direction": {
		"ignore_failure": false,
		"internal_networks_field": "network.private_ranges",
		"ignore_missing": true
	}
}`

const expectedJSONNetworkDirectionBooleanVariations = `{
	"network_direction": {
		"ignore_failure": true,
		"internal_networks": [
			"private"
		],
		"ignore_missing": false
	}
}`

const expectedJSONNetworkDirectionOnFailure = `{
	"network_direction": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "network direction failed"
				}
			}
		],
		"internal_networks": [
			"private"
		],
		"ignore_missing": true
	}
}`
