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

package watch_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	watchTriggerCreateExpected   = `{"schedule":{"cron":"0 0/1 * * * ?"}}`
	watchTriggerUpdateExpected   = `{"schedule":{"cron":"0 0/2 * * * ?"}}`
	watchInputNoneExpected       = `{"none":{}}`
	watchConditionAlways         = `{"always":{}}`
	watchActionsEmpty            = `{}`
	watchMetadataEmpty           = `{}`
	watchInputSimpleExpected     = `{"simple":{"name":"example"}}`
	watchInputSecondExpected     = `{"simple":{"count":2,"environment":"staging"}}`
	watchConditionNever          = `{"never":{}}`
	watchConditionScriptExpected = `{"script":{"lang":"painless","source":"return true"}}`
	watchActionsLogExpected      = `{"log":{"logging":{"level":"info","text":"example logging text"}}}`
	watchMetadataExample         = `{"example_key":"example_value"}`
	watchMetadataSecondExpected  = `{"env":"staging","priority":2}`
	watchTransformExpected       = `{"search":{"request":{"body":{"query":{"match_all":{}}},"indices":[],"rest_total_hits_as_int":true,` +
		`"search_type":"query_then_fetch"}}}`
	watchTransformScriptExpected = `{"script":{"lang":"painless","source":"return ctx.payload"}}`

	watchResourceName = "elasticstack_elasticsearch_watch.test"
)

//go:embed testdata/TestAccResourceWatchFromSDK/upgrade/main.tf
var watchFromSDKCreateConfig string

// canonicalJSONBytes re-encodes JSON so semantically equivalent documents compare equal (key order, spacing).
func canonicalJSONBytes(raw string) ([]byte, error) {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func testCheckWatchTransformSemanticallyEqual(t *testing.T, expected string) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[watchResourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", watchResourceName)
		}
		got, ok := rs.Primary.Attributes["transform"]
		if !ok {
			return fmt.Errorf("transform not found in state for %s", watchResourceName)
		}
		wantCanon, err := canonicalJSONBytes(expected)
		if err != nil {
			return fmt.Errorf("canonical expected transform: %w", err)
		}
		gotCanon, err := canonicalJSONBytes(got)
		if err != nil {
			return fmt.Errorf("canonical actual transform: %w", err)
		}
		if !bytes.Equal(wantCanon, gotCanon) {
			return fmt.Errorf("transform JSON mismatch (semantic)\nwant: %s\ngot:  %s", string(wantCanon), string(gotCanon))
		}
		return nil
	}
}

func testCheckWatchTransformCleared(t *testing.T) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[watchResourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", watchResourceName)
		}
		got, ok := rs.Primary.Attributes["transform"]
		if !ok || got == "" {
			return nil
		}
		return fmt.Errorf("transform should be cleared in state for %s, got %q", watchResourceName, got)
	}
}

// testCheckWatchAttrSemanticallyEqual checks that the named attribute in the
// watch resource state is semantically equal to the expected JSON string,
// ignoring key order and insignificant whitespace.
func testCheckWatchAttrSemanticallyEqual(t *testing.T, attr, expected string) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[watchResourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", watchResourceName)
		}
		got, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("%s not found in state for %s", attr, watchResourceName)
		}
		wantCanon, err := canonicalJSONBytes(expected)
		if err != nil {
			return fmt.Errorf("canonical expected %s: %w", attr, err)
		}
		gotCanon, err := canonicalJSONBytes(got)
		if err != nil {
			return fmt.Errorf("canonical actual %s: %w", attr, err)
		}
		if !bytes.Equal(wantCanon, gotCanon) {
			return fmt.Errorf("%s JSON mismatch (semantic)\nwant: %s\ngot:  %s", attr, string(wantCanon), string(gotCanon))
		}
		return nil
	}
}

