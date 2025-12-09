package security_list_data_streams_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var MinListDataStreamsVersion = version.Must(version.NewVersion("7.10.0"))

func TestAccResourceSecurityListDataStreams(t *testing.T) {
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceListDataStreamsDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(MinListDataStreamsVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "list_index", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "list_item_index", "true"),
				),
			},
			{ // Import
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(MinListDataStreamsVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				ResourceName:      "elasticstack_kibana_security_list_data_streams.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecurityListDataStreamsWithSpace(t *testing.T) {
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceListDataStreamsDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(MinListDataStreamsVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_space"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "list_index", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_data_streams.test", "list_item_index", "true"),
				),
			},
			{ // Import
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(MinListDataStreamsVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_space"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				ResourceName:      "elasticstack_kibana_security_list_data_streams.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceListDataStreamsDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_security_list_data_streams" {
			continue
		}

		// The ID is the space_id
		spaceID := rs.Primary.ID

		// Check if the data streams still exist
		listIndex, listItemIndex, diags := kibana_oapi.ReadListIndex(context.Background(), oapiClient, spaceID)
		if diags.HasError() {
			// All errors must be 404s (resource doesn't exist), which is what we want
			// Any other error should fail the test
			for _, d := range diags {
				if d.Summary() != "Unexpected status code from server: got HTTP 404" {
					return fmt.Errorf("Unexpected error checking list data streams in space (%s): %s - %s", spaceID, d.Summary(), d.Detail())
				}
			}
			// All errors were 404s, which means destroyed - continue to next resource
			continue
		}

		if listIndex || listItemIndex {
			return fmt.Errorf("List data streams still exist in space (%s)", spaceID)
		}
	}
	return nil
}
