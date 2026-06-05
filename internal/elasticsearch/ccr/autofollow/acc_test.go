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

package autofollow_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const autoFollowResourceName = "elasticstack_elasticsearch_ccr_auto_follow_pattern.test"

const followIndexPatternReplica = "{{leader_index}}-replica"

var autoFollowImportStateVerifyIgnore = []string{
	"settings_raw",
	"max_outstanding_write_requests",
	"max_read_request_operation_count",
	"max_read_request_size",
	"max_retry_delay",
	"max_write_buffer_count",
	"max_write_buffer_size",
	"max_write_request_operation_count",
	"max_write_request_size",
	"read_poll_timeout",
}

func TestAccResourceCCRAutoFollowPattern_basic(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	patternName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	leaderPattern := leaderIndexName[:4] + "*"
	metricsPattern := "metrics-*"
	vars := ccrAutoFollowVariables(ccrEnv, autoFollowVariableOptions{
		leaderIndexName:    leaderIndexName,
		patternName:        patternName,
		leaderPatterns:     []string{leaderPattern},
		followIndexPattern: followIndexPatternReplica,
	})
	varsWithExclusion := ccrAutoFollowVariables(ccrEnv, autoFollowVariableOptions{
		leaderIndexName:    leaderIndexName,
		patternName:        patternName,
		leaderPatterns:     []string{leaderPattern, metricsPattern},
		exclusionPatterns:  []string{leaderIndexName},
		followIndexPattern: followIndexPatternReplica,
		maxOutstandingRead: 10,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkAutoFollowPatternDestroyed(patternName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(autoFollowResourceName, "name", patternName),
					resource.TestCheckResourceAttr(autoFollowResourceName, "remote_cluster", ccrEnv.RemoteClusterAlias),
					resource.TestCheckResourceAttr(autoFollowResourceName, "active", "true"),
					resource.TestCheckResourceAttr(autoFollowResourceName, "follow_index_pattern", followIndexPatternReplica),
					resource.TestCheckTypeSetElemAttr(autoFollowResourceName, "leader_index_patterns.*", leaderPattern),
					testAccCheckAutoFollowPatternActive(patternName, true),
					testAccCheckAutoFollowPatternFollowIndexPattern(patternName, followIndexPatternReplica),
					testAccCheckAutoFollowPatternLeaderIndexPatterns(patternName, leaderPattern),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_patterns"),
				ConfigVariables:          varsWithExclusion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(autoFollowResourceName, "leader_index_patterns.*", leaderPattern),
					resource.TestCheckTypeSetElemAttr(autoFollowResourceName, "leader_index_patterns.*", metricsPattern),
					testAccCheckAutoFollowPatternLeaderIndexPatterns(patternName, leaderPattern, metricsPattern),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_exclusions"),
				ConfigVariables:          varsWithExclusion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(
						autoFollowResourceName,
						"leader_index_exclusion_patterns.*",
						leaderIndexName,
					),
					testAccCheckAutoFollowPatternExclusion(patternName, leaderIndexName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_exclusions"),
				ConfigVariables:          varsWithExclusion,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceCCRAutoFollowPattern_active(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	patternName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	leaderPattern := leaderIndexName[:4] + "*"
	vars := ccrAutoFollowVariables(ccrEnv, autoFollowVariableOptions{
		leaderIndexName: leaderIndexName,
		patternName:     patternName,
		leaderPatterns:  []string{leaderPattern},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkAutoFollowPatternDestroyed(patternName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(autoFollowResourceName, "active", "true"),
					testAccCheckAutoFollowPatternActive(patternName, true),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_inactive"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(autoFollowResourceName, "active", "false"),
					testAccCheckAutoFollowPatternActive(patternName, false),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_active"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(autoFollowResourceName, "active", "true"),
					testAccCheckAutoFollowPatternActive(patternName, true),
				),
			},
		},
	})
}

func TestAccResourceCCRAutoFollowPattern_import(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	patternName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	leaderPattern := leaderIndexName[:4] + "*"
	vars := ccrAutoFollowVariables(ccrEnv, autoFollowVariableOptions{
		leaderIndexName:    leaderIndexName,
		patternName:        patternName,
		leaderPatterns:     []string{leaderPattern},
		maxOutstandingRead: 10,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkAutoFollowPatternDestroyed(patternName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(autoFollowResourceName, "name", patternName),
					resource.TestCheckResourceAttr(autoFollowResourceName, "remote_cluster", ccrEnv.RemoteClusterAlias),
					resource.TestCheckResourceAttr(autoFollowResourceName, "active", "true"),
					resource.TestCheckResourceAttr(autoFollowResourceName, "max_outstanding_read_requests", "10"),
					resource.TestCheckTypeSetElemAttr(autoFollowResourceName, "leader_index_patterns.*", leaderPattern),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             autoFollowResourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  autoFollowImportStateVerifyIgnore,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(autoFollowResourceName, "name", patternName),
					resource.TestCheckResourceAttr(autoFollowResourceName, "remote_cluster", ccrEnv.RemoteClusterAlias),
					resource.TestCheckResourceAttr(autoFollowResourceName, "active", "true"),
					resource.TestCheckResourceAttr(autoFollowResourceName, "max_outstanding_read_requests", "10"),
					resource.TestCheckTypeSetElemAttr(autoFollowResourceName, "leader_index_patterns.*", leaderPattern),
				),
			},
		},
	})
}

type autoFollowVariableOptions struct {
	leaderIndexName    string
	patternName        string
	leaderPatterns     []string
	exclusionPatterns  []string
	followIndexPattern string
	maxOutstandingRead int64
}

func ccrAutoFollowVariables(ccrEnv acctest.CCRTestEnv, opts autoFollowVariableOptions) config.Variables {
	vars := config.Variables{
		"remote_cluster_alias": config.StringVariable(ccrEnv.RemoteClusterAlias),
		"remote_proxy_address": config.StringVariable(ccrEnv.RemoteProxyAddress),
		"leader_index_name":    config.StringVariable(opts.leaderIndexName),
		"pattern_name":         config.StringVariable(opts.patternName),
	}

	if len(opts.leaderPatterns) > 0 {
		patternVars := make([]config.Variable, len(opts.leaderPatterns))
		for i, pattern := range opts.leaderPatterns {
			patternVars[i] = config.StringVariable(pattern)
		}
		vars["leader_index_patterns"] = config.ListVariable(patternVars...)
	}

	if len(opts.exclusionPatterns) > 0 {
		exclusionVars := make([]config.Variable, len(opts.exclusionPatterns))
		for i, pattern := range opts.exclusionPatterns {
			exclusionVars[i] = config.StringVariable(pattern)
		}
		vars["leader_index_exclusion_patterns"] = config.ListVariable(exclusionVars...)
	}

	if opts.followIndexPattern != "" {
		vars["follow_index_pattern"] = config.StringVariable(opts.followIndexPattern)
	}

	if opts.maxOutstandingRead > 0 {
		vars["max_outstanding_read_requests"] = config.IntegerVariable(opts.maxOutstandingRead)
	}

	return vars
}

func testAccCheckAutoFollowPatternActive(patternName string, wantActive bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		pattern, diags := esclient.GetAutoFollowPattern(ctx, client, patternName)
		if diags.HasError() {
			return fmt.Errorf("get auto-follow pattern %q: %v", patternName, diags)
		}
		if pattern == nil {
			return fmt.Errorf("auto-follow pattern %q not found in Elasticsearch", patternName)
		}
		if pattern.Active != wantActive {
			return fmt.Errorf("auto-follow pattern %q active=%t, want %t", patternName, pattern.Active, wantActive)
		}
		return nil
	}
}

