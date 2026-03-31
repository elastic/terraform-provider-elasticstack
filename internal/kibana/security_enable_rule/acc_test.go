// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package securityenablerule_test

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
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionEnableRule = version.Must(version.NewVersion("8.11.0"))

const defaultSpaceID = "default"

func TestAccResourceEnableRule(t *testing.T) {
	// Skip entire test if version is below 8.11.0
	skipFunc := versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skip("Test requires version 8.11.0 or higher")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", defaultSpaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", "test_tag"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", "terraform_test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "disable_on_destroy", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_enable_rule.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceEnableRuleDefaultSpaceID(t *testing.T) {
	// Skip entire test if version is below 8.11.0
	skipFunc := versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skip("Test requires version 8.11.0 or higher")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", defaultSpaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", "test_tag"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", "terraform_test_default_space"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "disable_on_destroy", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_enable_rule.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceEnableRuleWithManualDisable(t *testing.T) {
	// Skip entire test if version is below 8.11.0
	skipFunc := versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skip("Test requires version 8.11.0 or higher")
	}

	tagKey := "test_tag"
	tagValue := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	spaceID := defaultSpaceID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"tag_key":   config.StringVariable(tagKey),
					"tag_value": config.StringVariable(tagValue),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", tagKey),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", tagValue),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue),
				),
			},
			{
				// Manually disable one rule outside of Terraform to test drift detection
				PreConfig: func() {
					disableOneRule(t, spaceID, tagKey, tagValue)
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"tag_key":   config.StringVariable(tagKey),
					"tag_value": config.StringVariable(tagValue),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue),
				),
			},
		},
	})
}

func TestAccResourceEnableRuleDisableOnDestroyFalse(t *testing.T) {
	// Skip entire test if version is below 8.11.0
	skipFunc := versionutils.CheckIfVersionIsUnsupported(minVersionEnableRule)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skip("Test requires version 8.11.0 or higher")
	}

	tagKey := "test_tag"
	tagValue := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	spaceID := defaultSpaceID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("disable_on_destroy_false"),
				ConfigVariables: config.Variables{
					"tag_key":   config.StringVariable(tagKey),
					"tag_value": config.StringVariable(tagValue),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", tagKey),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", tagValue),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "disable_on_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("disable_on_destroy_true"),
				ConfigVariables: config.Variables{
					"tag_key":   config.StringVariable(tagKey),
					"tag_value": config.StringVariable(tagValue),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", tagKey),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", tagValue),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "disable_on_destroy", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory: acctest.NamedTestCaseDirectory("disable_on_destroy_false"),
				ConfigVariables: config.Variables{
					"tag_key":   config.StringVariable(tagKey),
					"tag_value": config.StringVariable(tagValue),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "key", tagKey),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "value", tagValue),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "disable_on_destroy", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_enable_rule.test", "all_rules_enabled", "true"),
					checkRulesEnabled(spaceID, tagKey, tagValue),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				// Destroy the enable_rule resource but keep the rules
				ConfigDirectory: acctest.NamedTestCaseDirectory("rules_only"),
				ConfigVariables: config.Variables{
					"tag_key":   config.StringVariable(tagKey),
					"tag_value": config.StringVariable(tagValue),
				},
				Check: resource.ComposeTestCheckFunc(
					// Verify rules are still enabled after destroying the enable_rule resource
					checkRulesEnabled(spaceID, tagKey, tagValue),
				),
			},
		},
	})
}

// checkRulesEnabled verifies that all rules matching the tag are in the expected enabled state
func checkRulesEnabled(spaceID, key, value string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
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

		resp, err := kbClient.API.FindRulesWithResponse(ctx, params, func(_ context.Context, req *http.Request) error {
			if spaceID != "" && spaceID != defaultSpaceID {
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
				if !queryRule.Enabled {
					return fmt.Errorf("rule has enabled=%v, expected %v", queryRule.Enabled, true)
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

	resp, err := kbClient.API.FindRulesWithResponse(ctx, params, func(_ context.Context, req *http.Request) error {
		if spaceID != "" && spaceID != defaultSpaceID {
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

	bulkResp, err := kbClient.API.PerformRulesBulkActionWithBodyWithResponse(
		ctx,
		&kbapi.PerformRulesBulkActionParams{},
		"application/json",
		bytes.NewReader(bodyBytes),
		func(_ context.Context, req *http.Request) error {
			if spaceID != "" && spaceID != defaultSpaceID {
				req.URL.Path = fmt.Sprintf("/s/%s%s", spaceID, req.URL.Path)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to disable rule: %v", err)
	}

	if bulkResp.StatusCode() != 200 {
		t.Fatalf("failed to disable rule, status: %d", bulkResp.StatusCode())
	}
}
