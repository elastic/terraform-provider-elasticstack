package security_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSecurityUser(t *testing.T) {
	// generate a random username
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceSecurityUserDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityUserCreate(username),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", ""),
				),
			},
			{
				Config: testAccResourceSecurityUpdate(username, "kibana_user"),
				Check:  resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", "test@example.com"),
			},
		},
	})
}

func TestAccImportedUserDoesNotResetPassword(t *testing.T) {
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	initialPassword := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	userUpdatedPassword := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceSecurityUserDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityUserUpdateNoPassword(username),
				SkipFunc: func() (bool, error) {
					client, err := clients.NewAcceptanceTestingClient()
					if err != nil {
						return false, err
					}
					body := fmt.Sprintf("{\"roles\": [\"kibana_admin\"], \"password\": \"%s\"}", initialPassword)

					resp, err := client.GetESClient().Security.PutUser(username, strings.NewReader(body))
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
				Check: checkUserCanAuthenticate(username, initialPassword),
			},
			{
				Config: testAccResourceSecurityUserUpdateNoPassword(username),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_admin"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_user.test", "password"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "full_name", "Test User"),
					checkUserCanAuthenticate(username, initialPassword),
				),
			},
			{
				Config: testAccResourceSecurityUpdate(username, "kibana_admin"),
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
					esClient := client.GetESClient()
					body := fmt.Sprintf("{\"password\": \"%s\"}", userUpdatedPassword)

					req := esClient.API.Security.ChangePassword.WithUsername(username)
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
				Config: testAccResourceSecurityUpdate(username, "kibana_user"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", "test@example.com"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_user.test", "password"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
					checkUserCanAuthenticate(username, userUpdatedPassword),
				),
			},
		},
	})
}

func checkUserCanAuthenticate(username string, password string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}
		esClient := client.GetESClient()
		credentials := fmt.Sprintf("%s:%s", username, password)
		authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(credentials)))

		req := esClient.API.Security.Authenticate.WithHeader(map[string]string{"Authorization": authHeader})
		resp, err := esClient.Security.Authenticate(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.IsError() {
			body, err := io.ReadAll(resp.Body)

			return fmt.Errorf("failed to authenticate as test user [%s] %s %s", username, body, err)
		}
		return nil
	}
}

func testAccResourceSecurityUserCreate(username string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = "%s"
  roles     = ["kibana_user"]
  full_name = "Test User"
  password  = "qwerty123"
}
	`, username)
}

func testAccResourceSecurityUserUpdateNoPassword(username string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = "%s"
  roles     = ["kibana_admin"]
  full_name = "Test User"
}
	`, username)
}

func testAccResourceSecurityUpdate(username string, role string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = "%s"
  roles     = ["%s"]
  full_name = "Test User"
  email     = "test@example.com"
  password  = "qwerty123"
}
	`, username, role)
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

		req := client.GetESClient().Security.GetUser.WithUsername(compId.ResourceId)
		res, err := client.GetESClient().Security.GetUser(req)
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
