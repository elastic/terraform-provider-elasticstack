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

package followerindex_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const followerIndexResourceName = "elasticstack_elasticsearch_ccr_follower_index.test"

func TestAccResourceCCRFollowerIndex_basic(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	followerIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName, "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkFollowerIndexPromotedToRegular(followerIndexName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "name", followerIndexName),
					resource.TestCheckResourceAttr(followerIndexResourceName, "remote_cluster", ccrEnv.RemoteClusterAlias),
					resource.TestCheckResourceAttr(followerIndexResourceName, "leader_index", leaderIndexName),
					resource.TestCheckResourceAttr(followerIndexResourceName, "status", "active"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_outstanding_read_requests", "12"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "delete_index_on_destroy", "false"),
					testAccCheckFollowerIndexStatus(followerIndexName, "active"),
					testAccCheckFollowerIndexMaxOutstandingReadRequests(followerIndexName, 12),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_outstanding_read_requests", "24"),
					testAccCheckFollowerIndexMaxOutstandingReadRequests(followerIndexName, 24),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceCCRFollowerIndex_deleteIndexOnDestroy(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	followerIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName, "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkFollowerIndexDeleted(followerIndexName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_delete"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "delete_index_on_destroy", "true"),
					testAccCheckFollowerIndexStatus(followerIndexName, "active"),
				),
			},
		},
	})
}

func TestAccResourceCCRFollowerIndex_status(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	followerIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName, "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkFollowerIndexPromotedToRegular(followerIndexName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_paused"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "status", "paused"),
					testAccCheckFollowerIndexStatus(followerIndexName, "paused"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_active"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "status", "active"),
					testAccCheckFollowerIndexStatus(followerIndexName, "active"),
				),
			},
		},
	})
}

func TestAccResourceCCRFollowerIndex_dataStreamNameImport(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	dataStreamName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha)
	followerIndexName := dataStreamName
	vars := ccrDataStreamFollowerVariables(ccrEnv, dataStreamName, followerIndexName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkFollowerIndexPromotedToRegular(followerIndexName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "name", followerIndexName),
					resource.TestCheckResourceAttr(followerIndexResourceName, "leader_index", dataStreamName),
					resource.TestCheckResourceAttr(followerIndexResourceName, "data_stream_name", dataStreamName),
					testAccCheckFollowerIndexStatus(followerIndexName, "active"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ConfigVariables:          vars,
				ResourceName:             followerIndexResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"settings_raw",
					"data_stream_name",
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(followerIndexResourceName, "data_stream_name"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan_with_data_stream_name"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(followerIndexResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func ccrFollowerVariables(ccrEnv acctest.CCRTestEnv, leaderIndexName, followerIndexName, dataStreamName string) config.Variables {
	vars := config.Variables{
		"remote_cluster_alias": config.StringVariable(ccrEnv.RemoteClusterAlias),
		"remote_proxy_address": config.StringVariable(ccrEnv.RemoteProxyAddress),
		"leader_index_name":    config.StringVariable(leaderIndexName),
		"follower_index_name":  config.StringVariable(followerIndexName),
	}
	if dataStreamName != "" {
		vars["data_stream_name"] = config.StringVariable(dataStreamName)
	}
	return vars
}

func ccrDataStreamFollowerVariables(ccrEnv acctest.CCRTestEnv, dataStreamName, followerIndexName string) config.Variables {
	return config.Variables{
		"remote_cluster_alias": config.StringVariable(ccrEnv.RemoteClusterAlias),
		"remote_proxy_address": config.StringVariable(ccrEnv.RemoteProxyAddress),
		"data_stream_name":     config.StringVariable(dataStreamName),
		"follower_index_name":  config.StringVariable(followerIndexName),
	}
}

func testAccCheckFollowerIndexStatus(indexName, wantStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		follower, diags := esclient.GetFollowerIndex(ctx, client, indexName)
		if diags.HasError() {
			return fmt.Errorf("get follower index %q: %v", indexName, diags)
		}
		if follower == nil {
			return fmt.Errorf("follower index %q not found in Elasticsearch", indexName)
		}
		if follower.Status.String() != wantStatus {
			return fmt.Errorf("follower index %q status %q, want %q", indexName, follower.Status.String(), wantStatus)
		}
		return nil
	}
}

func testAccCheckFollowerIndexMaxOutstandingReadRequests(indexName string, want int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		follower, diags := esclient.GetFollowerIndex(ctx, client, indexName)
		if diags.HasError() {
			return fmt.Errorf("get follower index %q: %v", indexName, diags)
		}
		if follower == nil || follower.Parameters == nil || follower.Parameters.MaxOutstandingReadRequests == nil {
			return fmt.Errorf("follower index %q has no readable max_outstanding_read_requests", indexName)
		}
		if *follower.Parameters.MaxOutstandingReadRequests != want {
			return fmt.Errorf(
				"follower index %q max_outstanding_read_requests %d, want %d",
				indexName,
				*follower.Parameters.MaxOutstandingReadRequests,
				want,
			)
		}
		return nil
	}
}

func checkFollowerIndexPromotedToRegular(indexName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		follower, diags := esclient.GetFollowerIndex(ctx, client, indexName)
		if diags.HasError() {
			return fmt.Errorf("get follower index %q: %v", indexName, diags)
		}
		if follower != nil {
			return fmt.Errorf("index %q is still a CCR follower after destroy", indexName)
		}

		index, diags := esclient.GetIndex(ctx, client, indexName)
		if diags.HasError() {
			return fmt.Errorf("get index %q: %v", indexName, diags)
		}
		if index == nil {
			return fmt.Errorf("index %q does not exist after destroy with delete_index_on_destroy=false", indexName)
		}

		return nil
	}
}

func checkFollowerIndexDeleted(indexName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		index, diags := esclient.GetIndex(ctx, client, indexName)
		if diags.HasError() {
			return fmt.Errorf("get index %q: %v", indexName, diags)
		}
		if index != nil {
			return fmt.Errorf("index %q still exists after destroy with delete_index_on_destroy=true", indexName)
		}

		return nil
	}
}