func testAccCheckAutoFollowPatternExclusion(patternName, exclusion string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		pattern, diags := esclient.GetAutoFollowPattern(ctx, client, patternName)
		if diags.HasError() {
			return fmt.Errorf("get auto-follow pattern %q: %v", patternName, diags)
		}
		if pattern == nil {
			return fmt.Errorf("auto-follow pattern %q not found in Elasticsearch", patternName)
		}
		if !slices.Contains(pattern.LeaderIndexExclusionPatterns, exclusion) {
			return fmt.Errorf(
				"auto-follow pattern %q exclusion patterns %v do not include %q",
				patternName,
				pattern.LeaderIndexExclusionPatterns,
				exclusion,
			)
		}
		return nil
	}
}

func testAccCheckAutoFollowPatternFollowIndexPattern(patternName, want string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		pattern, diags := esclient.GetAutoFollowPattern(ctx, client, patternName)
		if diags.HasError() {
			return fmt.Errorf("get auto-follow pattern %q: %v", patternName, diags)
		}
		if pattern == nil || pattern.FollowIndexPattern == nil {
			return fmt.Errorf("auto-follow pattern %q has no follow_index_pattern", patternName)
		}
		if *pattern.FollowIndexPattern != want {
			return fmt.Errorf(
				"auto-follow pattern %q follow_index_pattern %q, want %q",
				patternName,
				*pattern.FollowIndexPattern,
				want,
			)
		}
		return nil
	}
}

func testAccCheckAutoFollowPatternLeaderIndexPatterns(patternName string, want ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		pattern, diags := esclient.GetAutoFollowPattern(ctx, client, patternName)
		if diags.HasError() {
			return fmt.Errorf("get auto-follow pattern %q: %v", patternName, diags)
		}
		if pattern == nil {
			return fmt.Errorf("auto-follow pattern %q not found in Elasticsearch", patternName)
		}

		for _, expected := range want {
			if !slices.Contains(pattern.LeaderIndexPatterns, expected) {
				return fmt.Errorf(
					"auto-follow pattern %q leader_index_patterns %v do not include %q",
					patternName,
					pattern.LeaderIndexPatterns,
					expected,
				)
			}
		}
		return nil
	}
}

func checkAutoFollowPatternDestroyed(patternName string) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		pattern, diags := esclient.GetAutoFollowPattern(ctx, client, patternName)
		if diags.HasError() {
			return fmt.Errorf("get auto-follow pattern %q: %v", patternName, diags)
		}
		if pattern != nil {
			return fmt.Errorf("auto-follow pattern %q still exists after destroy", patternName)
		}
		return nil
	}
}
