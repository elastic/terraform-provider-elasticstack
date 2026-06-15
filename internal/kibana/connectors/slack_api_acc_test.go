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

package connectors_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccResourceKibanaConnectorSlackAPI is a regression test for the
// ".slack_api" connector "Provider produced inconsistent result after apply"
// bug.
//
// kbapi.SlackApiConfig.AllowedChannels[] has both `Id` and `Name` fields, and
// neither carries omitempty. Kibana echoes the channels back and the provider
// remarshals the read response through kbapi.SlackApiConfig, so a channel the
// user supplied with only one of the two fields gains an empty value for the
// other (a name-only channel gains `"id":""`, an id-only channel gains
// `"name":""`). Because .slack_api had no `defaults` handler, the config
// attribute's semantic-equality check did not normalize the planned value the
// same way, so the planned value and the value read back after apply diverged
// and Terraform failed with "Provider produced inconsistent result after
// apply". The fix wires the typed remarshal in as the .slack_api defaults
// function so both sides normalize identically. See
// internal/clients/kibanaoapi/connector_defaults.go.
//
// The two cases also confirm the resource works across stack versions. Verified
// against this provider's full CI version matrix:
//
//	8.0.1 - 8.7.1   .slack_api connector type not registered
//	8.8.2 - 8.10.3  connector registered, but no allowedChannels option
//	8.11.x - 9.2.x  allowedChannels present, channels must use `id`
//	9.3.0+          allowedChannels accepts `name`
//
// So the id-based case is gated to 8.11 (where allowedChannels first appears) and
// the name-based case to 9.3.
func TestAccResourceKibanaConnectorSlackAPI(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	testCases := []struct {
		name             string
		minVersion       *version.Version
		expectedChannels []string
	}{
		{
			// "Allowed channel names" (channels identified by `name`) is GA since
			// Kibana 9.3; older stacks reject a name-only channel.
			name:             "channels_by_name",
			minVersion:       version.Must(version.NewSemver("9.3.0")),
			expectedChannels: []string{"#kar_testing", "#test-prod-alerts"},
		},
		{
			// Channels identified by `id` work wherever the allowedChannels option
			// exists (Kibana 8.11+; CI proves 8.11.4). This exercises the symmetric
			// empty-`name` injection on versions older than the name-based case.
			name:             "channels_by_id",
			minVersion:       version.Must(version.NewSemver("8.11.0")),
			expectedChannels: []string{"C0123456789", "C9876543210"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vars := config.Variables{
				"connector_name": config.StringVariable(connectorName),
			}

			resource.Test(t, resource.TestCase{
				PreCheck:     func() { acctest.PreCheck(t) },
				CheckDestroy: checkResourceKibanaConnectorDestroy,
				Steps: []resource.TestStep{
					{
						// Create must succeed without an inconsistent-result error.
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(tc.minVersion),
						ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
						ConfigVariables:          vars,
						Check: resource.ComposeTestCheckFunc(
							testCommonAttributes(connectorName, ".slack_api"),
							// The user-supplied channels survive the round-trip and
							// the provider-internal context key must not leak into
							// state.
							resource.TestCheckResourceAttrWith("elasticstack_kibana_action_connector.test", "config", func(value string) error {
								if strings.Contains(value, "__tf_provider_context") {
									return fmt.Errorf("config leaked internal __tf_provider_context key into state: %s", value)
								}
								for _, channel := range tc.expectedChannels {
									if !strings.Contains(value, channel) {
										return fmt.Errorf("config missing expected channel %q: %s", channel, value)
									}
								}
								return nil
							}),
						),
					},
					{
						// A subsequent plan must produce no diff, confirming there is
						// no persistent drift from the Kibana-added empty fields.
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(tc.minVersion),
						ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
						ConfigVariables:          vars,
						PlanOnly:                 true,
					},
				},
			})
		})
	}
}
