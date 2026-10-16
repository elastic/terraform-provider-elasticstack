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

package rolloverindexseries_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Live policy cleanup on Elasticsearch 9.5.5. CI matrix versions were not run locally.
func TestAccRolloverIndexSeries_unsharedPolicyDeletesAfterSeries(t *testing.T) {
	prefix, alias := accNames()
	policyName := prefix + "-policy"
	bootstrap := prefix + "-000001"
	templateName := prefix + "-template"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			checkIndicesAbsent(bootstrap),
			checkPolicyAbsent(policyName),
		),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accPolicyVars(prefix, alias, policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "1"),
					checkIndexJSONContains(bootstrap, policyName),
					checkIndexJSONContains(bootstrap, alias),
					checkIndexJSONContains(bootstrap, "order_id"),
					checkIndexJSONContains(bootstrap, "30s"),
					checkTemplateOmitsRolloverAlias(templateName, alias),
					checkPolicyDeleteRejected(policyName),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_retainedPolicyReferenceBlocksDeletion(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}
	prefix, alias := accNames()
	policyName := prefix + "-policy"
	bootstrap := prefix + "-000001"
	kept := "kept" + strings.ToLower(sdkacctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz"))
	t.Cleanup(func() {
		deleteIndex(t, kept)
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkPolicyAbsent(policyName),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accPolicyVars(prefix, alias, policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					checkIndexPresent(bootstrap),
					checkPolicyDeleteRejected(policyName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_series"),
				ConfigVariables:          accPolicyVars(prefix, alias, policyName),
				PreConfig: func() {
					putIndexWithSettings(t, kept, map[string]any{
						"number_of_replicas":   0,
						"index.lifecycle.name": policyName,
					}, nil)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIndicesAbsent(bootstrap),
					checkIndexPresent(kept),
					checkIndexHasNoAlias(kept, alias),
					checkIndexJSONContains(kept, policyName),
					checkPolicyDeleteRejected(policyName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_series"),
				ConfigVariables:          accPolicyVars(prefix, alias, policyName),
				PreConfig: func() {
					deleteIndex(t, kept)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIndicesAbsent(bootstrap, kept),
					checkPolicyPresent(policyName),
				),
			},
		},
	})
}

// Live ILM shrink on Elasticsearch 9.5.5. _ilm/move is sent only for the fixture
// bootstrap index. Searchable snapshots are not executed here.
func TestAccRolloverIndexSeries_ilmShrinkRetainsAliasAndDeletesReplacement(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}
	prefix, alias := accNames()
	policyName := prefix + "-policy"
	bootstrap := prefix + "-000001"
	successor := prefix + "-000002"
	templateName := prefix + "-template"
	var shrinkName string
	t.Cleanup(func() {
		deleteIndex(t, bootstrap)
		deleteIndex(t, successor)
		if shrinkName != "" {
			deleteIndex(t, shrinkName)
		}
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		CheckDestroy: func(state *terraform.State) error {
			names := []string{bootstrap, successor}
			if shrinkName != "" {
				names = append(names, shrinkName)
			}
			if err := checkIndicesAbsent(names...)(state); err != nil {
				return err
			}
			return checkPolicyAbsent(policyName)(state)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accPolicyVars(prefix, alias, policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					checkFlatSetting(bootstrap, "index.number_of_shards", "2"),
					checkIndexJSONContains(bootstrap, "order_id"),
					checkIndexJSONContains(bootstrap, policyName),
					checkTemplateOmitsRolloverAlias(templateName, alias),
					checkSeriesID(alias),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accPolicyVars(prefix, alias, policyName),
				PreConfig: func() {
					shrinkName = shrinkOwnedIndex(t, alias, bootstrap)
				},
				Check: func(state *terraform.State) error {
					if !strings.HasPrefix(shrinkName, "shrink-") || !strings.Contains(shrinkName, bootstrap) {
						return fmt.Errorf("shrink replacement %q does not keep the source name", shrinkName)
					}
					return resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(seriesAddr, "write_index", successor),
						resource.TestCheckResourceAttr(seriesAddr, "indices.#", "2"),
						resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", shrinkName),
						resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", successor),
						checkConcreteIndexAbsent(bootstrap),
						checkIndexPresent(shrinkName),
						checkFlatSetting(shrinkName, "index.number_of_shards", "1"),
						checkWriteIndex(successor, alias, true),
						checkNonWriteAliasMember(shrinkName, alias),
						checkSeriesID(alias),
					)(state)
				},
			},
		},
	})
}

