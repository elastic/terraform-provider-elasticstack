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

package enrich_test

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
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceEnrichPolicyFW(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkEnrichPolicyDestroyFW(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_enrich_policy.policy", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "match_field", `email`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.*", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "first_name"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "last_name"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "query", "{\"match_all\": {}}\n"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
		},
	})
}

func TestAccResourceEnrichPolicyNoExecute(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkEnrichPolicyDestroyFW(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_execute"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "match_field", "email"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "false"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.*", name),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "first_name"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "last_name"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "query", "{\"match_all\": {}}\n"),
					checkEnrichPolicyIndexDoesNotExist(name),
				),
			},
		},
	})
}

func TestAccResourceEnrichPolicyQueryOmitted(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkEnrichPolicyDestroyFW(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "match_field", "email"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
					checkEnrichPolicyQueryNull("elasticstack_elasticsearch_enrich_policy.policy"),
				),
			},
		},
	})
}

func TestAccResourceEnrichPolicyRangePolicyType(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkEnrichPolicyDestroyFW(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "range"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "match_field", "range_field"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.*", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "range_label"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
		},
	})
}

func TestAccResourceEnrichPolicyGeoMatchPolicyType(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkEnrichPolicyDestroyFW(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_enrich_policy.policy", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "geo_match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "match_field", "location"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.*", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "name"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyFW(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "first_name"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "last_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "query", "{\"match_all\":{}}"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					// Absent-state: no elasticsearch_connection block should appear in state
					// when the data source does not configure one.
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyGeoMatch(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "geo_match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "location"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "name"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "description"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyRange(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "range"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "ip_range"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "department"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "description"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyMultiIndex(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name+"-a"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name+"-b"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "first_name"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "last_name"),
				),
			},
		},
	})
}

