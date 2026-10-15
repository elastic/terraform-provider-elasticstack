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
	"os"
	"regexp"
	"strings"
	"testing"

	elasticsearchclient "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

const seriesAddr = "elasticstack_elasticsearch_rollover_index_series.test"

func TestAccRolloverIndexSeries_bootstrap(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "prefix", prefix),
					resource.TestCheckResourceAttr(seriesAddr, "rollover_alias", alias),
					resource.TestCheckResourceAttr(seriesAddr, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", bootstrap),
					checkSeriesID(alias),
					checkIndexJSONContains(bootstrap, "30s"),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_detachedGeneration(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	detached := prefix + "-000002"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, detached, nil)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", bootstrap),
					checkIndexPresent(detached),
					checkIndexHasNoAlias(detached, alias),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_indexNameCollision(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesPresent(bootstrap),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, bootstrap, nil)
				},
				ExpectError: regexp.MustCompile("Bootstrap index name is already in use"),
			},
		},
	})
}

func TestAccRolloverIndexSeries_aliasNameCollision(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	other := prefix + "-other"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesPresent(other),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, other, map[string]*bool{bootstrap: nil})
				},
				ExpectError: regexp.MustCompile("Bootstrap index name is already in use"),
			},
		},
	})
}

func TestAccRolloverIndexSeries_adoptHealthy(t *testing.T) {
	prefix, alias := accNames()
	writer := prefix + "-000003"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(writer),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, writer, map[string]*bool{alias: new(true)})
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", writer),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", writer),
					resource.TestCheckResourceAttr(seriesAddr, "prefix", prefix),
					resource.TestCheckResourceAttr(seriesAddr, "rollover_alias", alias),
					checkSeriesID(alias),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_adoptOutOfPrefix(t *testing.T) {
	prefix, alias := accNames()
	writer := prefix + "-000003"
	unrelated := "u" + strings.ToLower(sdkacctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")) + "-1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(writer, unrelated),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, writer, map[string]*bool{alias: new(true)})
					putIndex(t, unrelated, map[string]*bool{alias: new(false)})
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", writer),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "2"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", writer),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", unrelated),
					checkSeriesID(alias),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_rejectMissingWriter(t *testing.T) {
	prefix, alias := accNames()
	first := prefix + "-000001"
	second := prefix + "-000002"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesPresent(first, second),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, first, map[string]*bool{alias: nil})
					putIndex(t, second, map[string]*bool{alias: nil})
				},
				ExpectError: regexp.MustCompile("no effective writer"),
			},
		},
	})
}

func TestAccRolloverIndexSeries_rejectNonNumericWriter(t *testing.T) {
	prefix, alias := accNames()
	writer := prefix + "-current"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesPresent(writer),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, writer, map[string]*bool{alias: new(true)})
				},
				ExpectError: regexp.MustCompile("successor"),
			},
		},
	})
}

func TestAccRolloverIndexSeries_rejectDataStream(t *testing.T) {
	prefix, alias := accNames()
	dataStream := prefix + "-stream"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkDataStreamPresent(dataStream),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putDataStreamAlias(t, dataStream, alias, nil)
				},
				ExpectError: regexp.MustCompile("ordinary index aliases"),
			},
		},
	})
}

func TestAccRolloverIndexSeries_rejectNumericDataStream(t *testing.T) {
	prefix, alias := accNames()
	dataStream := prefix + "-000001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkDataStreamPresent(dataStream),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putDataStreamAlias(t, dataStream, alias, new(true))
				},
				ExpectError: regexp.MustCompile("ordinary index aliases"),
			},
		},
	})
}