func TestResourceWatch(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceWatchDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(watchResourceName, "id"),
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestCheckResourceAttr(watchResourceName, "active", "false"),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerCreateExpected),
					resource.TestCheckResourceAttr(watchResourceName, "input", watchInputNoneExpected),
					resource.TestCheckResourceAttr(watchResourceName, "condition", watchConditionAlways),
					resource.TestCheckResourceAttr(watchResourceName, "actions", watchActionsEmpty),
					resource.TestCheckResourceAttr(watchResourceName, "metadata", watchMetadataEmpty),
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "5000"),
					resource.TestCheckNoResourceAttr(watchResourceName, "transform"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestCheckResourceAttr(watchResourceName, "active", "true"),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerUpdateExpected),
					resource.TestCheckResourceAttr(watchResourceName, "input", watchInputSimpleExpected),
					resource.TestCheckResourceAttr(watchResourceName, "condition", watchConditionNever),
					resource.TestCheckResourceAttr(watchResourceName, "actions", watchActionsLogExpected),
					resource.TestCheckResourceAttr(watchResourceName, "metadata", watchMetadataExample),
					testCheckWatchTransformSemanticallyEqual(t, watchTransformExpected),
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "10000"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ResourceName:             watchResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_transform"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerUpdateExpected),
					resource.TestCheckResourceAttr(watchResourceName, "input", watchInputSimpleExpected),
					resource.TestCheckResourceAttr(watchResourceName, "condition", watchConditionNever),
					resource.TestCheckResourceAttr(watchResourceName, "actions", watchActionsLogExpected),
					resource.TestCheckResourceAttr(watchResourceName, "metadata", watchMetadataExample),
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "10000"),
					testCheckWatchTransformCleared(t),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_transform"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ResourceName:             watchResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestCheckResourceAttr(watchResourceName, "active", "true"),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerUpdateExpected),
					resource.TestCheckResourceAttr(watchResourceName, "input", watchInputSimpleExpected),
					resource.TestCheckResourceAttr(watchResourceName, "condition", watchConditionNever),
					resource.TestCheckResourceAttr(watchResourceName, "actions", watchActionsLogExpected),
					resource.TestCheckResourceAttr(watchResourceName, "metadata", watchMetadataExample),
					testCheckWatchTransformSemanticallyEqual(t, watchTransformExpected),
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "10000"),
				),
			},
			{
				// update2: verify a second distinct payload shape for input, condition, metadata, and transform.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update2"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestCheckResourceAttr(watchResourceName, "active", "true"),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerUpdateExpected),
					testCheckWatchAttrSemanticallyEqual(t, "input", watchInputSecondExpected),
					testCheckWatchAttrSemanticallyEqual(t, "condition", watchConditionScriptExpected),
					resource.TestCheckResourceAttr(watchResourceName, "actions", watchActionsLogExpected),
					testCheckWatchAttrSemanticallyEqual(t, "metadata", watchMetadataSecondExpected),
					testCheckWatchAttrSemanticallyEqual(t, "transform", watchTransformScriptExpected),
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "15000"),
				),
			},
			{
				// defaults_reset: verify that all attributes return to defaults after a rich configuration.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults_reset"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestCheckResourceAttr(watchResourceName, "active", "true"),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerUpdateExpected),
					resource.TestCheckResourceAttr(watchResourceName, "input", watchInputNoneExpected),
					resource.TestCheckResourceAttr(watchResourceName, "condition", watchConditionAlways),
					resource.TestCheckResourceAttr(watchResourceName, "actions", watchActionsEmpty),
					resource.TestCheckResourceAttr(watchResourceName, "metadata", watchMetadataEmpty),
					testCheckWatchTransformCleared(t),
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "5000"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults_reset"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ResourceName:             watchResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
			},
		},
	})
}

func TestResourceWatch_defaultsOmitted(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceWatchDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestCheckResourceAttr(watchResourceName, "active", "true"),
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "5000"),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerCreateExpected),
					resource.TestCheckResourceAttr(watchResourceName, "input", watchInputNoneExpected),
					resource.TestCheckResourceAttr(watchResourceName, "condition", watchConditionAlways),
					resource.TestCheckResourceAttr(watchResourceName, "actions", watchActionsEmpty),
					resource.TestCheckResourceAttr(watchResourceName, "metadata", watchMetadataEmpty),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ResourceName:             watchResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
			},
		},
	})
}

