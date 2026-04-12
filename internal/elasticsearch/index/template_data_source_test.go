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

package index_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamlifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccIndexTemplateDataSource(t *testing.T) {
	// generate a random role name
	templateName := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	templateNameComponent := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("tf-acc-%s-*", templateName),
					),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "priority", "100"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ignore_component"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateNameComponent),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test_component",
						"index_patterns.*",
						fmt.Sprintf("tf-acc-component-%s-*", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"data.elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceTemplate covers the template subtree:
// aliases (name, filter, index_routing, search_routing, is_hidden, is_write_index), mappings, and settings.
func TestAccIndexTemplateDataSourceTemplate(t *testing.T) {
	templateName := "test-ds-tpl-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"mappings":      config.StringVariable(`{"properties":{"log_level":{"type":"keyword"}}}`),
					"settings":      config.StringVariable(`{"index":{"number_of_shards":"1"}}`),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":           "my_alias",
							"filter":         `{"term":{"status":"active"}}`,
							"index_routing":  "shard_1",
							"search_routing": "shard_1",
							"is_hidden":      "false",
							"is_write_index": "true",
						},
					),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"}}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"1"}}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"mappings":      config.StringVariable(`{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`),
					"settings":      config.StringVariable(`{"index":{"number_of_shards":"2"}}`),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_index_template.test",
						"template.0.alias.*",
						map[string]string{
							"name":           "my_alias",
							"filter":         `{"term":{"status":"active"}}`,
							"index_routing":  "shard_1",
							"search_routing": "shard_1",
							"is_hidden":      "false",
							"is_write_index": "true",
						},
					),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"},"severity":{"type":"integer"}}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"2"}}`),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceExplicitConnection(t *testing.T) {
	templateName := "test-ds-conn-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := indexTemplateDataSourcePrimaryESEndpoint(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { preCheckIndexTemplateDataSourceBasicAuth(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"endpoint":      config.StringVariable(endpoint),
					"username":      config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":      config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.username", os.Getenv("ELASTICSEARCH_USERNAME")),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.password"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.headers.%", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.headers.XTerraformTest", "basic-auth"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.insecure", "true"),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceExplicitConnectionAPIKey(t *testing.T) {
	templateName := "test-ds-api-key-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoints := indexTemplateDataSourceConnectionEndpoints(t)
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(endpoint))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"endpoints":     config.ListVariable(endpointVars...),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.api_key"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.username", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.password", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.bearer_token", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.1", endpoints[1]),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.headers.%", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.headers.XTerraformTest", "api-key"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.headers.XTrace", "index-template"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.insecure", "false"),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceExplicitConnectionBearerToken(t *testing.T) {
	templateName := "test-ds-bearer-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := indexTemplateDataSourcePrimaryESEndpoint(t)
	var bearerToken string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			preCheckIndexTemplateDataSourceBasicAuth(t)
			bearerToken = createIndexTemplateDataSourceESAccessToken(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"endpoint":      config.StringVariable(endpoint),
					"bearer_token":  config.StringVariable(bearerToken),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.bearer_token", bearerToken),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.es_client_authentication", "Authorization"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.api_key", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.username", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.password", ""),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceExplicitConnectionTLSInputs(t *testing.T) {
	templateName := "test-ds-tls-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := indexTemplateDataSourcePrimaryESEndpoint(t)
	tlsMaterial := createIndexTemplateDataSourceTLSMaterial(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("inline"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"endpoint":      config.StringVariable(endpoint),
					"ca_data":       config.StringVariable(tlsMaterial.caPEM),
					"cert_data":     config.StringVariable(tlsMaterial.certPEM),
					"key_data":      config.StringVariable(tlsMaterial.keyPEM),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.ca_data", tlsMaterial.caPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.cert_data", tlsMaterial.certPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.key_data", tlsMaterial.keyPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.ca_file", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.cert_file", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.key_file", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("file"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"endpoint":      config.StringVariable(endpoint),
					"ca_file":       config.StringVariable(tlsMaterial.caFile),
					"cert_file":     config.StringVariable(tlsMaterial.certFile),
					"key_file":      config.StringVariable(tlsMaterial.keyFile),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.ca_file", tlsMaterial.caFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.cert_file", tlsMaterial.certFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.key_file", tlsMaterial.keyFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.ca_data", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.cert_data", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test_conn", "elasticsearch_connection.0.key_data", ""),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceDataStream covers data_stream.0.hidden and data_stream.0.allow_custom_routing.
func TestAccIndexTemplateDataSourceDataStream(t *testing.T) {
	templateName := "test-ds-stream-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"hidden":        config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceMetadataVersionID covers metadata, version, and the id attribute.
func TestAccIndexTemplateDataSourceMetadataVersionID(t *testing.T) {
	templateName := "test-ds-meta-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name":    config.StringVariable(templateName),
					"metadata":         config.StringVariable(`{"owner":"team-a","description":"initial"}`),
					"template_version": config.StringVariable("5"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "version", "5"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"initial","owner":"team-a"}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name":    config.StringVariable(templateName),
					"metadata":         config.StringVariable(`{"owner":"team-b","description":"updated"}`),
					"template_version": config.StringVariable("7"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_index_template.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "version", "7"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"updated","owner":"team-b"}`),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceCountAssertions covers index_patterns.#, composed_of.#,
// and ignore_missing_component_templates.# with multi-value assertions.
func TestAccIndexTemplateDataSourceCountAssertions(t *testing.T) {
	templateName := "test-ds-count-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	comp1 := templateName + "-comp1@custom"
	comp2 := templateName + "-comp2@custom"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"component_1":   config.StringVariable(comp1),
					"component_2":   config.StringVariable(comp2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%s-a-*", templateName)),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%s-b-*", templateName)),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.*", comp1),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.*", comp2),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", comp1),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", comp2),
				),
			},
		},
	})
}

// TestAccIndexTemplateDataSourceLifecycle covers template.0.lifecycle.*.data_retention.
func TestAccIndexTemplateDataSourceLifecycle(t *testing.T) {
	templateName := "test-ds-lifecycle-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.lifecycle.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_index_template.test",
						"template.0.lifecycle.*",
						map[string]string{
							"data_retention": "30d",
						},
					),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceOrderedComponents(t *testing.T) {
	templateName := "test-ds-order-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	comp1 := templateName + "-comp1@custom"
	comp2 := templateName + "-comp2@custom"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
					"component_1":   config.StringVariable(comp1),
					"component_2":   config.StringVariable(comp2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.0", comp1),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.1", comp2),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.0", comp1),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.1", comp2),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.#", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.#", "0"),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceDataStreamEmptyObject(t *testing.T) {
	templateName := "test-ds-empty-stream-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.allow_custom_routing", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAllowCustomRoutingVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "data_stream.0.hidden", "false"),
					testCheckDataStreamAttrFalseOrAbsent("data.elasticstack_elasticsearch_index_template.test", "allow_custom_routing"),
				),
			},
		},
	})
}

func TestAccIndexTemplateDataSourceOptionalFieldRemoval(t *testing.T) {
	templateName := "test-ds-remove-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "metadata", `{"description":"initial","owner":"team-a"}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "version", "7"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings", `{"properties":{"log_level":{"type":"keyword"}}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.settings", `{"index":{"number_of_shards":"1"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.0.lifecycle.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_index_template.test",
						"template.0.lifecycle.*",
						map[string]string{
							"data_retention": "30d",
						},
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "template.#", "1"),
					testCheckDataSourceAttrEmptyOrAbsent("data.elasticstack_elasticsearch_index_template.test", "metadata"),
					testCheckAttrZeroOrAbsent("data.elasticstack_elasticsearch_index_template.test", "version"),
					testCheckDataSourceAttrEmptyOrAbsent("data.elasticstack_elasticsearch_index_template.test", "template.0.mappings"),
					testCheckDataSourceAttrEmptyOrAbsent("data.elasticstack_elasticsearch_index_template.test", "template.0.settings"),
					testCheckAttrZeroOrAbsent("data.elasticstack_elasticsearch_index_template.test", "template.0.lifecycle.#"),
					testCheckDataSourceTemplateLifecycleAttrCleared("data.elasticstack_elasticsearch_index_template.test", "data_retention"),
					testCheckNoTemplateAliases("data.elasticstack_elasticsearch_index_template.test"),
				),
			},
		},
	})
}

func preCheckIndexTemplateDataSourceBasicAuth(t *testing.T) {
	acctest.PreCheck(t)
	if os.Getenv("ELASTICSEARCH_USERNAME") == "" || os.Getenv("ELASTICSEARCH_PASSWORD") == "" {
		t.Skip("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD must be set for explicit basic auth coverage")
	}
}

func indexTemplateDataSourcePrimaryESEndpoint(t *testing.T) string {
	for endpoint := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		if trimmed := strings.TrimSpace(endpoint); trimmed != "" {
			return trimmed
		}
	}

	t.Fatal("ELASTICSEARCH_ENDPOINTS must contain at least one endpoint")
	return ""
}

func indexTemplateDataSourceConnectionEndpoints(t *testing.T) []string {
	endpoints := make([]string, 0, 2)
	for endpoint := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		if trimmed := strings.TrimSpace(endpoint); trimmed != "" {
			endpoints = append(endpoints, trimmed)
			if len(endpoints) == 2 {
				return endpoints
			}
		}
	}

	if len(endpoints) == 1 {
		return append(endpoints, endpoints[0])
	}

	t.Fatal("ELASTICSEARCH_ENDPOINTS must contain at least one endpoint")
	return nil
}

func createIndexTemplateDataSourceESAccessToken(t *testing.T) string {
	t.Helper()

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("failed to create acceptance testing client: %v", err)
	}
	esClient, err := client.GetESClient()
	if err != nil {
		t.Fatalf("failed to get Elasticsearch client: %v", err)
	}

	payload, err := json.Marshal(map[string]string{
		"grant_type": "password",
		"username":   os.Getenv("ELASTICSEARCH_USERNAME"),
		"password":   os.Getenv("ELASTICSEARCH_PASSWORD"),
	})
	if err != nil {
		t.Fatalf("failed to marshal token request: %v", err)
	}

	resp, err := esClient.Security.GetToken(
		bytes.NewReader(payload),
		esClient.Security.GetToken.WithContext(context.Background()),
	)
	if err != nil {
		t.Fatalf("failed to create Elasticsearch access token: %v", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			t.Fatalf("failed to create Elasticsearch access token: status %d (additionally failed to read error response: %v)", resp.StatusCode, readErr)
		}
		t.Fatalf("failed to create Elasticsearch access token: status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	if tokenResponse.AccessToken == "" {
		t.Fatalf("token response did not include an access_token")
	}

	return tokenResponse.AccessToken
}

type indexTemplateDataSourceTLSMaterial struct {
	caPEM    string
	certPEM  string
	keyPEM   string
	caFile   string
	certFile string
	keyFile  string
}

func createIndexTemplateDataSourceTLSMaterial(t *testing.T) indexTemplateDataSourceTLSMaterial {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "index-template-data-source-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}, &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "index-template-data-source-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER}))
	keyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}))

	tempDir := t.TempDir()
	caFile := filepath.Join(tempDir, "ca.pem")
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")

	for path, contents := range map[string]string{
		caFile:   certPEM,
		certFile: certPEM,
		keyFile:  keyPEM,
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatalf("failed to write TLS test file %s: %v", path, err)
		}
	}

	return indexTemplateDataSourceTLSMaterial{
		caPEM:    certPEM,
		certPEM:  certPEM,
		keyPEM:   keyPEM,
		caFile:   caFile,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func testCheckDataSourceAttrEmptyOrAbsent(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		value, ok := rs.Primary.Attributes[attrName]
		if ok && value != "" {
			return fmt.Errorf("expected %s to be empty or absent, got %q", attrName, value)
		}
		return nil
	}
}

func testCheckDataSourceTemplateLifecycleAttrCleared(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		for key, value := range rs.Primary.Attributes {
			if strings.HasPrefix(key, "template.0.lifecycle.") && strings.HasSuffix(key, "."+attrName) && value != "" {
				return fmt.Errorf("expected lifecycle attribute %s to be cleared, got %q", key, value)
			}
		}
		return nil
	}
}