func TestAccRolloverIndexSeries_externalRolloverDoesNotRepair(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	successor := prefix + "-000002"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap, successor),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", bootstrap),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					rolled := rolloverAlias(t, alias)
					require.Equal(t, successor, rolled)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", successor),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "2"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", bootstrap),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", successor),
					checkSeriesID(alias),
					checkWriteIndex(bootstrap, alias, false),
					checkWriteIndex(successor, alias, true),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_deletedBootstrapIsNotRecreated(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	successor := prefix + "-000002"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(successor),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					require.Equal(t, successor, rolloverAlias(t, alias))
					deleteIndex(t, bootstrap)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", successor),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", successor),
					checkSeriesID(alias),
					checkIndicesAbsent(bootstrap),
					checkWriteIndex(successor, alias, true),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_aliasAbsenceRebootstrapsDetachedData(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	detached := prefix + "-000002"
	t.Cleanup(func() {
		deleteIndex(t, detached)
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					require.Equal(t, detached, rolloverAlias(t, alias))
					deleteIndex(t, bootstrap)
					removeAlias(t, detached, alias)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", bootstrap),
					checkSeriesID(alias),
					checkIndexPresent(detached),
					checkIndexHasNoAlias(detached, alias),
					checkWriteIndex(bootstrap, alias, true),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_aliasAbsenceBootstrapCollision(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	var before string
	t.Cleanup(func() {
		deleteIndex(t, bootstrap)
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		CheckDestroy: func(*terraform.State) error {
			if err := requireIndex(bootstrap, true); err != nil {
				return err
			}
			if err := checkIndexHasNoAlias(bootstrap, alias)(nil); err != nil {
				return err
			}
			after, err := indexJSON(bootstrap)
			if err != nil {
				return err
			}
			if after != before {
				return fmt.Errorf("index %s changed after the failed create\nbefore %s\nafter %s", bootstrap, before, after)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					removeAlias(t, bootstrap, alias)
					raw, err := indexJSON(bootstrap)
					require.NoError(t, err)
					before = raw
				},
				ExpectError: regexp.MustCompile("Bootstrap index name is already in use"),
			},
		},
	})
}

func TestAccRolloverIndexSeries_prefixReplacementDestroysBeforeCreate(t *testing.T) {
	prefix, alias := accNames()
	replacement := prefix + "v2"
	bootstrap := prefix + "-000001"
	next := replacement + "-000001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(next),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					resource.TestCheckResourceAttr(seriesAddr, "rollover_alias", alias),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(replacement, alias),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "prefix", replacement),
					resource.TestCheckResourceAttr(seriesAddr, "rollover_alias", alias),
					resource.TestCheckResourceAttr(seriesAddr, "write_index", next),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", next),
					checkSeriesID(alias),
					checkIndicesAbsent(bootstrap),
					checkWriteIndex(next, alias, true),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_aliasReplacementDestroysBeforeCreate(t *testing.T) {
	prefix, alias := accNames()
	replacement := prefix + "-v2"
	bootstrap := prefix + "-000001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "rollover_alias", alias),
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, replacement),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "prefix", prefix),
					resource.TestCheckResourceAttr(seriesAddr, "rollover_alias", replacement),
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					checkSeriesID(replacement),
					checkWriteIndex(bootstrap, replacement, true),
					checkIndexHasNoAlias(bootstrap, alias),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_deletionProtectionUpdatesStateOnly(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	var before string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVarsWithProtection(prefix, alias, true, "1m"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(seriesAddr, "timeouts.update", "1m"),
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					checkWriteIndex(bootstrap, alias, true),
					func(*terraform.State) error {
						raw, err := indexJSON(bootstrap)
						before = raw
						return err
					},
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVarsWithProtection(prefix, alias, false, "4m"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(seriesAddr, "timeouts.update", "4m"),
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "1"),
					checkSeriesID(alias),
					checkWriteIndex(bootstrap, alias, true),
					func(*terraform.State) error {
						after, err := indexJSON(bootstrap)
						if err != nil {
							return err
						}
						if after != before {
							return fmt.Errorf("index %s changed during a state-only update\nbefore %s\nafter %s", bootstrap, before, after)
						}
						return nil
					},
				),
			},
		},
	})
}

func accVarsWithProtection(prefix, alias string, protection bool, updateTimeout string) config.Variables {
	vars := accVars(prefix, alias)
	vars["deletion_protection"] = config.BoolVariable(protection)
	vars["update_timeout"] = config.StringVariable(updateTimeout)
	return vars
}