// TestAccResourceWatchFromSDK verifies that state created by the last SDK-based
// provider release (v0.14.3) can be read and updated without recreation by the
// current Plugin Framework implementation.
func TestAccResourceWatch_redactedWebhookAuthPreserved(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceWatchDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`acc-redacted-webhook-secret-9f2c`)),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`127\.0\.0\.1`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_throttle"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "12000"),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`acc-redacted-webhook-secret-9f2c`)),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`127\.0\.0\.1`)),
				),
			},
			{
				// Import after update: verify that non-redacted attributes round-trip correctly.
				// The actions attribute is excluded from import verification because Elasticsearch
				// redacts secret values on Get Watch; there is no prior state to restore from.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_throttle"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ResourceName:             watchResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection", "actions"},
			},
		},
	})
}

// TestAccResourceWatch_redactedScriptHeaderPreserved verifies the same
// drift-free behavior as TestAccResourceWatch_redactedWebhookAuthPreserved,
// but for the case where the prior known value at the redacted path is a
// non-string (an inline-script object Authorization header). Elasticsearch
// returns the redacted string sentinel at that path on Get Watch, and the
// provider must substitute the prior script object back so unrelated updates
// (here, throttle_period_in_millis) do not perpetually re-apply the actions.
func TestAccResourceWatch_redactedScriptHeaderPreserved(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceWatchDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`acc-script-header-3a91`)),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`"lang":"painless"`)),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`"Content-Type":"application/json"`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_throttle"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "throttle_period_in_millis", "12000"),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`acc-script-header-3a91`)),
					resource.TestMatchResourceAttr(watchResourceName, "actions", regexp.MustCompile(`"lang":"painless"`)),
				),
			},
			{
				// Import after update: verify that non-redacted attributes round-trip correctly.
				// The actions attribute is excluded from import verification because Elasticsearch
				// redacts the Authorization header on Get Watch; there is no prior state to restore from.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_throttle"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ResourceName:             watchResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection", "actions"},
			},
		},
	})
}

// TestResourceWatch_watchIDReplace verifies that changing watch_id forces resource replacement.
func TestResourceWatch_watchIDReplace(t *testing.T) {
	watchID1 := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	watchID2 := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceWatchDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID1)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(watchResourceName, "id"),
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID1),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerCreateExpected),
				),
			},
			{
				// Changing watch_id must trigger destroy-before-create (RequiresReplace).
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID2)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(watchResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(watchResourceName, "watch_id", watchID2),
					resource.TestCheckResourceAttr(watchResourceName, "trigger", watchTriggerCreateExpected),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID2)},
				ResourceName:             watchResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
			},
		},
	})
}

func TestAccResourceWatchFromSDK(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceWatchDestroy,
		Steps: []resource.TestStep{
			{
				// Create the watch with the last provider version where the watch resource was built on the SDK.
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source: "elastic/elasticstack",
						// last SDK-backed release — do not bump without re-checking upgrade compatibility
						VersionConstraint: "<= 0.14.3",
					},
				},
				Config: watchFromSDKCreateConfig,
				ConfigVariables: config.Variables{
					"watch_id": config.StringVariable(watchID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "active", "false"),
				),
			},
			{
				// Read and verify with the current PF implementation — must not force recreation.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("upgrade"),
				ConfigVariables:          config.Variables{"watch_id": config.StringVariable(watchID)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "active", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watch.test", "trigger", watchTriggerCreateExpected),
				),
			},
		},
	})
}

func checkResourceWatchDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_watch" {
			continue
		}
		compID, idDiags := clients.CompositeIDFromStr(rs.Primary.ID)
		if idDiags.HasError() {
			return fmt.Errorf("failed to parse resource ID: %v", idDiags)
		}

		typedClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		_, err = typedClient.Watcher.GetWatch(compID.ResourceID).Do(context.Background())
		if err != nil {
			if acctest.IsNotFoundElasticsearchError(err) {
				continue
			}
			return err
		}
		return fmt.Errorf("watch (%s) still exists", compID.ResourceID)
	}
	return nil
}
