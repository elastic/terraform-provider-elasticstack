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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr/followerindex"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const followerIndexResourceName = "elasticstack_elasticsearch_ccr_follower_index.test"

func TestAccResourceCCRFollowerIndex_basic(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	followerIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName)

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
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName)

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
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName)

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
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_paused"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "status", "paused"),
					testAccCheckFollowerIndexStatus(followerIndexName, "paused"),
				),
			},
		},
	})
}

func TestAccResourceCCRFollowerIndex_settingsRaw(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	followerIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkFollowerIndexPromotedToRegular(followerIndexName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						followerIndexResourceName,
						"settings_raw",
						`{"index.refresh_interval":"30s"}`,
					),
					testAccCheckFollowerIndexRefreshInterval(followerIndexName, "30s"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						followerIndexResourceName,
						"settings_raw",
						`{"index.refresh_interval":"60s"}`,
					),
					testAccCheckFollowerIndexRefreshInterval(followerIndexName, "60s"),
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

func TestAccResourceCCRFollowerIndex_dataStreamNameImport(t *testing.T) {
	// data_stream_name on the CCR follow API requires Elasticsearch 8.4.0+.
	versionutils.SkipIfUnsupported(t, followerindex.MinVersionDataStreamName, versionutils.FlavorStateful)

	ccrEnv := acctest.PreCheckCCR(t)
	// Leader and follower data streams share a cluster in the self-remote
	// environment, so their names must differ. A follower attaches to a local
	// data stream by following a backing index of the leader data stream.
	leaderDataStreamName := "leader-" + strings.ToLower(sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))
	followerDataStreamName := "follower-" + strings.ToLower(sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))
	followerIndexName := strings.ToLower(sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha))
	vars := ccrDataStreamFollowerVariables(ccrEnv, leaderDataStreamName, followerDataStreamName, followerIndexName)

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
					resource.TestCheckResourceAttr(followerIndexResourceName, "data_stream_name", followerDataStreamName),
					resource.TestMatchResourceAttr(
						followerIndexResourceName,
						"leader_index",
						regexp.MustCompile(`^\.ds-`+regexp.QuoteMeta(leaderDataStreamName)+`-`),
					),
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
				// data_stream_name is a create-only request parameter that GET
				// /_ccr/info never returns, so it cannot be verified on import.
				ImportStateVerifyIgnore: []string{
					"settings_raw",
					"data_stream_name",
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(followerIndexResourceName, "data_stream_name"),
				),
			},
		},
	})
}

// TestAccResourceCCRFollowerIndex_dataStreamNameVersionGate verifies that
// setting data_stream_name on a cluster older than 8.4.0 fails fast with a
// clear version-requirement error instead of the raw Elasticsearch parse error.
// It only runs on clusters below 8.4.0.
func TestAccResourceCCRFollowerIndex_dataStreamNameVersionGate(t *testing.T) {
	constraints, err := version.NewConstraint("< " + followerindex.MinVersionDataStreamName.String())
	if err != nil {
		t.Fatal(err)
	}
	versionutils.SkipIfUnsupportedConstraints(t, constraints, versionutils.FlavorStateful)

	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := strings.ToLower(sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha))
	followerIndexName := strings.ToLower(sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlpha))
	dataStreamName := "follower-" + strings.ToLower(sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))

	vars := config.Variables{
		"remote_cluster_alias": config.StringVariable(ccrEnv.RemoteClusterAlias),
		"remote_proxy_address": config.StringVariable(ccrEnv.RemoteProxyAddress),
		"leader_index_name":    config.StringVariable(leaderIndexName),
		"follower_index_name":  config.StringVariable(followerIndexName),
		"data_stream_name":     config.StringVariable(dataStreamName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheckCCR(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile(`data_stream_name attribute is only supported on Elasticsearch 8\.4\.0`),
			},
		},
	})
}

