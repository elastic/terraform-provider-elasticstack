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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
