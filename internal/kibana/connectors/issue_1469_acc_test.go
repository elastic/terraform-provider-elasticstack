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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue1469 is a regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/1469
//
// A webhook connector with hasAuth=false and an Authorization header in config
// previously failed with "Provider produced inconsistent result after apply:
// .config: inconsistent values for sensitive attribute" when the config
// referenced a sensitive variable. The Kibana API adds method="post" as a
// default field when it is omitted from the user's config. The fix adds
// method="post" to ConnectorConfigWithDefaults for .webhook so that the
// post-apply read matches the plan.
func TestAccReproduceIssue1469(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				// Kibana adds method="post" as a default when method is omitted.
				// The provider now includes this default in ConnectorConfigWithDefaults
				// for .webhook, so the post-apply read matches the plan and no
				// inconsistency error is produced.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				Check: testCommonAttributes(connectorName, ".webhook"),
			},
			{
				// A subsequent plan must produce no diff, confirming there is no
				// persistent drift from the Kibana-added method="post" default.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"connector_name": config.StringVariable(connectorName),
				},
				PlanOnly: true,
			},
		},
	})
}
