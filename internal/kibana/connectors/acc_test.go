package connectors_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

//go:embed testdata/TestAccResourceKibanaConnectorFromSDK/connector.tf
var sdkIndexConnectorConfig string

//go:embed testdata/TestAccResourceKibanaConnectorEmptyConfigFromSDK/connector.tf
var sdkSlackConnectorConfig string

func TestAccResourceKibanaConnectorCasesWebhook(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.4.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	testCases := []struct {
		name        string
		connectorID string
		minVersion  *version.Version
	}{
		{
			name:        "with_empty_connector_id",
			connectorID: "",
			minVersion:  minSupportedVersion,
		},
		{
			name:        "with_predefined_connector_id",
			connectorID: uuid.NewString(),
			minVersion:  connectors.MinVersionSupportingPreconfiguredIDs,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vars := config.Variables{
				"connector_name": config.StringVariable(connectorName),
			}
			if tc.connectorID != "" {
				vars["connector_id"] = config.StringVariable(tc.connectorID)
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				CheckDestroy:             checkResourceKibanaConnectorDestroy,
				ProtoV6ProviderFactories: acctest.Providers,
				Steps: []resource.TestStep{
					{
						SkipFunc:        versionutils.CheckIfVersionIsUnsupported(tc.minVersion),
						ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
						ConfigVariables: vars,
						Check: resource.ComposeTestCheckFunc(
							testCommonAttributes(connectorName, ".cases-webhook"),

							resource.TestCheckResourceAttrWith("elasticstack_kibana_action_connector.test", "connector_id", func(value string) error {
								if tc.connectorID == "" {
									if _, err := uuid.Parse(value); err != nil {
										return fmt.Errorf("expected connector_id to be a uuid: %w", err)
									}

									return nil
								}

								if tc.connectorID != value {
									return fmt.Errorf("expected connector_id to match pre-defined id. '%s' != %s", tc.connectorID, value)
								}

								return nil
							}),

							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentJson\":\"{}\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentResponseKey\":\"key\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentResponseExternalTitleKey\":\"title\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentJson\":\"{}\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentUrl\":\"https://www.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"viewIncidentUrl\":\"https://www\.elastic\.co/\"`)),

							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"user\":\"user1\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1\"`)),
						),
					},
					{
						SkipFunc:        versionutils.CheckIfVersionIsUnsupported(tc.minVersion),
						ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
						ConfigVariables: vars,
						Check: resource.ComposeTestCheckFunc(
							testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".cases-webhook"),

							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentJson\":\"{}\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentResponseKey\":\"key\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentResponseExternalTitleKey\":\"title\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentJson\":\"{}\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentUrl\":\"https://elasticsearch\.com/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"viewIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentMethod\":\"put\"`)),

							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"user\":\"user2\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password2\"`)),
						),
					},
					{
						SkipFunc:        versionutils.CheckIfVersionIsUnsupported(tc.minVersion),
						ConfigDirectory: acctest.NamedTestCaseDirectory("null_headers"),
						ConfigVariables: vars,
						Check: resource.ComposeTestCheckFunc(
							testCommonAttributes(connectorName, ".cases-webhook"),

							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentJson\":\"{}\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentResponseKey\":\"key\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentResponseExternalTitleKey\":\"title\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentJson\":\"{}\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentUrl\":\"https://www.elastic\.co/\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"viewIncidentUrl\":\"https://www\.elastic\.co/\"`)),
							// Verify that null headers field is removed from the config
							func(s *terraform.State) error {
								rs, ok := s.RootModule().Resources["elasticstack_kibana_action_connector.test"]
								if !ok {
									return fmt.Errorf("resource not found")
								}
								configStr := rs.Primary.Attributes["config"]
								if regexp.MustCompile(`\"headers\"`).MatchString(configStr) {
									return fmt.Errorf("headers field should not be present in config when null, got: %s", configStr)
								}
								return nil
							},

							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"user\":\"user1\"`)),
							resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1\"`)),
						),
					},
				},
			})
		})
	}
}

func testCommonAttributes(connectorName, connectorTypeID string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", connectorName),
		resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", connectorTypeID),
		resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
		resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
		resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
	)
}

func checkResourceKibanaConnectorDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_action_connector" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		connector, diags := kibana_oapi.GetConnector(context.Background(), oapiClient, compId.ResourceId, compId.ClusterId)
		if diags.HasError() {
			return fmt.Errorf("Failed to get connector: %v", diags)
		}

		if connector != nil {
			return fmt.Errorf("Action connector (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}

func TestAccResourceKibanaConnectorIndex(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceKibanaConnectorDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":true`)),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".index"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":false`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorFromSDK(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				// Create the connector with the last provider version where the connector resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.17",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   sdkIndexConnectorConfig,
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":true`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":true`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorEmptyConfigFromSDK(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				// Create the connector with the last provider version where the connector resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.17",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   sdkSlackConnectorConfig,
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".slack"),

					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "config", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".slack"),

					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "config", `{"__tf_provider_connector_type_id":".slack"}`),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorAI(t *testing.T) {
	testCases := []struct {
		name                string
		connectorTypeID     string
		minSupportedVersion *version.Version
		createChecks        []resource.TestCheckFunc
		updateChecks        []resource.TestCheckFunc
	}{
		{
			name:                "bedrock",
			connectorTypeID:     ".bedrock",
			minSupportedVersion: version.Must(version.NewSemver("8.16.2")),
			createChecks: []resource.TestCheckFunc{
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://bedrock-runtime\.us-east-1\.amazonaws\.com\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"anthropic\.claude-v2\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"accessKey\":\"test-access-key\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"secret\":\"test-secret-key\"`)),
			},
			updateChecks: []resource.TestCheckFunc{
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://bedrock-runtime\.us-west-2\.amazonaws\.com\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"anthropic\.claude-3-5-sonnet-20240620-v1:0\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"accessKey\":\"updated-access-key\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"secret\":\"updated-secret-key\"`)),
			},
		},
		{
			name:                "gen-ai",
			connectorTypeID:     ".gen-ai",
			minSupportedVersion: version.Must(version.NewSemver("8.10.3")),
			createChecks: []resource.TestCheckFunc{
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiProvider\":\"OpenAI\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://api\.openai\.com/v1\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"gpt-4\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKey\":\"test-api-key\"`)),
			},
			updateChecks: []resource.TestCheckFunc{
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiProvider\":\"OpenAI\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://api\.openai\.com/v1\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"gpt-4o\"`)),
				resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKey\":\"updated-api-key\"`)),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				CheckDestroy:             checkResourceKibanaConnectorDestroy,
				ProtoV6ProviderFactories: acctest.Providers,
				Steps: []resource.TestStep{
					{
						SkipFunc:        versionutils.CheckIfVersionIsUnsupported(tc.minSupportedVersion),
						ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
						ConfigVariables: config.Variables{
							"connector_name": config.StringVariable(connectorName),
						},
						Check: resource.ComposeTestCheckFunc(
							append([]resource.TestCheckFunc{testCommonAttributes(connectorName, tc.connectorTypeID)}, tc.createChecks...)...,
						),
					},
					{
						SkipFunc:        versionutils.CheckIfVersionIsUnsupported(tc.minSupportedVersion),
						ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
						ConfigVariables: config.Variables{
							"connector_name": config.StringVariable(connectorName),
						},
						Check: resource.ComposeTestCheckFunc(
							append([]resource.TestCheckFunc{testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), tc.connectorTypeID)}, tc.updateChecks...)...,
						),
					},
				},
			})
		})
	}
}
