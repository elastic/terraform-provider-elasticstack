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
// A user creates a .slack_api connector with allowedChannels that only carry a
// `name`. kbapi.SlackApiConfig.AllowedChannels[].Id lacks omitempty, so when the
// provider remarshals the Kibana read response each channel gains an empty
// `"id":""`. Because .slack_api had no `defaults` handler, the config attribute's
// semantic-equality check did not normalize the planned value the same way, so
// the planned value (no id) and the value read back after apply (id="") diverged
// and Terraform failed with "Provider produced inconsistent result after apply".
// The fix wires the typed remarshal in as the .slack_api defaults function so
// both sides normalize identically. See
// internal/clients/kibanaoapi/connector_defaults.go.
//
// Version support, verified against this provider's full CI version matrix:
//
//	8.0.1 - 8.7.1   .slack_api connector type not registered
//	8.8.2 - 8.10.3  connector registered, but no allowedChannels option
//	8.11.x - 9.2.x  allowedChannels present; each channel requires both id and name
//	9.3.0+          id becomes optional, so a name-only channel is accepted
//
// The bug only manifests when a channel omits a field that Kibana injects back as
// "". Since `name` is always required, the only omittable field is `id`, which is
// optional only from 9.3 onward. The test therefore uses a name-only channel and
// is gated to Kibana 9.3.0+.
func TestAccResourceKibanaConnectorSlackAPI(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("9.3.0"))

	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				// Create must succeed without an inconsistent-result error.
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCommonAttributes(connectorName, ".slack_api"),
					// The user-supplied channels survive the round-trip and the
					// provider-internal context key must not leak into state.
					resource.TestCheckResourceAttrWith("elasticstack_kibana_action_connector.test", "config", func(value string) error {
						if strings.Contains(value, "__tf_provider_context") {
							return fmt.Errorf("config leaked internal __tf_provider_context key into state: %s", value)
						}
						for _, channel := range []string{"#kar_testing", "#test-prod-alerts"} {
							if !strings.Contains(value, channel) {
								return fmt.Errorf("config missing expected channel %q: %s", channel, value)
							}
						}
						return nil
					}),
				),
			},
			{
				// A subsequent plan must produce no diff, confirming there is no
				// persistent drift from the Kibana-added empty `id` fields.
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				PlanOnly: true,
			},
		},
	})
}
