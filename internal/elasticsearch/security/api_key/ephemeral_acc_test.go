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

package apikey_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	apikey "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/api_key"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func ephemeralTestProviders() map[string]func() (tfprotov6.ProviderServer, error) {
	providers := make(map[string]func() (tfprotov6.ProviderServer, error), len(acctest.Providers)+1)
	maps.Copy(providers, acctest.Providers)
	providers["echo"] = echoprovider.NewProviderServer()
	return providers
}

func ephemeralTerraformVersionChecks() []tfversion.TerraformVersionCheck {
	return []tfversion.TerraformVersionCheck{
		tfversion.SkipBelow(tfversion.Version1_10_0),
	}
}

func TestAccEphemeralResourceSecurityAPIKey(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: ephemeralTerraformVersionChecks(),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: ephemeralTestProviders(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.capture", tfjsonpath.New("data").AtMapKey("api_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.capture", tfjsonpath.New("data").AtMapKey("encoded"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.capture", tfjsonpath.New("data").AtMapKey("key_id"), knownvalue.NotNull()),
				},
				Check: resource.ComposeTestCheckFunc(
					checkElasticstackStateHasNoAPIKeySecrets,
					checkEchoCaptureInt64Equals("expiration_timestamp", 0),
					checkEphemeralAPIKeyExistsInElasticsearch(false),
					cleanupEphemeralAPIKeyFromEchoCapture,
				),
			},
		},
	})
}

func TestAccEphemeralResourceSecurityAPIKeyInvalidateOnClose(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: ephemeralTerraformVersionChecks(),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: ephemeralTestProviders(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.capture", tfjsonpath.New("data").AtMapKey("key_id"), knownvalue.NotNull()),
				},
				Check: resource.ComposeTestCheckFunc(
					checkEphemeralAPIKeyExistsInElasticsearch(true),
				),
			},
		},
	})
}

func TestAccEphemeralResourceSecurityAPIKeyWithExpiration(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: ephemeralTerraformVersionChecks(),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: ephemeralTestProviders(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkEchoCaptureInt64GreaterThanZero("expiration_timestamp"),
					cleanupEphemeralAPIKeyFromEchoCapture,
				),
			},
		},
	})
}

func TestAccEphemeralResourceSecurityAPIKeyCrossCluster(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersionWithCrossCluster, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: ephemeralTerraformVersionChecks(),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: ephemeralTestProviders(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.capture", tfjsonpath.New("data").AtMapKey("encoded"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.capture", tfjsonpath.New("data").AtMapKey("key_id"), knownvalue.NotNull()),
				},
				Check: resource.ComposeTestCheckFunc(
					checkEphemeralAPIKeyExistsInElasticsearch(false),
					cleanupEphemeralAPIKeyFromEchoCapture,
				),
			},
		},
	})
}

func TestAccEphemeralResourceSecurityAPIKeyExplicitConnection(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoints := ephemeralESEndpoints()
	if len(endpoints) == 0 {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must contain at least one endpoint")
	}

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: ephemeralTerraformVersionChecks(),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: ephemeralTestProviders(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: func() config.Variables {
					endpointVars := make([]config.Variable, len(endpoints))
					for i, ep := range endpoints {
						endpointVars[i] = config.StringVariable(ep)
					}
					return config.Variables{
						"api_key_name": config.StringVariable(apiKeyName),
						"endpoints":    config.ListVariable(endpointVars...),
						"api_key":      config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
						"username":     config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
						"password":     config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
					}
				}(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.capture", tfjsonpath.New("data").AtMapKey("key_id"), knownvalue.NotNull()),
				},
				Check: resource.ComposeTestCheckFunc(
					checkEphemeralAPIKeyExistsInElasticsearch(true),
				),
			},
		},
	})
}

func ephemeralESEndpoints() []string {
	rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	parts := strings.Split(rawEndpoints, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			endpoints = append(endpoints, part)
		}
	}
	return endpoints
}

