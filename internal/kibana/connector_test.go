package kibana_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceKibanaConnectorCasesWebhook(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.4.0"))

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
		user = "test"
		password = "test"
	  })
	  connector_type_id = ".cases-webhook"
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
		createIncidentJson = "{}"
		createIncidentResponseKey = "key"
		createIncidentUrl = "https://www.elastic.co/"
		getIncidentResponseExternalTitleKey = "title"
		getIncidentUrl = "https://www.elastic.co/"
		updateIncidentJson = "{}"
		updateIncidentUrl = "https://www.elastic.co/"
		viewIncidentUrl = "https://www.elastic.co/"
		createIncidentMethod = "put"
	  })
	  secrets = jsonencode({
		user = "test"
		password = "test"
	  })
	  connector_type_id = ".cases-webhook"
	}`,
			name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceKibanaConnectorDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", connectorName),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".cases-webhook"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", fmt.Sprintf("Updated %s", connectorName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".cases-webhook"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorEmail(t *testing.T) {
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
		from = "test@elastic.co"
		port = 111
		host = "localhost"
		  })
	  secrets = jsonencode({})
	  connector_type_id = ".email"
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
		from = "test2@elastic.co"
		port = 222
		host = "127.0.0.1"
	  })
	  secrets = jsonencode({
		user = "user"
		password = "password"
	  })
	  connector_type_id = ".email"
	}`,
			name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceKibanaConnectorDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", connectorName),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".email"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", fmt.Sprintf("Updated %s", connectorName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".email"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
				),
			},
		},
	})
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
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", connectorName),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", fmt.Sprintf("Updated %s", connectorName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),
				),
			},
		},
	})
}

func checkResourceKibanaConnectorDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_action_connector" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		connector, diags := kibana.GetConnector(context.Background(), client, compId.ResourceId, compId.ClusterId)
		if diags.HasError() {
			return fmt.Errorf("Failed to get connector: %v", diags)
		}

		if connector != nil {
			return fmt.Errorf("Action connector (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
