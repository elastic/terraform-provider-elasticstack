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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue1469 reproduces the bug reported in
// https://github.com/elastic/terraform-provider-elasticstack/issues/1469
//
// A webhook connector with hasAuth=false and an Authorization header in config
// fails with "Provider produced inconsistent result after apply: .config:
// inconsistent values for sensitive attribute" when the Terraform config
// references a sensitive variable. The Kibana API adds method="post" as a
// default field when it is omitted from the user's config. The provider's
// ConnectorConfigWithDefaults for .webhook does not account for this default,
// causing the post-apply read (which includes method="post") to differ from
// the planned config (which does not include method), triggering the
// inconsistency error from Terraform.
func TestAccReproduceIssue1469(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	// The config mirrors the reporter's setup: hasAuth=false, Authorization
	// header in config (not secrets), no method field, and the auth token
	// is a sensitive variable so Terraform treats the entire config value
	// as sensitive - matching the "(sensitive value)" plan output in the report.
	makeConfig := func(name string) string {
		return fmt.Sprintf(`
variable "auth_token" {
  type      = string
  sensitive = true
  default   = "test-bearer-token-for-issue-1469"
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = %q
  connector_type_id = ".webhook"
  config = jsonencode({
    hasAuth = false
    headers = {
      "Authorization" = "Bearer ${var.auth_token}"
      "Content-Type"  = "application/json"
    }
    url = "https://example.com/webhook"
  })
  secrets = jsonencode({})
}
`, name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceKibanaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				// Kibana adds method="post" as a default when method is omitted.
				// The sensitive config value referencing var.auth_token causes
				// Terraform to report "inconsistent values for sensitive attribute"
				// when the post-apply state (which includes method="post") is
				// compared to the planned config (which does not include method).
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   makeConfig(connectorName),
				ExpectError:              regexp.MustCompile(`inconsistent`),
			},
		},
	})
}
