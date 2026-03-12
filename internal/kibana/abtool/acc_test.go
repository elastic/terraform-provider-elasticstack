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

package abtool_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	providerConfig = `
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}
`
)

var (
	minKibanaAgentBuilderAPIVersion = version.Must(version.NewVersion("9.3.0"))
)

func preCheckWithWorkflowsEnabled(t *testing.T) {
	acctest.PreCheck(t)

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	serverVersion, diags := client.ServerVersion(context.Background())
	if diags.HasError() {
		t.Fatalf("Failed to get server version: %v", diags)
	}
	if serverVersion.LessThan(minKibanaAgentBuilderAPIVersion) {
		t.Skipf("Skipping test: server version %s is below minimum %s", serverVersion, minKibanaAgentBuilderAPIVersion)
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Fatalf("Failed to get Kibana client: %v", err)
	}

	// Try the internal settings API endpoint
	settingsURL := fmt.Sprintf("%s/internal/kibana/settings/workflows:ui:enabled", kibanaClient.URL)
	body := map[string]any{
		"value": true,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", settingsURL, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("x-elastic-internal-origin", "Kibana")

	resp, err := kibanaClient.HTTP.Do(req)
	if err != nil {
		t.Fatalf("Failed to enable workflows: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to enable workflows (status %d): %s. Make sure workflows are enabled in kibana.yml with 'xpack.aiAssistant.workflows.enabled: true'", resp.StatusCode, string(respBody))
	}
}

func TestAccResourceToolEsql(t *testing.T) {
	toolID := "test-esql-tool-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_ab_tool.test_esql"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckWithWorkflowsEnabled(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_ab_tool" "test_esql" {
	id          = "` + toolID + `"
	type        = "esql"
	description = "Test ES|QL tool"
	tags        = ["test", "esql"]
	configuration = jsonencode({
		query = "FROM logs-* | LIMIT ?limit"
		params = {
			limit = {
				type        = "integer"
				description = "Maximum number of results to return"
			}
		}
	})
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", toolID),
					resource.TestCheckResourceAttr(resourceID, "type", "esql"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test ES|QL tool"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "test"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "esql"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_ab_tool" "test_esql" {
	id          = "` + toolID + `"
	type        = "esql"
	description = "Updated ES|QL tool"
	tags        = ["test", "esql", "updated"]
	configuration = jsonencode({
		query = "FROM logs-* | WHERE cloud.region == ?region | LIMIT ?limit"
		params = {
			limit = {
				type        = "integer"
				description = "Maximum number of results to return"
			}
			region = {
				type        = "keyword"
				description = "Region filter"
			}
		}
	})
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", toolID),
					resource.TestCheckResourceAttr(resourceID, "description", "Updated ES|QL tool"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
				),
			},
		},
	})
}

func TestAccResourceToolIndexSearch(t *testing.T) {
	toolID := "test-index-search-tool-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_ab_tool.test_index_search"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckWithWorkflowsEnabled(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_ab_tool" "test_index_search" {
	id          = "` + toolID + `"
	type        = "index_search"
	description = "Test index search tool"
	tags        = ["test", "search"]
	configuration = jsonencode({
		pattern = "logs-*"
	})
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", toolID),
					resource.TestCheckResourceAttr(resourceID, "type", "index_search"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test index search tool"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_elasticsearch_index" "logs_system" {
	name                = "logs-system.test"
	deletion_protection = false
}

resource "elasticstack_kibana_ab_tool" "test_index_search" {
	id          = "` + toolID + `"
	type        = "index_search"
	description = "Updated index search tool"
	tags        = ["test", "search", "updated"]
	configuration = jsonencode({
		pattern = "logs-system.*"
	})
	depends_on = [elasticstack_elasticsearch_index.logs_system]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", toolID),
					resource.TestCheckResourceAttr(resourceID, "description", "Updated index search tool"),
				),
			},
		},
	})
}

func TestAccResourceToolWorkflow(t *testing.T) {
	toolID := "test-workflow-tool-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_ab_tool.test_workflow"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckWithWorkflowsEnabled(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_ab_workflow" "test" {
	configuration = <<-EOT
name: New workflow
enabled: false
description: This is a new workflow
triggers:
  - type: manual
inputs:
  - name: message
    type: string
    default: "hello world"
steps:
  - name: hello_world_step
    type: console
    with:
      message: "{{ inputs.message }}"
EOT
}

resource "elasticstack_kibana_ab_tool" "test_workflow" {
	id          = "` + toolID + `"
	type        = "workflow"
	description = "Test workflow tool"
	tags        = ["test", "workflow"]
	configuration = jsonencode({
		workflow_id = elasticstack_kibana_ab_workflow.test.id
	})
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", toolID),
					resource.TestCheckResourceAttr(resourceID, "type", "workflow"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test workflow tool"),
					resource.TestCheckResourceAttrSet(resourceID, "configuration"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaAgentBuilderAPIVersion),
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