func TestAccResourceEnrichPolicyFromSDK(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkEnrichPolicyDestroyFW(name),
		Steps: []resource.TestStep{
			{
				// Create the enrich policy with the last provider version where the enrich policy resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.17",
					},
				},
				Config: testAccEnrichPolicyFW(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("upgrade"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyConnection(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(name),
					"endpoints": config.ListVariable(config.StringVariable(primaryESEndpoint())),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "first_name"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyQueryNull(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "first_name"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "last_name"),
					checkEnrichPolicyQueryNull("data.elasticstack_elasticsearch_enrich_policy.test"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyQueryUpdate(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "city"),
					checkEnrichPolicyQueryNull("data.elasticstack_elasticsearch_enrich_policy.test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "city"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "query", `{"term":{"active":{"value":true}}}`),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyConnectionAPIKey(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := primaryESEndpoint()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name), "endpoint": config.StringVariable(primaryESEndpoint())},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "first_name"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "last_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "query", `{"match_all":{}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.0", endpoint),
					// api_key is sensitive — assert it is stored in state (non-empty)
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.api_key"),
					// headers map must contain exactly one entry with the expected value
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.headers.%", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.headers.X-Terraform-Test", "enrich-policy"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name), "endpoint": config.StringVariable(primaryESEndpoint())},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.api_key"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.headers.%"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.headers.X-Terraform-Test"),
				),
			},
			// Removal step: drop the elasticsearch_connection block and verify the
			// attributes are no longer present in state.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyConnectionBasicAuth(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := primaryESEndpoint()
	username := os.Getenv("ELASTICSEARCH_USERNAME")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheckESBasicAuth(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(name),
					"endpoint": config.StringVariable(primaryESEndpoint()),
					"username": config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password": config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.*", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "first_name"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.*", "last_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "query", `{"match_all":{}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.0", endpoint),
					// Basic auth credentials must be reflected in state
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.username", username),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.password"),
					// insecure=false was explicitly set in the config and must round-trip
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.insecure", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(name),
					"endpoint": config.StringVariable(primaryESEndpoint()),
					"username": config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password": config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.username", username),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.password"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.insecure", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyConnectionTLSInputs(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := primaryESEndpoint()
	tlsMaterial := createEnrichPolicyTLSMaterial(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("inline"),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(name),
					"endpoint":  config.StringVariable(endpoint),
					"ca_data":   config.StringVariable(tlsMaterial.caPEM),
					"cert_data": config.StringVariable(tlsMaterial.certPEM),
					"key_data":  config.StringVariable(tlsMaterial.keyPEM),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.ca_data", tlsMaterial.caPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.cert_data", tlsMaterial.certPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.key_data", tlsMaterial.keyPEM),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.ca_file"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.cert_file"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.key_file"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("file"),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(name),
					"endpoint":  config.StringVariable(endpoint),
					"ca_file":   config.StringVariable(tlsMaterial.caFile),
					"cert_file": config.StringVariable(tlsMaterial.certFile),
					"key_file":  config.StringVariable(tlsMaterial.keyFile),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.ca_file", tlsMaterial.caFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.cert_file", tlsMaterial.certFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.key_file", tlsMaterial.keyFile),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.ca_data"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.cert_data"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.key_data"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyConnectionValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ca_conflict"),
				ExpectError:              regexp.MustCompile(`(?s)(Invalid Attribute Combination|ca_file.*ca_data|ca_data.*ca_file)`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("cert_data"),
				ExpectError:              regexp.MustCompile(`(?s)(Missing Configuration for Required Attribute|cert_data.*key_data|key_data.*cert_data)`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("cert_file"),
				ExpectError:              regexp.MustCompile(`(?s)(Missing Configuration for Required Attribute|cert_file.*key_file|key_file.*cert_file)`),
			},
			// key_file without cert_file must also be rejected
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("key_file"),
				ExpectError:              regexp.MustCompile(`(?s)(Missing Configuration for Required Attribute|key_file.*cert_file|cert_file.*key_file)`),
			},
			// key_data without cert_data must also be rejected
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("key_data"),
				ExpectError:              regexp.MustCompile(`(?s)(Missing Configuration for Required Attribute|key_data.*cert_data|cert_data.*key_data)`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("client_auth"),
				ExpectError:              regexp.MustCompile(`(?s)(Missing Configuration for Required Attribute|es_client_authentication.*bearer_token|bearer_token.*es_client_authentication)`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multiple_blocks"),
				ExpectError:              regexp.MustCompile(`(?s)(at most 1 elements|at most 1 element|elasticsearch_connection)`),
			},
		},
	})
}

func TestAccResourceEnrichPolicyConnection(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkEnrichPolicyDestroyFW(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(name),
					"endpoints": config.ListVariable(config.StringVariable(primaryESEndpoint())),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_enrich_policy.policy", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "match_field", "email"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.*", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.*", "first_name"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyTermQuery(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "query", `{"term":{"active":{"value":true}}}`),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyBearerToken(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoint := primaryESEndpoint()
	var bearerToken string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			preCheckESBasicAuth(t)
			bearerToken = createEnrichPolicyESAccessToken(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":         config.StringVariable(name),
					"endpoint":     config.StringVariable(endpoint),
					"bearer_token": config.StringVariable(bearerToken),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_enrich_policy.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					// Connection block must be present with correct endpoint
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.bearer_token", bearerToken),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "elasticsearch_connection.0.es_client_authentication", "Authorization"),
				),
			},
		},
	})
}

func testAccEnrichPolicyFW(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      email      = { type = "text" }
      first_name = { type = "text" }
      last_name  = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = "%s"
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "email"
  enrich_fields = ["first_name", "last_name"]
	query = <<-EOD
	{"match_all": {}}
	EOD
}
	`, name, name)
}

func preCheckESBasicAuth(t *testing.T) {
	acctest.PreCheck(t)
	if os.Getenv("ELASTICSEARCH_USERNAME") == "" || os.Getenv("ELASTICSEARCH_PASSWORD") == "" {
		t.Skip("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD must be set for explicit basic auth coverage")
	}
}

func primaryESEndpoint() string {
	for endpoint := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint != "" {
			return endpoint
		}
	}

	return "http://localhost:9200"
}

func createEnrichPolicyESAccessToken(t *testing.T) string {
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

type enrichPolicyTLSMaterial struct {
	caPEM    string
	certPEM  string
	keyPEM   string
	caFile   string
	certFile string
	keyFile  string
}

func createEnrichPolicyTLSMaterial(t *testing.T) enrichPolicyTLSMaterial {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "enrich-policy-test",
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
			CommonName: "enrich-policy-test",
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

	return enrichPolicyTLSMaterial{
		caPEM:    certPEM,
		certPEM:  certPEM,
		keyPEM:   keyPEM,
		caFile:   caFile,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func checkEnrichPolicyDestroyFW(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_enrich_policy" {
				continue
			}
			compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)
			if compID.ResourceID != name {
				return fmt.Errorf("Found unexpectedly enrich policy: %s", compID.ResourceID)
			}
			esClient, err := client.GetESClient()
			if err != nil {
				return err
			}
			req := esClient.EnrichGetPolicy.WithName(compID.ResourceID)
			res, err := esClient.EnrichGetPolicy(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode == http.StatusFound {
				var policiesResponse map[string]any
				if err := json.NewDecoder(res.Body).Decode(&policiesResponse); err != nil {
					return err
				}
				if len(policiesResponse["policies"].([]any)) != 0 {
					return fmt.Errorf("Enrich policy (%s) still exists", compID.ResourceID)
				}
			}
		}
		return nil
	}
}

func checkEnrichPolicyIndexDoesNotExist(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		indexName := fmt.Sprintf(".enrich-%s", name)
		res, err := esClient.Indices.Exists([]string{indexName})
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Expected enrich index alias %s to be missing, got status %d", indexName, res.StatusCode)
		}

		return nil
	}
}

func checkEnrichPolicyQueryNull(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		value, ok := rs.Primary.Attributes["query"]
		if !ok || value == "" || value == "null" {
			return nil
		}

		return fmt.Errorf("Expected query to be null, got %q", value)
	}
}
