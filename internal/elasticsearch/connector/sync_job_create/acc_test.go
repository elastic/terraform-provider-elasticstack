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

package sync_job_create_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func actionTerraformVersionChecks() []tfversion.TerraformVersionCheck {
	return []tfversion.TerraformVersionCheck{
		tfversion.SkipBelow(tfversion.Version1_14_0),
	}
}

func skipConnectorUnsupported() func() (bool, error) {
	return versionutils.CheckIfVersionIsUnsupported(connector.MinSupportedVersion)
}

func syncJobConfigVariables(connectorID string, waitForCompletion bool) config.Variables {
	return config.Variables{
		"connector_id":        config.StringVariable(connectorID),
		"wait_for_completion": config.BoolVariable(waitForCompletion),
	}
}

func checkDestroyConnectorSyncJobCreate(connectorID string) func(*terraform.State) error {
	return func(*terraform.State) error {
		return cleanupConnectorAndSyncJobs(connectorID)
	}
}

func cleanupConnectorAndSyncJobs(connectorID string) error {
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	typedClient := client.GetESClient()
	jobs, err := typedClient.Connector.SyncJobList().ConnectorId(connectorID).Do(ctx)
	if err == nil {
		for _, job := range jobs.Results {
			_, delErr := typedClient.Connector.SyncJobDelete(job.Id).Do(ctx)
			if delErr != nil && !esclient.IsNotFoundElasticsearchError(delErr) {
				return fmt.Errorf("delete sync job %q: %w", job.Id, delErr)
			}
		}
	}

	if diags := esclient.DeleteConnector(ctx, client, connectorID); diags.HasError() {
		return fmt.Errorf("delete connector %q: %s", connectorID, diags[0].Summary())
	}

	return nil
}

func testAccCheckSyncJobsExist(connectorID string, minCount int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		resp, err := client.GetESClient().Connector.SyncJobList().ConnectorId(connectorID).Do(ctx)
		if err != nil {
			return fmt.Errorf("list sync jobs for connector %q: %w", connectorID, err)
		}
		if len(resp.Results) < minCount {
			return fmt.Errorf("expected at least %d sync job(s) for connector %q, got %d", minCount, connectorID, len(resp.Results))
		}
		return nil
	}
}

func TestAccActionConnectorSyncJobCreate_async(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-action-async")

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkDestroyConnectorSyncJobCreate(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sync"),
				ConfigVariables:          syncJobConfigVariables(connectorID, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSyncJobsExist(connectorID, 1),
				),
			},
		},
	})
}

func TestAccActionConnectorSyncJobCreate_syncWaitCompletion(t *testing.T) {
	if os.Getenv("CONNECTOR_SERVICE_RUNNING") != "1" {
		t.Skip("requires a running connector service (set CONNECTOR_SERVICE_RUNNING=1 to enable)")
	}

	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-action-sync")

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkDestroyConnectorSyncJobCreate(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sync"),
				ConfigVariables:          syncJobConfigVariables(connectorID, true),
			},
		},
	})
}

func TestAccActionConnectorSyncJobCreate_timeout(t *testing.T) {
	connectorID := sdkacctest.RandomWithPrefix("tf-acc-test-action-timeout")

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		CheckDestroy:           checkDestroyConnectorSyncJobCreate(connectorID),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sync"),
				ConfigVariables: config.Variables{
					"connector_id":        config.StringVariable(connectorID),
					"wait_for_completion": config.BoolVariable(true),
					"invoke_timeout":      config.StringVariable("5s"),
				},
				ExpectError: regexp.MustCompile(`(?s)Sync job did not complete within timeout.*Sync job.*last observed status`),
			},
		},
	})
}

// TestAccActionConnectorSyncJobCreate_errorStatus covers REQ-SYNC-001-E.
// Terminal error status requires a running connector service; REQ-SYNC-001-E
// wording is verified in unit TestClassifyTerminalStatus/error_with_message.
func TestAccActionConnectorSyncJobCreate_errorStatus(t *testing.T) {
	if os.Getenv("CONNECTOR_SERVICE_RUNNING") != "1" {
		t.Skip("error status terminal requires a running connector service to transition sync job to error; see CONNECTOR_SERVICE_RUNNING gate on TestAccActionConnectorSyncJobCreate_syncWaitCompletion")
	}
	t.Skip("error status acceptance scenario requires a connector service that fails sync with status=error; not yet automated")
}

func TestAccActionConnectorSyncJobCreate_connectorNotFound(t *testing.T) {
	connectorID := "tf-acc-test-nonexistent"

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acctest.PreCheck(t) },
		TerraformVersionChecks: actionTerraformVersionChecks(),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipConnectorUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("not_found"),
				ConfigVariables:          syncJobConfigVariables(connectorID, false),
				ExpectError:              regexp.MustCompile(`(?s).*(?i)connector.*does not exist.*`),
			},
		},
	})
}

