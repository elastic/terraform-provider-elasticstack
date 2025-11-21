package connectors_test

import (
	"context"
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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceKibanaConnectorCasesWebhook(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.4.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	create := func(name, id string) string {
		idAttribute := ""
		if id != "" {
			idAttribute = fmt.Sprintf(`connector_id = "%s"`, id)
		}
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  %s
	  config       = jsonencode({
		createIncidentJson = "{}"
		createIncidentResponseKey = "key"
		createIncidentUrl = "https://www.elastic.co/"
		getIncidentResponseExternalTitleKey = "title"
		getIncidentUrl = "https://www.elastic.co/"
		updateIncidentJson = "{}"
		updateIncidentUrl = "https://www.elastic.co/"
		viewIncidentUrl = "https://www.elastic.co/"
	  })
	  secrets = jsonencode({
		user = "user1"
		password = "password1"
	  })
	  connector_type_id = ".cases-webhook"
	}`,
			name, idAttribute)
	}

	update := func(name, id string) string {
		idAttribute := ""
		if id != "" {
			idAttribute = fmt.Sprintf(`connector_id = "%s"`, id)
		}
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "Updated %s"
	  %s
	  config = jsonencode({
		createIncidentJson = "{}"
		createIncidentResponseKey = "key"
		createIncidentUrl = "https://www.elastic.co/"
		getIncidentResponseExternalTitleKey = "title"
		getIncidentUrl = "https://www.elastic.co/"
		updateIncidentJson = "{}"
		updateIncidentUrl = "https://elasticsearch.com/"
		viewIncidentUrl = "https://www.elastic.co/"
		createIncidentMethod = "put"
	  })
	  secrets = jsonencode({
		user = "user2"
		password = "password2"
	  })
	  connector_type_id = ".cases-webhook"
	}`,
			name, idAttribute)
	}

	for _, connectorID := range []string{"", uuid.NewString()} {
		t.Run(fmt.Sprintf("with connector ID '%s'", connectorID), func(t *testing.T) {
			minVersion := minSupportedVersion
			if connectorID != "" {
				minVersion = connectors.MinVersionSupportingPreconfiguredIDs
			}
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				CheckDestroy:             checkResourceKibanaConnectorDestroy,
				ProtoV6ProviderFactories: acctest.Providers,
				Steps: []resource.TestStep{
					{
						SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersion),
						Config:   create(connectorName, connectorID),
						Check: resource.ComposeTestCheckFunc(
							testCommonAttributes(connectorName, ".cases-webhook"),

							resource.TestCheckResourceAttrWith("elasticstack_kibana_action_connector.test", "connector_id", func(value string) error {
								if connectorID == "" {
									if _, err := uuid.Parse(value); err != nil {
										return fmt.Errorf("expected connector_id to be a uuid: %w", err)
									}

									return nil
								}

								if connectorID != value {
									return fmt.Errorf("expected connector_id to match pre-defined id. '%s' != %s", connectorID, value)
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
						SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersion),
						Config:   update(connectorName, connectorID),
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

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  config       = jsonencode({
		index             = ".kibana"
		refresh             = true
	  })
	  connector_type_id = ".index"
	}`,
			name)
	}

	update := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "Updated %s"
	  config       = jsonencode({
		index             = ".kibana"
		refresh             = false
	  })
	  connector_type_id = ".index"
	}`,
			name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceKibanaConnectorDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":true`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
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

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  config       = jsonencode({
		index             = ".kibana"
		refresh             = true
	  })
	  connector_type_id = ".index"
	}`,
			name)
	}

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
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".index"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":true`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:                   create(connectorName),
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

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
      name              = "%s"
      connector_type_id = ".slack"
      secrets = jsonencode({
        webhookUrl = "https://example.com/webhook"
      })
    }`,
			name)
	}

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
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".slack"),

					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "config", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:                   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".slack"),

					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "config", `{"__tf_provider_connector_type_id":".slack"}`),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorBedrock(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.16.2"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  config       = jsonencode({
		apiUrl       = "https://bedrock-runtime.us-east-1.amazonaws.com"
		defaultModel = "anthropic.claude-v2"
	  })
	  secrets = jsonencode({
		accessKey = "test-access-key"
		secret    = "test-secret-key"
	  })
	  connector_type_id = ".bedrock"
	}`,
			name)
	}

	update := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "Updated %s"
	  config = jsonencode({
		apiUrl       = "https://bedrock-runtime.us-west-2.amazonaws.com"
		defaultModel = "anthropic.claude-3-5-sonnet-20240620-v1:0"
	  })
	  secrets = jsonencode({
		accessKey = "updated-access-key"
		secret    = "updated-secret-key"
	  })
	  connector_type_id = ".bedrock"
	}`,
			name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceKibanaConnectorDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".bedrock"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://bedrock-runtime\.us-east-1\.amazonaws\.com\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"anthropic\.claude-v2\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"accessKey\":\"test-access-key\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"secret\":\"test-secret-key\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".bedrock"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://bedrock-runtime\.us-west-2\.amazonaws\.com\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"anthropic\.claude-3-5-sonnet-20240620-v1:0\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"accessKey\":\"updated-access-key\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"secret\":\"updated-secret-key\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorGenAi(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.8.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  config       = jsonencode({
		apiProvider  = "OpenAI"
		apiUrl       = "https://api.openai.com/v1"
		defaultModel = "gpt-4"
	  })
	  secrets = jsonencode({
		apiKey = "test-api-key"
	  })
	  connector_type_id = ".gen-ai"
	}`,
			name)
	}

	update := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "Updated %s"
	  config = jsonencode({
		apiProvider  = "OpenAI"
		apiUrl       = "https://api.openai.com/v1"
		defaultModel = "gpt-4o"
	  })
	  secrets = jsonencode({
		apiKey = "updated-api-key"
	  })
	  connector_type_id = ".gen-ai"
	}`,
			name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceKibanaConnectorDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".gen-ai"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiProvider\":\"OpenAI\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://api\.openai\.com/v1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"gpt-4\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKey\":\"test-api-key\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".gen-ai"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiProvider\":\"OpenAI\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://api\.openai\.com/v1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"gpt-4o\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKey\":\"updated-api-key\"`)),
				),
			},
		},
	})
}