func TestAccResourceCCRFollowerIndex_params(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	followerIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	vars := ccrFollowerVariables(ccrEnv, leaderIndexName, followerIndexName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkFollowerIndexPromotedToRegular(followerIndexName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_outstanding_read_requests", "15"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_outstanding_write_requests", "10"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_read_request_operation_count", "5120"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_read_request_size", "40mb"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_retry_delay", "500ms"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_buffer_count", "512"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_buffer_size", "512mb"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_request_operation_count", "5120"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_request_size", "40mb"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "read_poll_timeout", "5s"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_outstanding_read_requests", "30"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_outstanding_write_requests", "20"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_read_request_operation_count", "2048"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_read_request_size", "20mb"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_retry_delay", "1s"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_buffer_count", "256"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_buffer_size", "256mb"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_request_operation_count", "2048"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "max_write_request_size", "20mb"),
					resource.TestCheckResourceAttr(followerIndexResourceName, "read_poll_timeout", "10s"),
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

func ccrFollowerVariables(ccrEnv acctest.CCRTestEnv, leaderIndexName, followerIndexName string) config.Variables {
	return config.Variables{
		"remote_cluster_alias": config.StringVariable(ccrEnv.RemoteClusterAlias),
		"remote_proxy_address": config.StringVariable(ccrEnv.RemoteProxyAddress),
		"leader_index_name":    config.StringVariable(leaderIndexName),
		"follower_index_name":  config.StringVariable(followerIndexName),
	}
}

func ccrDataStreamFollowerVariables(ccrEnv acctest.CCRTestEnv, leaderDataStreamName, followerDataStreamName, followerIndexName string) config.Variables {
	return config.Variables{
		"remote_cluster_alias":    config.StringVariable(ccrEnv.RemoteClusterAlias),
		"remote_proxy_address":    config.StringVariable(ccrEnv.RemoteProxyAddress),
		"leader_data_stream_name": config.StringVariable(leaderDataStreamName),
		"data_stream_name":        config.StringVariable(followerDataStreamName),
		"follower_index_name":     config.StringVariable(followerIndexName),
	}
}

func testAccCheckFollowerIndexStatus(indexName, wantStatus string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
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
	return func(_ *terraform.State) error {
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

func testAccCheckFollowerIndexRefreshInterval(indexName, want string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		index, diags := esclient.GetIndex(ctx, client, indexName)
		if diags.HasError() {
			return fmt.Errorf("get index %q: %v", indexName, diags)
		}
		got, ok := indexRefreshIntervalFromState(index)
		if !ok {
			return fmt.Errorf("index %q has no readable refresh_interval setting", indexName)
		}
		if got != want {
			return fmt.Errorf("index %q refresh_interval %q, want %q", indexName, got, want)
		}
		return nil
	}
}

func indexRefreshIntervalFromState(index *types.IndexState) (string, bool) {
	if index == nil || index.Settings == nil {
		return "", false
	}

	settings := index.Settings
	if settings.Index != nil {
		if value, ok := durationString(settings.Index.RefreshInterval); ok {
			return value, true
		}
	}
	if value, ok := durationString(settings.RefreshInterval); ok {
		return value, true
	}
	if raw, ok := settings.IndexSettings["index.refresh_interval"]; ok {
		var value string
		if err := json.Unmarshal(raw, &value); err == nil && value != "" {
			return value, true
		}
	}

	return "", false
}

func durationString(value types.Duration) (string, bool) {
	if value == nil {
		return "", false
	}
	switch v := value.(type) {
	case string:
		if v == "" {
			return "", false
		}
		return v, true
	default:
		rendered := fmt.Sprintf("%v", value)
		if rendered == "" || rendered == "<nil>" {
			return "", false
		}
		return rendered, true
	}
}

func checkFollowerIndexPromotedToRegular(indexName string) func(*terraform.State) error {
	return func(_ *terraform.State) error {
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
	return func(_ *terraform.State) error {
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