func accPolicyVars(prefix, alias, policyName string) config.Variables {
	return config.Variables{
		"prefix":         config.StringVariable(prefix),
		"rollover_alias": config.StringVariable(alias),
		"policy_name":    config.StringVariable(policyName),
	}
}

func checkTemplateOmitsRolloverAlias(templateName, alias string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		res, err := es.Indices.GetIndexTemplate().Name(templateName).Perform(context.Background())
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if res.StatusCode >= 300 {
			return fmt.Errorf("get index template %s: %s", templateName, body)
		}
		var decoded struct {
			IndexTemplates []struct {
				IndexTemplate struct {
					Template struct {
						Aliases map[string]json.RawMessage `json:"aliases"`
					} `json:"template"`
				} `json:"index_template"`
			} `json:"index_templates"`
		}
		if err := json.Unmarshal(body, &decoded); err != nil {
			return err
		}
		if len(decoded.IndexTemplates) != 1 {
			return fmt.Errorf("template %s count = %d, want 1", templateName, len(decoded.IndexTemplates))
		}
		aliases := decoded.IndexTemplates[0].IndexTemplate.Template.Aliases
		if _, ok := aliases[alias]; ok {
			return fmt.Errorf("template %s defines rollover alias %s in aliases", templateName, alias)
		}
		return nil
	}
}

func checkPolicyDeleteRejected(name string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		_, err = es.Ilm.DeleteLifecycle(name).Do(context.Background())
		if err == nil {
			return fmt.Errorf("deleted in-use policy %s", name)
		}
		if !strings.Contains(strings.ToLower(err.Error()), "in use") {
			return fmt.Errorf("policy %s delete error %w, want in use", name, err)
		}
		if _, getErr := es.Ilm.GetLifecycle().Policy(name).Do(context.Background()); getErr != nil {
			return fmt.Errorf("policy %s missing after rejected delete: %w", name, getErr)
		}
		return nil
	}
}

func checkPolicyPresent(name string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		if _, err = es.Ilm.GetLifecycle().Policy(name).Do(context.Background()); err != nil {
			return fmt.Errorf("ilm policy %s: %w", name, err)
		}
		return nil
	}
}

func checkPolicyAbsent(name string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		_, err = es.Ilm.GetLifecycle().Policy(name).Do(context.Background())
		if err == nil {
			return fmt.Errorf("ilm policy %s still exists", name)
		}
		if esclient.IsNotFoundElasticsearchError(err) {
			return nil
		}
		return err
	}
}

func checkFlatSetting(name, key, want string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		res, err := es.Indices.Get(name).FlatSettings(true).Perform(context.Background())
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if res.StatusCode >= 300 {
			return fmt.Errorf("get index %s: %s", name, body)
		}
		var decoded map[string]struct {
			Settings map[string]any `json:"settings"`
		}
		if err := json.Unmarshal(body, &decoded); err != nil {
			return err
		}
		got := fmt.Sprint(decoded[name].Settings[key])
		if got != want {
			return fmt.Errorf("index %s setting %s = %s, want %s", name, key, got, want)
		}
		return nil
	}
}

func checkNonWriteAliasMember(indexName, alias string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		indices, err := es.Indices.Get(indexName).Do(context.Background())
		if err != nil {
			return err
		}
		definition, ok := indices[indexName].Aliases[alias]
		if !ok {
			return fmt.Errorf("index %s has no alias %s", indexName, alias)
		}
		if definition.IsWriteIndex != nil && *definition.IsWriteIndex {
			return fmt.Errorf("index %s is the write index for %s", indexName, alias)
		}
		return nil
	}
}

type ilmPosition struct {
	Phase           string
	Action          string
	Step            string
	StepInfo        string
	ShrinkIndexName string
	Missing         bool
}

func shrinkOwnedIndex(t *testing.T, alias, original string) string {
	t.Helper()
	issuedRollover := false
	deadline := time.Now().Add(3 * time.Minute)
	var last string
	for {
		if replacement, ok := ownedShrinkReplacement(t, alias, original); ok {
			return replacement
		}
		pos := explainOwned(t, original)
		requireILM(t, pos.Missing || pos.Step != "ERROR", "shrink of %s failed: %s", original, pos.StepInfo)
		switch {
		case !issuedRollover && !pos.Missing && pos.Action != "shrink" && pos.Phase != "warm":
			if err := moveOwnedILM(t, original, "hot", "rollover", "attempt-rollover"); err != nil {
				t.Fatalf("rollover move for %s: %v", original, err)
			}
			issuedRollover = true
		case !pos.Missing && pos.Step == "wait-for-shard-history-leases":
			if err := moveOwnedILM(t, original, "warm", "shrink", "readonly"); err != nil && !strings.Contains(err.Error(), "not on current step") {
				t.Fatalf("readonly move for %s: %v", original, err)
			}
		}
		last = fmt.Sprintf("members pending explain=%+v", pos)
		requireILM(t, time.Now().Before(deadline), "shrink replacement for %s did not join %s: %s", original, alias, last)
		time.Sleep(200 * time.Millisecond)
	}
}