func TestAccRolloverIndexSeries_missingWriterRefreshKeepsState(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	second := prefix + "-000002"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap, second),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					putIndex(t, second, map[string]*bool{alias: new(false)})
					setAliasWriteIndex(t, bootstrap, alias, false)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNullAttr(seriesAddr, "write_index"),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "2"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", bootstrap),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", second),
					resource.TestCheckResourceAttr(seriesAddr, "prefix", prefix),
					checkSeriesID(alias),
					checkWriteIndex(bootstrap, alias, false),
					checkWriteIndex(second, alias, false),
				),
			},
		},
	})
}

func TestAccRolloverIndexSeries_nonNumericWriterRefreshExposesName(t *testing.T) {
	prefix, alias := accNames()
	bootstrap := prefix + "-000001"
	current := prefix + "-current"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkIndicesAbsent(bootstrap, current),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", bootstrap),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          accVars(prefix, alias),
				PreConfig: func() {
					setAliasWriteIndex(t, bootstrap, alias, false)
					putIndex(t, current, map[string]*bool{alias: new(true)})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(seriesAddr, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(seriesAddr, "write_index", current),
					resource.TestCheckResourceAttr(seriesAddr, "indices.#", "2"),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", bootstrap),
					resource.TestCheckTypeSetElemAttr(seriesAddr, "indices.*", current),
					checkSeriesID(alias),
					checkWriteIndex(current, alias, true),
					checkWriteIndex(bootstrap, alias, false),
				),
			},
		},
	})
}

func accNames() (string, string) {
	prefix := "a" + sdkacctest.RandStringFromCharSet(10, "abcdefghijklmnopqrstuvwxyz")
	return prefix, prefix + "-write"
}

func accVars(prefix, alias string) config.Variables {
	return config.Variables{
		"prefix":         config.StringVariable(prefix),
		"rollover_alias": config.StringVariable(alias),
	}
}

func mustTypedClient(t *testing.T) *elasticsearchclient.TypedClient {
	t.Helper()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	require.NoError(t, err)
	return client.GetESClient()
}

func putIndex(t *testing.T, name string, aliases map[string]*bool) {
	t.Helper()
	es := mustTypedClient(t)
	body := map[string]any{
		"settings": map[string]any{"number_of_replicas": 0},
	}
	if aliases != nil {
		encoded := make(map[string]any, len(aliases))
		for alias, write := range aliases {
			spec := map[string]any{}
			if write != nil {
				spec["is_write_index"] = *write
			}
			encoded[alias] = spec
		}
		body["aliases"] = encoded
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	_, err = es.Indices.Create(name).Raw(bytes.NewReader(raw)).Do(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = es.Indices.Delete(name).IgnoreUnavailable(true).Do(context.Background())
	})
}

func putDataStreamAlias(t *testing.T, dataStream, alias string, write *bool) {
	t.Helper()
	es := mustTypedClient(t)
	templateName := dataStream + "-template"
	template, err := json.Marshal(map[string]any{
		"index_patterns": []string{dataStream},
		"priority":       500,
		"data_stream":    map[string]any{},
		"template": map[string]any{
			"settings": map[string]any{"number_of_replicas": 0},
		},
	})
	require.NoError(t, err)
	_, err = es.Indices.PutIndexTemplate(templateName).Raw(bytes.NewReader(template)).Do(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = es.Indices.DeleteIndexTemplate(templateName).Do(context.Background())
	})

	_, err = es.Indices.CreateDataStream(dataStream).Do(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = es.Indices.DeleteDataStream(dataStream).Do(context.Background())
	})

	action := map[string]any{"index": dataStream, "alias": alias}
	if write != nil {
		action["is_write_index"] = *write
	}
	aliasBody, err := json.Marshal(map[string]any{
		"actions": []any{map[string]any{"add": action}},
	})
	require.NoError(t, err)
	_, err = es.Indices.UpdateAliases().Raw(bytes.NewReader(aliasBody)).Do(context.Background())
	require.NoError(t, err)
}

func checkSeriesID(alias string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}
		uuid, diags := client.ClusterID(context.Background())
		if diags.HasError() {
			return fmt.Errorf("cluster uuid: %v", diags)
		}
		return resource.TestCheckResourceAttr(seriesAddr, "id", *uuid+"/"+alias)(state)
	}
}

