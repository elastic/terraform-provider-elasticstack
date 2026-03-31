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

package provider_test

import (
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	apikey "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/api_key"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/go-version"
	tfconfig "github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionForFleet = version.Must(version.NewVersion("8.6.0"))

func TestProvider(t *testing.T) {
	if err := provider.New("dev").InternalValidate(); err != nil {
		t.Fatalf("Failed to validate provider: %s", err)
	}
}

func TestElasticsearchAPIKeyConnection(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(apikey.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: tfconfig.Variables{
					"api_key_name": tfconfig.StringVariable(apiKeyName),
					"endpoints":    tfconfig.StringVariable(os.Getenv("ELASTICSEARCH_ENDPOINTS")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "username", "elastic"),
				),
			},
		},
	})
}

func TestElasticsearchBearerTokenConnection(t *testing.T) {
	bearerToken := os.Getenv("ELASTICSEARCH_BEARER_TOKEN")
	if bearerToken == "" {
		t.Skip("ELASTICSEARCH_BEARER_TOKEN not set, skipping bearer token test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: tfconfig.Variables{
					"endpoints":    tfconfig.StringVariable(os.Getenv("ELASTICSEARCH_ENDPOINTS")),
					"bearer_token": tfconfig.StringVariable(bearerToken),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "username", "elastic"),
				),
			},
		},
	})
}

func TestFleetConfiguration(t *testing.T) {
	envConfig := config.NewFromEnv("acceptance-testing")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionForFleet),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          fleetConfigVariables(envConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.#"),
				),
			},
		},
	})
}

func TestFleetBearerTokenConfiguration(t *testing.T) {
	bearerToken := os.Getenv("FLEET_BEARER_TOKEN")
	if bearerToken == "" {
		t.Skip("FLEET_BEARER_TOKEN not set, skipping bearer token test")
	}

	envConfig := config.NewFromEnv("acceptance-testing")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionForFleet),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          fleetBearerTokenConfigVariables(envConfig, bearerToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.#"),
				),
			},
		},
	})
}

func TestKibanaConfiguration(t *testing.T) {
	var envConfig config.Client

	testCases := []struct {
		name string
		tc   func() resource.TestCase
		pre  func(t *testing.T)
		post func(t *testing.T)
	}{
		{
			name: "with username and password",
			pre: func(_ *testing.T) {
				envConfig = config.NewFromEnv("acceptance-testing")
			},
			post: func(_ *testing.T) {},
			tc: func() resource.TestCase {
				return resource.TestCase{
					PreCheck: func() { acctest.PreCheck(t) },
					Steps: []resource.TestStep{
						{
							ProtoV6ProviderFactories: acctest.Providers,
							SkipFunc: func() (bool, error) {
								return envConfig.Kibana.Username == "", nil
							},
							ConfigDirectory: acctest.NamedTestCaseDirectory("username_password"),
							ConfigVariables: kibanaUsernamePasswordConfigVariables(envConfig),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttrSet("elasticstack_kibana_space.acc_test", "name"),
							),
						},
					},
				}
			},
		},
		{
			name: "with api key",
			pre: func(t *testing.T) {
				apiKey := os.Getenv("KIBANA_API_KEY")
				t.Setenv("KIBANA_USERNAME", "")
				t.Setenv("KIBANA_PASSWORD", "")
				t.Setenv("KIBANA_API_KEY", apiKey)
				envConfig = config.NewFromEnv("acceptance-testing")
			},
			post: func(_ *testing.T) {},
			tc: func() resource.TestCase {
				return resource.TestCase{
					PreCheck: func() { acctest.PreCheck(t) },
					Steps: []resource.TestStep{
						{
							ProtoV6ProviderFactories: acctest.Providers,
							SkipFunc: func() (bool, error) {
								return os.Getenv("KIBANA_API_KEY") == "", nil
							},
							ConfigDirectory: acctest.NamedTestCaseDirectory("api_key"),
							ConfigVariables: kibanaAPIKeyConfigVariables(envConfig),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttrSet("elasticstack_kibana_space.acc_test", "name"),
							),
						},
					},
				}
			},
		},
		{
			name: "with bearer token",
			pre: func(t *testing.T) {
				bearerToken := os.Getenv("KIBANA_BEARER_TOKEN")
				t.Setenv("KIBANA_USERNAME", "")
				t.Setenv("KIBANA_PASSWORD", "")
				t.Setenv("KIBANA_API_KEY", "")
				t.Setenv("KIBANA_BEARER_TOKEN", bearerToken)
				envConfig = config.NewFromEnv("acceptance-testing")
			},
			post: func(_ *testing.T) {},
			tc: func() resource.TestCase {
				return resource.TestCase{
					PreCheck: func() { acctest.PreCheck(t) },
					Steps: []resource.TestStep{
						{
							ProtoV6ProviderFactories: acctest.Providers,
							SkipFunc: func() (bool, error) {
								return os.Getenv("KIBANA_BEARER_TOKEN") == "", nil
							},
							ConfigDirectory: acctest.NamedTestCaseDirectory("bearer_token"),
							ConfigVariables: kibanaBearerTokenConfigVariables(envConfig),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttrSet("elasticstack_kibana_space.acc_test", "name"),
							),
						},
					},
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.pre(t)
			resource.Test(t, tc.tc())
			tc.post(t)
		})

	}
}

func kibanaUsernamePasswordConfigVariables(cfg config.Client) tfconfig.Variables {
	return tfconfig.Variables{
		"kibana_endpoint": tfconfig.StringVariable(cfg.Kibana.Address),
		"kibana_username": tfconfig.StringVariable(cfg.Kibana.Username),
		"kibana_password": tfconfig.StringVariable(cfg.Kibana.Password),
	}
}

func kibanaAPIKeyConfigVariables(cfg config.Client) tfconfig.Variables {
	return tfconfig.Variables{
		"kibana_endpoint": tfconfig.StringVariable(cfg.Kibana.Address),
		"kibana_api_key":  tfconfig.StringVariable(cfg.Kibana.ApiKey),
	}
}

func fleetConfigVariables(cfg config.Client) tfconfig.Variables {
	caCertVars := make([]tfconfig.Variable, len(cfg.Fleet.CACerts))
	for i, ca := range cfg.Fleet.CACerts {
		caCertVars[i] = tfconfig.StringVariable(ca)
	}

	return tfconfig.Variables{
		"fleet_endpoint": tfconfig.StringVariable(cfg.Fleet.URL),
		"fleet_username": tfconfig.StringVariable(cfg.Fleet.Username),
		"fleet_password": tfconfig.StringVariable(cfg.Fleet.Password),
		"fleet_ca_certs": tfconfig.ListVariable(caCertVars...),
	}
}

func fleetBearerTokenConfigVariables(cfg config.Client, bearerToken string) tfconfig.Variables {
	caCertVars := make([]tfconfig.Variable, len(cfg.Fleet.CACerts))
	for i, ca := range cfg.Fleet.CACerts {
		caCertVars[i] = tfconfig.StringVariable(ca)
	}

	return tfconfig.Variables{
		"fleet_endpoint": tfconfig.StringVariable(cfg.Fleet.URL),
		"bearer_token":   tfconfig.StringVariable(bearerToken),
		"fleet_ca_certs": tfconfig.ListVariable(caCertVars...),
	}
}

func kibanaBearerTokenConfigVariables(cfg config.Client) tfconfig.Variables {
	return tfconfig.Variables{
		"kibana_endpoint":     tfconfig.StringVariable(cfg.Kibana.Address),
		"kibana_bearer_token": tfconfig.StringVariable(cfg.Kibana.BearerToken),
	}
}
