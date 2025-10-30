package system_user_test

import (
	_ "embed"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const remoteMonitoringUser = "remote_monitoring_user"

func TestAccResourceSecuritySystemUser(t *testing.T) {
	newPassword := "new_password"
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
					"password": config.StringVariable(newPassword),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
					checks.CheckUserCanAuthenticate(remoteMonitoringUser, newPassword),
				),
			},
		},
	})
}

func TestAccResourceSecuritySystemUserNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"username": config.StringVariable("not_system_user"),
					"password": config.StringVariable("new_password"),
				},
				ExpectError: regexp.MustCompile(`System user "not_system_user" not found`),
			},
		},
	})
}

//go:embed testdata/TestAccResourceSecuritySystemUserFromSDK/system_user.tf
var sdkCreateTestConfig string

func TestAccResourceSecuritySystemUserFromSDK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the system user with the last provider version where the system user resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.15",
					},
				},
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Config: sdkCreateTestConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(remoteMonitoringUser),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", remoteMonitoringUser),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
				),
			},
		},
	})
}