func checkEchoCaptureInt64Equals(field string, expected int64) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rawValue, err := echoCaptureValue(state, field)
		if err != nil {
			return err
		}
		value, ok := rawValue.(float64)
		if !ok {
			return fmt.Errorf("expected echo.capture.data.%s to be a number, got %#v", field, rawValue)
		}
		if int64(value) != expected {
			return fmt.Errorf("expected echo.capture.data.%s to be %d, got %d", field, expected, int64(value))
		}
		return nil
	}
}

func checkElasticstackStateHasNoAPIKeySecrets(state *terraform.State) error {
	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type == "elasticstack_elasticsearch_security_api_key" {
			return fmt.Errorf("managed elasticstack_elasticsearch_security_api_key resource must not exist when using the ephemeral resource")
		}
		if resourceState.Provider == "provider[\"registry.terraform.io/elastic/elasticstack\"]" {
			if _, ok := resourceState.Primary.Attributes["api_key"]; ok {
				return fmt.Errorf("elasticstack resource %s must not persist api_key in state", resourceState.Type)
			}
			if _, ok := resourceState.Primary.Attributes["encoded"]; ok {
				return fmt.Errorf("elasticstack resource %s must not persist encoded in state", resourceState.Type)
			}
		}
	}
	return nil
}

func checkEphemeralAPIKeyExistsInElasticsearch(expectInvalidated bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		keyID, err := echoCaptureString(state, "key_id")
		if err != nil {
			return err
		}

		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		apiKey, diags := elasticsearch.GetAPIKey(context.Background(), client, keyID)
		if diags.HasError() {
			return fmt.Errorf("unable to get API key %q: %v", keyID, diags)
		}
		if apiKey == nil {
			if expectInvalidated {
				return nil
			}
			return fmt.Errorf("expected API key %q to exist in Elasticsearch", keyID)
		}
		if expectInvalidated && !apiKey.Invalidated {
			return fmt.Errorf("expected API key %q to be invalidated", keyID)
		}
		if !expectInvalidated && apiKey.Invalidated {
			return fmt.Errorf("expected API key %q to remain valid", keyID)
		}
		return nil
	}
}

func checkEchoCaptureInt64GreaterThanZero(field string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rawValue, err := echoCaptureValue(state, field)
		if err != nil {
			return err
		}
		value, ok := rawValue.(float64)
		if !ok || value <= 0 {
			return fmt.Errorf("expected echo.capture.data.%s to be a positive number, got %#v", field, rawValue)
		}
		return nil
	}
}

func cleanupEphemeralAPIKeyFromEchoCapture(state *terraform.State) error {
	keyID, err := echoCaptureString(state, "key_id")
	if err != nil {
		return err
	}

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	diags := elasticsearch.DeleteAPIKey(context.Background(), client, keyID)
	if diags.HasError() {
		return fmt.Errorf("failed to invalidate API key %q: %v", keyID, diags)
	}
	return nil
}

func echoCaptureString(state *terraform.State, field string) (string, error) {
	value, err := echoCaptureValue(state, field)
	if err != nil {
		return "", err
	}
	stringValue, ok := value.(string)
	if !ok || stringValue == "" {
		return "", fmt.Errorf("expected echo.capture.data.%s to be a non-empty string, got %#v", field, value)
	}
	return stringValue, nil
}

func echoCaptureValue(state *terraform.State, field string) (any, error) {
	resourceState, ok := state.RootModule().Resources["echo.capture"]
	if !ok {
		return nil, fmt.Errorf("echo.capture resource not found in state")
	}

	dataJSON, ok := resourceState.Primary.Attributes["data"]
	if !ok {
		return nil, fmt.Errorf("echo.capture.data attribute not found in state")
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return nil, fmt.Errorf("failed to parse echo.capture.data: %w", err)
	}

	value, ok := data[field]
	if !ok {
		return nil, fmt.Errorf("field %q not found in echo.capture.data", field)
	}
	return value, nil
}