func moveOwnedILM(t *testing.T, indexName, phase, action, name string) error {
	t.Helper()
	current := explainOwned(t, indexName)
	if current.Missing {
		return fmt.Errorf("index %s disappeared before _ilm/move", indexName)
	}
	if current.Phase == phase && current.Action == action && current.Step == name {
		return nil
	}
	if current.Step == "ERROR" {
		return fmt.Errorf("index %s ILM error before move: %s", indexName, current.StepInfo)
	}
	body, err := json.Marshal(map[string]any{
		"current_step": map[string]string{
			"phase":  current.Phase,
			"action": current.Action,
			"name":   current.Step,
		},
		"next_step": map[string]string{
			"phase":  phase,
			"action": action,
			"name":   name,
		},
	})
	if err != nil {
		return err
	}
	es := mustTypedClient(t)
	_, err = es.Ilm.MoveToStep(indexName).Raw(bytes.NewReader(body)).Do(context.Background())
	if err != nil {
		return fmt.Errorf("_ilm/move %s to %s/%s/%s: %w", indexName, phase, action, name, err)
	}
	return nil
}

func ownedShrinkReplacement(t *testing.T, alias, original string) (string, bool) {
	t.Helper()
	var shrinkName string
	for name := range aliasMemberNames(t, alias) {
		if strings.HasPrefix(name, "shrink-") && strings.Contains(name, original) {
			shrinkName = name
		}
	}
	if shrinkName == "" || !indexIsAbsent(t, original) {
		return "", false
	}
	return shrinkName, true
}

func explainOwned(t *testing.T, indexName string) ilmPosition {
	t.Helper()
	es := mustTypedClient(t)
	res, err := es.Ilm.ExplainLifecycle(indexName).Perform(context.Background())
	requireILM(t, err == nil, "explain %s: %v", indexName, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	requireILM(t, err == nil, "read explain %s: %v", indexName, err)
	if res.StatusCode == 404 {
		return ilmPosition{Missing: true}
	}
	requireILM(t, res.StatusCode < 300, "explain %s: %s", indexName, body)
	var decoded struct {
		Indices map[string]struct {
			Phase           string         `json:"phase"`
			Action          string         `json:"action"`
			Step            string         `json:"step"`
			StepInfo        map[string]any `json:"step_info"`
			ShrinkIndexName string         `json:"shrink_index_name"`
		} `json:"indices"`
	}
	requireILM(t, json.Unmarshal(body, &decoded) == nil, "decode explain %s: %s", indexName, body)
	entry, ok := decoded.Indices[indexName]
	if !ok {
		return ilmPosition{Missing: true}
	}
	info, err := json.Marshal(entry.StepInfo)
	requireILM(t, err == nil, "encode step info: %v", err)
	return ilmPosition{
		Phase:           entry.Phase,
		Action:          entry.Action,
		Step:            entry.Step,
		StepInfo:        string(info),
		ShrinkIndexName: entry.ShrinkIndexName,
	}
}

func aliasMemberNames(t *testing.T, alias string) map[string]struct{} {
	t.Helper()
	es := mustTypedClient(t)
	res, err := es.Indices.GetAlias().Name(alias).Do(context.Background())
	requireILM(t, err == nil, "get alias %s: %v", alias, err)
	names := make(map[string]struct{}, len(res))
	for name := range res {
		names[name] = struct{}{}
	}
	return names
}

func indexIsAbsent(t *testing.T, name string) bool {
	t.Helper()
	err := concreteIndexAbsent(name)
	if err == nil {
		return true
	}
	if strings.Contains(err.Error(), "still exists") {
		return false
	}
	t.Fatalf("get index %s: %v", name, err)
	return false
}

func checkConcreteIndexAbsent(name string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		return concreteIndexAbsent(name)
	}
}

func concreteIndexAbsent(name string) error {
	es, err := typedClient()
	if err != nil {
		return err
	}
	res, err := es.Indices.Get(name).Do(context.Background())
	if esclient.IsNotFoundElasticsearchError(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, ok := res[name]; ok {
		return fmt.Errorf("concrete index %s still exists", name)
	}
	return nil
}

func requireILM(t *testing.T, ok bool, format string, args ...any) {
	t.Helper()
	if !ok {
		t.Fatalf(format, args...)
	}
}
