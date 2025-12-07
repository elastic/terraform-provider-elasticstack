package user_test

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSecurityUser(t *testing.T) {
	// generate a random username
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityUserDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
					"role":     config.StringVariable("kibana_user"),
				},
				Check: resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", "test@example.com"),
			},
		},
	})
}

func TestAccImportedUserDoesNotResetPassword(t *testing.T) {
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	initialPassword := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	userUpdatedPassword := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityUserDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_password"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
				},
				SkipFunc: func() (bool, error) {
					client, err := clients.NewAcceptanceTestingClient()
					if err != nil {
						return false, err
					}
					body := fmt.Sprintf("{\"roles\": [\"kibana_admin\"], \"password\": \"%s\"}", initialPassword)

					esClient, err := client.GetESClient()
					if err != nil {
						return false, err
					}
					resp, err := esClient.Security.PutUser(username, strings.NewReader(body))
					if err != nil {
						return false, err
					}

					defer resp.Body.Close()

					if resp.IsError() {
						body, err := io.ReadAll(resp.Body)
						return false, fmt.Errorf("failed to manually create import test user [%s] %s %s", username, body, err)
					}
					return false, err
				},
				ResourceName: "elasticstack_elasticsearch_security_user.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					client, err := clients.NewAcceptanceTestingClient()
					if err != nil {
						return "", err
					}
					clusterId, diag := client.ClusterID(context.Background())
					if diag.HasError() {
						return "", fmt.Errorf("failed to get cluster uuid: %s", diag[0].Summary)
					}

					return fmt.Sprintf("%s/%s", *clusterId, username), nil
				},
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					importedUsername := is[0].Attributes["username"]
					if importedUsername != username {
						return fmt.Errorf("expected imported username [%s] to equal [%s]", importedUsername, username)
					}

					if _, ok := is[0].Attributes["password"]; ok {
						return fmt.Errorf("expected imported user to not have a password set - got [%s]", is[0].Attributes["password"])
					}

					return nil
				},
				Check: checks.CheckUserCanAuthenticate(username, initialPassword),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_password"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_admin"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_user.test", "password"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "full_name", "Test User"),
					checks.CheckUserCanAuthenticate(username, initialPassword),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
					"role":     config.StringVariable("kibana_admin"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", "test@example.com"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_user.test", "password"),
				),
			},
			{
				SkipFunc: func() (bool, error) {
					client, err := clients.NewAcceptanceTestingClient()
					if err != nil {
						return false, err
					}
					esClient, err := client.GetESClient()
					if err != nil {
						return false, err
					}
					body := fmt.Sprintf("{\"password\": \"%s\"}", userUpdatedPassword)

					req := esClient.Security.ChangePassword.WithUsername(username)
					resp, err := esClient.Security.ChangePassword(strings.NewReader(body), req)
					if err != nil {
						return false, nil
					}

					defer resp.Body.Close()

					if resp.IsError() {
						body, err := io.ReadAll(resp.Body)
						return false, fmt.Errorf("failed to manually change import test user password [%s] %s %s", username, body, err)
					}
					return false, err
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
					"role":     config.StringVariable("kibana_user"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", "test@example.com"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_user.test", "password"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
					checks.CheckUserCanAuthenticate(username, userUpdatedPassword),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceSecurityUserFromSDK/create/user.tf
var sdkCreateConfig string

func TestAccResourceSecurityUserFromSDK(t *testing.T) {
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the user with the last provider version where the user resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.12.1",
					},
				},
				Config: sdkCreateConfig,
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"username": config.StringVariable(username),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
				),
			},
		},
	})
}

func TestAccResourceSecurityUserWithPasswordWo(t *testing.T) {
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	password1 := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	password2 := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityUserDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"username":         config.StringVariable(username),
					"password":         config.StringVariable(password1),
					"password_version": config.StringVariable("v1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "password_wo_version", "v1"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_user.test", "password"),
					checks.CheckUserCanAuthenticate(username, password1),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"username":         config.StringVariable(username),
					"password":         config.StringVariable(password2),
					"password_version": config.StringVariable("v2"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "password_wo_version", "v2"),
					checks.CheckUserCanAuthenticate(username, password2),
				),
			},
		},
	})
}

func checkResourceSecurityUserDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_user" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Security.GetUser.WithUsername(compId.ResourceId)
		res, err := esClient.Security.GetUser(req)
		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.StatusCode != 404 {
			return fmt.Errorf("User (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
