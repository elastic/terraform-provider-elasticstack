package prebuilt_rules_test

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var minVersionPrebuiltRules = version.Must(version.NewVersion("8.0.0"))

func TestAccResourcePrebuiltRules(t *testing.T) {
	testAccResourcePrebuiltRules(t, "default")
}

func TestAccResourcePrebuiltRulesInSpace(t *testing.T) {
	spaceID := "security_rules" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	testAccResourcePrebuiltRules(t, spaceID)
}

func testAccResourcePrebuiltRules(t *testing.T, spaceID string) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionPrebuiltRules),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "space_id", spaceID),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_updated"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_updated"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionPrebuiltRules),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				PreConfig: func() {
					deleteSingleDetectionRule(t, spaceID)
				},
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minVersionPrebuiltRules),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectKnownValue("elasticstack_kibana_install_prebuilt_rules.test", tfjsonpath.New("rules_not_installed"), knownvalue.Int64Exact(0)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "space_id", spaceID),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_updated"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_updated"),
				),
			},
		},
	})
}

func deleteSingleDetectionRule(t *testing.T, spaceID string) {
	client, err := clients.NewAcceptanceTestingClient()
	require.NoError(t, err)

	oapiClient, err := client.GetKibanaOapiClient()
	require.NoError(t, err)

	resp, err := oapiClient.API.FindRulesWithResponse(t.Context(), &kbapi.FindRulesParams{}, kibana_oapi.SpaceAwarePathRequestEditor(spaceID))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())

	ruleBytes, err := resp.JSON200.Data[0].MarshalJSON()
	require.NoError(t, err)

	var ruleMap map[string]interface{}
	err = json.Unmarshal(ruleBytes, &ruleMap)
	require.NoError(t, err)

	id, ok := ruleMap["id"].(string)
	require.True(t, ok, "rule ID not found or not a string")

	idUUID, err := uuid.Parse(id)
	require.NoError(t, err)

	deleteResp, err := oapiClient.API.DeleteRuleWithResponse(t.Context(), spaceID, &kbapi.DeleteRuleParams{Id: &idUUID})
	require.NoError(t, err)
	require.Equal(t, 200, deleteResp.StatusCode())
}
