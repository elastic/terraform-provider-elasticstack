package security_enable_rule_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionEnableRule = version.Must(version.NewVersion("8.11.0"))

func TestAccResourceEnableRule(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule),
				Config:   testAccResourceEnableRuleBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", "test_tag"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", "terraform_test"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_enable_rule.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceEnableRuleWithManualDisable(t *testing.T) {
	tagKey := "test_tag"
	tagValue := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	spaceID := "default"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule),
				Config:   testAccResourceEnableRuleWithRules(tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", tagKey),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", tagValue),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue, true),
				),
			},
			{
				// Manually disable one rule outside of Terraform to test drift detection
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule),
				PreConfig: func() {
					disableOneRule(t, spaceID, tagKey, tagValue)
				},
				Config: testAccResourceEnableRuleWithRules(tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue, true),
				),
			},
		},
	})
}

func TestAccResourceEnableRuleDisableOnDestroyFalse(t *testing.T) {
	tagKey := "test_tag"
	tagValue := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	spaceID := "default"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule),
				Config:   testAccResourceEnableRuleDisableOnDestroyFalse(tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", tagKey),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", tagValue),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "disable_on_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue, true),
				),
			},
			{
				// Destroy the enable_rule resource but keep the rules
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule),
				Config:   testAccResourceEnableRuleDisableOnDestroyFalseRulesOnly(tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					// Verify rules are still enabled after destroying the enable_rule resource
					checkRulesEnabled(spaceID, tagKey, tagValue, true),
				),
			},
		},
	})
}

// testAccCreateDetectionRules creates two test detection rules with the specified tag
func testAccCreateDetectionRules(tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "elasticstack_kibana_security_detection_rule" "test_rule_1" {
  name        = "Test Rule 1 - %s"
  type        = "query"
  query       = "event.action:test"
  language    = "kuery"
  description = "Test rule for enable_rule resource"
  severity    = "low"
  risk_score  = 21
  index       = ["logs-*"]
  tags        = ["%s: %s", "test"]
  
  lifecycle {
    ignore_changes = [enabled]
  }
}

resource "elasticstack_kibana_security_detection_rule" "test_rule_2" {
  name        = "Test Rule 2 - %s"
  type        = "query"
  query       = "event.action:test2"
  language    = "kuery"
  description = "Test rule for enable_rule resource"
  severity    = "low"
  risk_score  = 21
  index       = ["logs-*"]
  tags        = ["%s: %s", "test"]
  
  lifecycle {
    ignore_changes = [enabled]
  }
}
`, tagValue, tagKey, tagValue, tagValue, tagKey, tagValue)
}

func testAccResourceEnableRuleBasic() string {
	return `
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_enable_rule" "test" {
  space_id = "default"
  key      = "test_tag"
  value    = "terraform_test"
}
`
}

func testAccResourceEnableRuleWithRules(tagKey, tagValue string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

%s

resource "elasticstack_kibana_security_enable_rule" "test" {
  space_id = "default"
  key      = "%s"
  value    = "%s"

  depends_on = [
    elasticstack_kibana_security_detection_rule.test_rule_1,
    elasticstack_kibana_security_detection_rule.test_rule_2
  ]
}
`, testAccCreateDetectionRules(tagKey, tagValue), tagKey, tagValue)
}

func testAccResourceEnableRuleDisableOnDestroyFalse(tagKey, tagValue string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

%s

resource "elasticstack_kibana_security_enable_rule" "test" {
  space_id          = "default"
  key               = "%s"
  value             = "%s"
  disable_on_destroy = false
  
  depends_on = [
    elasticstack_kibana_security_detection_rule.test_rule_1,
    elasticstack_kibana_security_detection_rule.test_rule_2
  ]
}
`, testAccCreateDetectionRules(tagKey, tagValue), tagKey, tagValue)
}

func testAccResourceEnableRuleDisableOnDestroyFalseRulesOnly(tagKey, tagValue string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

%s
`, testAccCreateDetectionRules(tagKey, tagValue))
}

// checkRulesEnabled verifies that all rules matching the tag are in the expected enabled state
func checkRulesEnabled(spaceID, key, value string, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		kbClient, err := client.GetKibanaOapiClient()
		if err != nil {
			return fmt.Errorf("failed to get Kibana client: %w", err)
		}

		ctx := context.Background()
		filter := fmt.Sprintf("alert.attributes.tags:(\"%s: %s\")", key, value)
		perPage := 100
		page := 1
		params := &kbapi.FindRulesParams{
			Filter:  &filter,
			Page:    &page,
			PerPage: &perPage,
		}

		resp, err := kbClient.API.FindRulesWithResponse(ctx, params, func(ctx context.Context, req *http.Request) error {
			if spaceID != "" && spaceID != "default" {
				req.URL.Path = fmt.Sprintf("/s/%s%s", spaceID, req.URL.Path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to query rules: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("failed to query rules, status: %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("empty response from FindRules")
		}

		if resp.JSON200.Total == 0 {
			return fmt.Errorf("no rules found matching tag %s: %s", key, value)
		}

		for _, ruleResp := range resp.JSON200.Data {
			queryRule, err := ruleResp.AsSecurityDetectionsAPIQueryRule()
			if err == nil {
				if queryRule.Enabled != expectedEnabled {
					return fmt.Errorf("rule has enabled=%v, expected %v", queryRule.Enabled, expectedEnabled)
				}
				continue
			}
		}

		return nil
	}
}

// disableOneRule manually disables one rule matching the tag (for testing drift detection)
func disableOneRule(t *testing.T, spaceID, key, value string) {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	kbClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Fatalf("failed to get Kibana client: %v", err)
	}

	ctx := context.Background()
	filter := fmt.Sprintf("alert.attributes.tags:(\"%s: %s\")", key, value)
	perPage := 1
	page := 1
	params := &kbapi.FindRulesParams{
		Filter:  &filter,
		Page:    &page,
		PerPage: &perPage,
	}

	resp, err := kbClient.API.FindRulesWithResponse(ctx, params, func(ctx context.Context, req *http.Request) error {
		if spaceID != "" && spaceID != "default" {
			req.URL.Path = fmt.Sprintf("/s/%s%s", spaceID, req.URL.Path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to query rules: %v", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || len(resp.JSON200.Data) == 0 {
		t.Fatalf("failed to find rules to disable")
	}

	queryRule, err := resp.JSON200.Data[0].AsSecurityDetectionsAPIQueryRule()
	if err != nil {
		t.Fatalf("failed to get rule ID from response")
	}

	bulkAction := kbapi.SecurityDetectionsAPIBulkDisableRules{
		Action: kbapi.Disable,
		Ids:    &[]string{queryRule.Id.String()},
	}

	bodyBytes, err := json.Marshal(bulkAction)
	if err != nil {
		t.Fatalf("failed to marshal bulk action: %v", err)
	}

	bulkResp, err := kbClient.API.PerformRulesBulkActionWithBodyWithResponse(ctx, &kbapi.PerformRulesBulkActionParams{}, "application/json", bytes.NewReader(bodyBytes), func(ctx context.Context, req *http.Request) error {
		if spaceID != "" && spaceID != "default" {
			req.URL.Path = fmt.Sprintf("/s/%s%s", spaceID, req.URL.Path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to disable rule: %v", err)
	}

	if bulkResp.StatusCode() != 200 {
		t.Fatalf("failed to disable rule, status: %d", bulkResp.StatusCode())
	}
}
