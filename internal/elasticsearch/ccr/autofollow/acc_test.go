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

func TestAccResourceCCRAutoFollowPattern_basic(t *testing.T) {
	ccrEnv := acctest.PreCheckCCR(t)
	leaderIndexName := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	patternName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	leaderPattern := leaderIndexName[:4] + "*"
	vars := ccrAutoFollowVariables(ccrEnv, leaderIndexName, patternName, leaderPattern, nil)
	varsWithExclusion := ccrAutoFollowVariables(ccrEnv, leaderIndexName, patternName, leaderPattern, []string{leaderIndexName})

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
					resource.TestCheckTypeSetElemAttr(autoFollowResourceName, "leader_index_patterns.*", leaderPattern),
					testAccCheckAutoFollowPatternActive(patternName, true),
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
	vars := ccrAutoFollowVariables(ccrEnv, leaderIndexName, patternName, leaderPattern, nil)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheckCCR(t) },
		CheckDestroy: checkAutoFollowPatternDestroyed(patternName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_inactive"),
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

func ccrAutoFollowVariables(
	ccrEnv acctest.CCRTestEnv,
	leaderIndexName, patternName, leaderPattern string,
	exclusionPatterns []string,
) config.Variables {
	vars := config.Variables{
		"remote_cluster_alias": config.StringVariable(ccrEnv.RemoteClusterAlias),
		"remote_proxy_address": config.StringVariable(ccrEnv.RemoteProxyAddress),
		"leader_index_name":    config.StringVariable(leaderIndexName),
		"pattern_name":         config.StringVariable(patternName),
		"leader_index_pattern": config.StringVariable(leaderPattern),
	}
	if len(exclusionPatterns) > 0 {
		exclusionVars := make([]config.Variable, len(exclusionPatterns))
		for i, pattern := range exclusionPatterns {
			exclusionVars[i] = config.StringVariable(pattern)
		}
		vars["leader_index_exclusion_patterns"] = config.ListVariable(exclusionVars...)
	}
	return vars
}

func testAccCheckAutoFollowPatternActive(patternName string, wantActive bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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
	return func(s *terraform.State) error {
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

func checkAutoFollowPatternDestroyed(patternName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
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