func indexJSON(name string) (string, error) {
	es, err := typedClient()
	if err != nil {
		return "", err
	}
	indices, err := es.Indices.Get(name).FlatSettings(true).Do(context.Background())
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(indices[name])
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func checkIndexJSONContains(name, needle string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		raw, err := indexJSON(name)
		if err != nil {
			return err
		}
		if !strings.Contains(raw, needle) {
			return fmt.Errorf("index %s json %s does not contain %s", name, raw, needle)
		}
		return nil
	}
}

func checkIndexPresent(name string) resource.TestCheckFunc {
	return checkIndicesPresent(name)
}

func checkIndicesPresent(names ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		for _, name := range names {
			if err := requireIndex(name, true); err != nil {
				return err
			}
		}
		return nil
	}
}

func checkIndicesAbsent(names ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		for _, name := range names {
			if err := requireIndex(name, false); err != nil {
				return err
			}
		}
		return nil
	}
}

func requireIndex(name string, present bool) error {
	es, err := typedClient()
	if err != nil {
		return err
	}
	_, err = es.Indices.Get(name).Do(context.Background())
	if present {
		if err != nil {
			return fmt.Errorf("index %s: %w", name, err)
		}
		return nil
	}
	if err == nil {
		return fmt.Errorf("index %s still exists", name)
	}
	if esclient.IsNotFoundElasticsearchError(err) {
		return nil
	}
	return err
}

func rolloverAlias(t *testing.T, alias string) string {
	t.Helper()
	es := mustTypedClient(t)
	res, err := es.Indices.Rollover(alias).Do(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res)
	return res.NewIndex
}

func deleteIndex(t *testing.T, name string) {
	t.Helper()
	es := mustTypedClient(t)
	_, err := es.Indices.Delete(name).IgnoreUnavailable(true).Do(context.Background())
	require.NoError(t, err)
}

func setAliasWriteIndex(t *testing.T, indexName, alias string, write bool) {
	t.Helper()
	es := mustTypedClient(t)
	body, err := json.Marshal(map[string]any{
		"actions": []any{map[string]any{"add": map[string]any{
			"index":          indexName,
			"alias":          alias,
			"is_write_index": write,
		}}},
	})
	require.NoError(t, err)
	_, err = es.Indices.UpdateAliases().Raw(bytes.NewReader(body)).Do(context.Background())
	require.NoError(t, err)
}

func checkNullAttr(addr, name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource %s is missing", addr)
		}
		if value, exists := rs.Primary.Attributes[name]; exists && value != "" {
			return fmt.Errorf("%s = %q, want null", name, value)
		}
		return nil
	}
}

func removeAlias(t *testing.T, indexName, alias string) {
	t.Helper()
	es := mustTypedClient(t)
	_, err := es.Indices.DeleteAlias(indexName, alias).Do(context.Background())
	require.NoError(t, err)
}

func checkWriteIndex(indexName, alias string, write bool) resource.TestCheckFunc {
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
		if definition.IsWriteIndex == nil || *definition.IsWriteIndex != write {
			return fmt.Errorf("index %s alias %s is_write_index = %v, want %v", indexName, alias, definition.IsWriteIndex, write)
		}
		return nil
	}
}

func checkIndexHasNoAlias(indexName, alias string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		indices, err := es.Indices.Get(indexName).Do(context.Background())
		if err != nil {
			return err
		}
		if _, ok := indices[indexName].Aliases[alias]; ok {
			return fmt.Errorf("index %s has alias %s", indexName, alias)
		}
		return nil
	}
}

func checkDataStreamPresent(name string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		es, err := typedClient()
		if err != nil {
			return err
		}
		res, err := es.Indices.GetDataStream().Name(name).Do(context.Background())
		if err != nil {
			return err
		}
		if res == nil || len(res.DataStreams) == 0 {
			return fmt.Errorf("data stream %s is missing", name)
		}
		return nil
	}
}

func typedClient() (*elasticsearchclient.TypedClient, error) {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return nil, err
	}
	return client.GetESClient(), nil
}
