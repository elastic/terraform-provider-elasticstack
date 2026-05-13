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
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	apikey "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/api_key"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

//go:embed testdata/TestAccResourceSecurityAPIKeyFromSDK/no_expiration/main.tf
var testAccResourceSecurityAPIKeyFromSDKConfig string

func TestAccResourceSecurityAPIKey(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(apikey.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: new(false),
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration_timestamp"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "key_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(apikey.MinVersionWithUpdate),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"manage"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-b*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: new(false),
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration_timestamp"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "key_id"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyRotation(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	var firstKeyID string
	var firstRotationID string

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("rotation"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
					"epoch":        config.StringVariable("1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("time_rotating.api_key_rotation", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "key_id"),
					resource.TestCheckResourceAttrWith("time_rotating.api_key_rotation", "id", func(value string) error {
						firstRotationID = value
						if value == "" {
							return fmt.Errorf("expected time_rotating id to be non-empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "key_id", func(value string) error {
						firstKeyID = value
						if value == "" {
							return fmt.Errorf("expected key_id to be non-empty")
						}
						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("rotation"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
					"epoch":        config.StringVariable("2"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("time_rotating.api_key_rotation", "id", func(value string) error {
						if value == firstRotationID {
							return fmt.Errorf("expected rotation id to change after updating epoch, got %q", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "key_id", func(value string) error {
						if value == firstKeyID {
							return fmt.Errorf("expected api key key_id to change after rotation, got %q", value)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyWithRemoteIndices(t *testing.T) {
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))

	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, minSupportedRemoteIndicesVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: new(false),
								}},
								RemoteIndices: []models.RemoteIndexPerms{{
									Clusters: []string{"*"},
									IndexPerms: models.IndexPerms{
										Names:                  []string{"index-a*"},
										Privileges:             []string{"read"},
										AllowRestrictedIndices: new(true),
									},
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyWithWorkflowRestriction(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersionWithRestriction, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						allowRestrictedIndices := false
						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: &allowRestrictedIndices,
								}},
								Restriction: &models.Restriction{Workflows: []string{"search_application_query"}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyWithWorkflowRestrictionOnElasticPre8_9_x(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	errorPattern := fmt.Sprintf(".*Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch. Role descriptor\\(s\\) %s.*", "role-a")
	errorPattern = strings.ReplaceAll(errorPattern, " ", "\\s+")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 SkipWhenAPIKeysAreNotSupportedOrRestrictionsAreSupported(apikey.MinVersion, apikey.MinVersionWithRestriction),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				ExpectError: regexp.MustCompile(errorPattern),
			},
		},
	})
}

func SkipWhenAPIKeysAreNotSupportedOrRestrictionsAreSupported(minAPIKeySupportedVersion *version.Version, minRestrictionSupportedVersion *version.Version) func() (bool, error) {
	return func() (b bool, err error) {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return false, err
		}
		serverVersion, diags := client.ServerVersion(context.Background())
		if diags.HasError() {
			return false, fmt.Errorf("failed to parse the elasticsearch version %v", diags)
		}

		return serverVersion.LessThan(minAPIKeySupportedVersion) || serverVersion.GreaterThanOrEqual(minRestrictionSupportedVersion), nil
	}
}

func TestAccResourceSecurityAPIKeyFromSDK(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	var initialAPIKey string

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				// Create the api_key with the last provider version where the api_key resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.9",
					},
				},
				Config: testAccResourceSecurityAPIKeyFromSDKConfig,
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "role_descriptors"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "api_key", func(value string) error {
						initialAPIKey = value

						if value == "" {
							return fmt.Errorf("expected api_key to be non-empty")
						}

						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_expiration"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "api_key", func(value string) error {
						if value != initialAPIKey {
							return fmt.Errorf("expected api_key to be unchanged")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyNonExpiring(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "type", "rest"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "expiration_timestamp", "0"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "key_id"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyExplicitConnection(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	endpoints := apiKeyESEndpoints()
	if len(endpoints) == 0 {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must contain at least one endpoint")
	}

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "elasticsearch_connection.0.insecure", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "key_id"),
				),
			},
		},
	})
}

func apiKeyESEndpoints() []string {
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

func checkResourceSecurityAPIKeyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_api_key" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		apiKey, diags := elasticsearch.GetAPIKey(context.Background(), client, compID.ResourceID)
		if diags.HasError() {
			return fmt.Errorf("Unable to get API key %v", diags)
		}

		if !apiKey.Invalidated {
			return fmt.Errorf("API key (%s) has not been invalidated", compID.ResourceID)
		}
	}
	return nil
}

func TestAccResourceSecurityAPIKeyCrossCluster(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersionWithCrossCluster, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "type", "cross_cluster"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.search.0.names.0", "logs-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.search.0.names.1", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.search.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.replication.0.names.0", "archive-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "metadata", `{"description":"Cross-cluster test key","environment":"test"}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration_timestamp"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "type", "cross_cluster"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.search.0.names.0", "log-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.search.0.names.1", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.search.0.allow_restricted_indices", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "access.replication.0.names.0", "archives-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "metadata", `{"description":"Cross-cluster test key updated","environment":"test"}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration_timestamp"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyWithDefaultAllowRestrictedIndices(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-default": {
								Cluster: []string{"monitor"},
								Indices: []models.IndexPerms{{
									Names:      []string{"logs-*", "metrics-*"},
									Privileges: []string{"read", "view_index_metadata"},
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("role descriptor mismatch:\nexpected: %+v\nactual: %+v", expectedRoleDescriptor, testRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyNoRoleDescriptors(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apikey.MinVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"api_key_name": config.StringVariable(apiKeyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "key_id"),
				),
			},
		},
	})
}
