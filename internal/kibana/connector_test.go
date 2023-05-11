package kibana_test

import (
	"context"
	"fmt"
	"regexp"
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
		user = "user1"
		password = "password1"
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

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentJson\":\"{}\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentResponseKey\":\"key\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentUrl\":\"https://www\.elastic\.co/\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentResponseExternalTitleKey\":\"title\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentUrl\":\"https://www\.elastic\.co/\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentJson\":\"{}\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentUrl\":\"https://www.elastic\.co/\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"viewIncidentUrl\":\"https://www\.elastic\.co/\"`)),
					// `post` is the default value that is returned by backend
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`"createIncidentMethod\":\"post\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"user\":\"user1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1\"`)),
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

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentJson\":\"{}\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentResponseKey\":\"key\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"createIncidentUrl\":\"https://www\.elastic\.co/\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentResponseExternalTitleKey\":\"title\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"getIncidentUrl\":\"https://www\.elastic\.co/\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentJson\":\"{}\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"updateIncidentUrl\":\"https://elasticsearch\.com/\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"viewIncidentUrl\":\"https://www\.elastic\.co/\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`createIncidentMethod\":\"put\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"user\":\"user2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password2\"`)),
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
		host = "localhost"
	  })
	  secrets = jsonencode({
		user = "user1"
		password = "password1"
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

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"from\":\"test@elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"port\":111`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"host\":\"localhost\"`)),
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

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"from\":\"test2@elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"port\":222`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"host\":\"localhost\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"user\":\"user1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1\"`)),
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

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":true`)),
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

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"index\":\"\.kibana\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"refresh\":false`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorJira(t *testing.T) {
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
		apiUrl = "url1"
		projectKey = "project1"
	  })
	  secrets = jsonencode({
		apiToken = "secret1"
		email = "email1"
	  })	
	  connector_type_id = ".jira"
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
		apiUrl = "url2"
		projectKey = "project2"
	  })
	  secrets = jsonencode({
		apiToken = "secret2"
		email = "email2"
	  })	
	  connector_type_id = ".jira"
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
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".jira"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"url1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"projectKey\":\"project1\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiToken\":\"secret1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"email\":\"email1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", fmt.Sprintf("Updated %s", connectorName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".jira"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"url2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"projectKey\":\"project2\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiToken\":\"secret2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"email\":\"email2\"`)),
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
