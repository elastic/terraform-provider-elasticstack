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
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
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
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".cases-webhook"),

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
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".email"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"from\":\"test@elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"port\":111`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"host\":\"localhost\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".email"),

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

func TestAccResourceKibanaConnectorGemini(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.15.0"))

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
      apiUrl       = "https://elastic.co",
      gcpRegion    = "us-central1",
      gcpProjectID = "project1",
      defaultModel = "gemini-1.5-pro-001"
    })
	  secrets = jsonencode({
      credentialsJson = "secret1"
    })
	  connector_type_id = ".gemini"
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
      apiUrl       = "https://elasticsearch.com",
      gcpRegion    = "us-east4",
      gcpProjectID = "project2",
      defaultModel = "gemini-1.5-pro-001"
	  })
	  secrets = jsonencode({
      credentialsJson = "secret2"
	  })
	  connector_type_id = ".gemini"
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
					testCommonAttributes(connectorName, ".gemini"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"gcpRegion\":\"us-central1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"gcpProjectID\":\"project1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"gemini-1.5-pro-001\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"credentialsJson\":\"secret1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".gemini"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"gcpRegion\":\"us-east4\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"gcpProjectID\":\"project2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"defaultModel\":\"gemini-1.5-pro-001\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"credentialsJson\":\"secret2\"`)),
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
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   create(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".jira"),

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
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".jira"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"url2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"projectKey\":\"project2\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiToken\":\"secret2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"email\":\"email2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorOpsgenie(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.6.0"))

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
		apiUrl = "https://elastic.co"
	  })
	  secrets = jsonencode({
		apiKey = "key1"
	  })	
	  connector_type_id = ".opsgenie"
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
		apiUrl = "https://elasticsearch.com"
	  })
	  secrets = jsonencode({
		apiKey = "key2"
	  })	
	  connector_type_id = ".opsgenie"
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
					testCommonAttributes(connectorName, ".opsgenie"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKey\":\"key1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".opsgenie"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKey\":\"key2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorPagerduty(t *testing.T) {
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
   		apiUrl = "https://elastic.co"
	  })
	  secrets = jsonencode({
		routingKey = "test1"
	  })	
	  connector_type_id = ".pagerduty"
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
   		apiUrl = "https://elasticsearch.com"
	})
	secrets = jsonencode({
	  routingKey = "test2"
	})	
	connector_type_id = ".pagerduty"
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
					testCommonAttributes(connectorName, ".pagerduty"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"routingKey\":\"test1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".pagerduty"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"routingKey\":\"test2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorResilient(t *testing.T) {
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
		apiUrl = "https://elastic.co"
		orgId = "id1"
	  })
	  secrets = jsonencode({
		apiKeyId = "key1"
		apiKeySecret = "secret1"
	  })	
	  connector_type_id = ".resilient"
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
		apiUrl = "https://elasticsearch.com"
		orgId = "id2"
	  })
	  secrets = jsonencode({
		apiKeyId = "key2"
		apiKeySecret = "secret2"
	  })
	  connector_type_id = ".resilient"
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
					testCommonAttributes(connectorName, ".resilient"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"orgId\":\"id1\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKeyId\":\"key1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKeySecret\":\"secret1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".resilient"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"orgId\":\"id2\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKeyId\":\"key2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiKeySecret\":\"secret2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorServerLog(t *testing.T) {
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
	  connector_type_id = ".server-log"
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
	  connector_type_id = ".server-log"
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
					testCommonAttributes(connectorName, ".server-log"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".server-log"),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorServicenow(t *testing.T) {
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
	  config = jsonencode({
		apiUrl = "https://elastic.co"
	  })
	  secrets = jsonencode({
		username = "user1"
		password = "password1"
	  })
	  connector_type_id = ".servicenow"
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
		apiUrl = "https://elasticsearch.com"

	  })
	  secrets = jsonencode({
		username = "user2"
		password = "password2"
	  })
	  connector_type_id = ".servicenow"
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
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", connectorName),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".servicenow"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"username\":\"user1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", fmt.Sprintf("Updated %s", connectorName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".servicenow"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"username\":\"user2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorServicenowItom(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.3.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  config = jsonencode({
		apiUrl = "https://elastic.co"
	  })
	  secrets = jsonencode({
		username = "user1"
		password = "password1"
	  })
	  connector_type_id = ".servicenow-itom"
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
		apiUrl = "https://elasticsearch.com"
	  })
	  secrets = jsonencode({
		username = "user2"
		password = "password2"
	  })
	  connector_type_id = ".servicenow-itom"
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
					testCommonAttributes(connectorName, ".servicenow-itom"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"username\":\"user1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".servicenow-itom"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"username\":\"user2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorServicenowSir(t *testing.T) {
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
	  config = jsonencode({
		apiUrl = "https://elastic.co"
	  })
	  secrets = jsonencode({
		username = "user1"
		password = "password1"
	  })
	  connector_type_id = ".servicenow-sir"
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
		apiUrl = "https://elasticsearch.com"
	  })
	  secrets = jsonencode({
		username = "user2"
		password = "password2"
	  })
	  connector_type_id = ".servicenow-sir"
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
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", connectorName),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".servicenow-sir"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"username\":\"user1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "name", fmt.Sprintf("Updated %s", connectorName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "connector_type_id", ".servicenow-sir"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test", "is_preconfigured", "false"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"username\":\"user2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorSlack(t *testing.T) {
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
	  secrets = jsonencode({
		webhookUrl = "https://elastic.co"
	  })
	  connector_type_id = ".slack"
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
	  secrets = jsonencode({
		webhookUrl = "https://elasticsearch.com"
	  })
	  connector_type_id = ".slack"
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
					testCommonAttributes(connectorName, ".slack"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"webhookUrl\":\"https://elastic\.co\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".slack"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"webhookUrl\":\"https://elasticsearch\.com\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorSlackApi(t *testing.T) {
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
	  secrets = jsonencode({
		token = "my-token"
	  })
	  connector_type_id = ".slack_api"
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
	  secrets = jsonencode({
		token = "my-updated-token"
	  })
	  connector_type_id = ".slack_api"
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
					testCommonAttributes(connectorName, ".slack_api"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"token\":\"my-token\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".slack_api"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"token\":\"my-updated-token\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorSwimlane(t *testing.T) {
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
	  config = jsonencode({
		apiUrl = "https://elastic.co"
		appId = "test1"
		connectorType = "all"
		mappings = {
		  alertIdConfig = {
			fieldType = "type1"
			id = "id1"
			key = "key1"
			name = "name1"
		  }
		}
	  })
	  secrets = jsonencode({
		apiToken = "token1"
	  })
	  connector_type_id = ".swimlane"
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
		apiUrl = "https://elasticsearch.com"
		appId = "test2"
		connectorType = "all"
		mappings = {
		  alertIdConfig = {
			fieldType = "type2"
			id = "id2"
			key = "key2"
			name = "name2"
		  }
		}
	  })
	  secrets = jsonencode({
		apiToken = "token2"
	  })
	  connector_type_id = ".swimlane"
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
					testCommonAttributes(connectorName, ".swimlane"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"appId\":\"test1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"connectorType\":\"all\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"fieldType\":\"type1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"id\":\"id1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"key\":\"key1\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"name\":\"name1\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiToken\":\"token1\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".swimlane"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"apiUrl\":\"https://elasticsearch\.com\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"appId\":\"test2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"connectorType\":\"all\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"fieldType\":\"type2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"id\":\"id2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"key\":\"key2\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"name\":\"name2\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"apiToken\":\"token2\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorTeams(t *testing.T) {
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
	  secrets = jsonencode({
		webhookUrl = "https://elastic.co"
	  })
	  connector_type_id = ".teams"
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
	  secrets = jsonencode({
		webhookUrl = "https://elasticsearch.com"
	  })
	  connector_type_id = ".teams"
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
					testCommonAttributes(connectorName, ".teams"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"webhookUrl\":\"https://elastic\.co\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".teams"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"webhookUrl\":\"https://elasticsearch\.com\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorTines(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.6.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  config = jsonencode({
		url = "https://elastic.co"
	  })
	  secrets = jsonencode({
		email = "test@elastic.co"
		token = "token1"
	  })
	  connector_type_id = ".tines"
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
		url = "https://elasticsearch.com"
	  })
	  secrets = jsonencode({
		email = "test@elasticsearch.com"
		token = "token2"
	  })
	  connector_type_id = ".tines"
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
					testCommonAttributes(connectorName, ".tines"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"url\":\"https://elastic\.co\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"email\":\"test@elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"token\":\"token1"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".tines"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"url\":\"https://elasticsearch\.com\"`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"email\":\"test@elasticsearch\.com\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"token\":\"token2"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorWebhook(t *testing.T) {
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
	  config = jsonencode({
		url = "https://elastic.co"
  		hasAuth = true
  		method = "post"
	  })
	  secrets = jsonencode({})
	  connector_type_id = ".webhook"
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
		url = "https://elasticsearch.com"
		hasAuth = true
		method = "post"
	  })
	  secrets = jsonencode({})
	  connector_type_id = ".webhook"
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
					testCommonAttributes(connectorName, ".webhook"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"url\":\"https://elastic\.co\"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".webhook"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"url\":\"https://elasticsearch\.com\"`)),
				),
			},
		},
	})
}

func TestAccResourceKibanaConnectorXmatters(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.2.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	create := func(name string) string {
		return fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}

	resource "elasticstack_kibana_action_connector" "test" {
	  name         = "%s"
	  config = jsonencode({
		configUrl = "https://elastic.co"
		usesBasic = true
	  })
	  secrets = jsonencode({
		user = "user1"
		password = "password1"
	  })
	  connector_type_id = ".xmatters"
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
		usesBasic = false
	  })
	  secrets = jsonencode({
		secretsUrl = "https://elasticsearch.com"
	  })
	  connector_type_id = ".xmatters"
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
					testCommonAttributes(connectorName, ".xmatters"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"configUrl\":\"https://elastic\.co\"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"usesBasic\":true`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"user\":\"user1"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"password\":\"password1"`)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   update(connectorName),
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(fmt.Sprintf("Updated %s", connectorName), ".xmatters"),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "config", regexp.MustCompile(`\"usesBasic\":false`)),

					resource.TestMatchResourceAttr("elasticstack_kibana_action_connector.test", "secrets", regexp.MustCompile(`\"secretsUrl\":\"https://elasticsearch\.com\"`)),
				),
			},
		},
	})
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
