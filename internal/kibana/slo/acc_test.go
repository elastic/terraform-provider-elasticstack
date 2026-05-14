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

package slo_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var sloTimesliceMetricsMinVersion = version.Must(version.NewVersion("8.12.0"))

// skipKqlSLOOrSettingsSyncFieldUnsupported gates acceptance steps that apply settings.sync_field
// (and the enabled field exercised together with it). Kibana <8.18 rejects that key with HTTP 400
// (excess keys: body.settings.syncField). Plan-only and tests that omit sync_field use
// CheckIfVersionMeetsConstraints(SLOKqlAccTestConstraints) only.
func skipKqlSLOOrSettingsSyncFieldUnsupported() (bool, error) {
	if skip, err := versionutils.CheckIfVersionMeetsConstraints(slo.SLOKqlAccTestConstraints)(); err != nil || skip {
		return skip, err
	}
	return versionutils.CheckIfVersionIsUnsupported(slo.SLOSettingsSyncFieldMinVersion)()
}

func TestAccResourceSlo(t *testing.T) {
	// This test exposes a bug in Kibana present in 8.11.x
	slo8_9Constraints, err := version.NewConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
	require.NoError(t, err)

	slo8_10Constraints, err := version.NewConstraint(">=8.10.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
	require.NoError(t, err)

	for _, testWithDataViewID := range []bool{true, false} {
		t.Run("with-data-view-id="+fmt.Sprint(testWithDataViewID), func(t *testing.T) {
			dataviewCheckFunc := func(indicator string) resource.TestCheckFunc {
				if !testWithDataViewID {
					return func(_ *terraform.State) error {
						return nil
					}
				}

				return resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", indicator+".0.data_view_id", "my-data-view-id")
			}
			withOptionalDataViewID := func(vars config.Variables) config.Variables {
				if testWithDataViewID {
					vars["data_view_id"] = config.StringVariable("my-data-view-id")
				}
				return vars
			}
			sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { acctest.PreCheck(t) },
				CheckDestroy: checkResourceSloDestroy,
				Steps: []resource.TestStep{
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("apm_latency_indicator"),
						ConfigVariables: config.Variables{
							"name": config.StringVariable(sloName),
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "slo_id", "id-"+sloName),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "description", "fully sick SLO"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.environment", "production"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.service", "my-service"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_type", "request"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_name", "GET /sup/dawg"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.index", "my-index-"+sloName),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.threshold", "500"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.duration", "7d"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.type", "rolling"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "budgeting_method", "timeslices"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.target", "0.999"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_target", "0.95"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_window", "5m"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
						),
					},
					{
						// check that name can be updated
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("update_name"),
						ConfigVariables: config.Variables{
							"name": config.StringVariable(fmt.Sprintf("updated-%s", sloName)),
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", fmt.Sprintf("updated-%s", sloName)),
						),
					},
					{ // check that settings can be updated from api-computed defaults
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("update_settings"),
						ConfigVariables: config.Variables{
							"name": config.StringVariable(sloName),
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.sync_delay", "5m"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.frequency", "5m"),
						),
					},
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("apm_availability_indicator"),
						ConfigVariables: config.Variables{
							"name": config.StringVariable(sloName),
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.environment", "production"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.service", "my-service"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.transaction_type", "request"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.transaction_name", "GET /sup/dawg"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.index", "my-index-"+sloName),
						),
					},
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("kql_custom_indicator"),
						ConfigVariables: withOptionalDataViewID(config.Variables{
							"name": config.StringVariable(sloName),
						}),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.index", "my-index-"+sloName),
							dataviewCheckFunc("kql_custom_indicator"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.good", "latency < 300"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.total", "*"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.filter", "labels.groupId: group-0"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.timestamp_field", "custom_timestamp"),
						),
					},
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("histogram_custom_indicator"),
						ConfigVariables: withOptionalDataViewID(config.Variables{
							"name": config.StringVariable(sloName),
						}),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.index", "my-index-"+sloName),
							dataviewCheckFunc("histogram_custom_indicator"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.field", "test"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.aggregation", "value_count"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.filter", "latency < 300"),
							resource.TestCheckNoResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.from"),
							resource.TestCheckNoResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.to"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.total.0.field", "test"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.total.0.aggregation", "value_count"),
						),
					},
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("metric_custom_indicator_group_by"),
						ConfigVariables: withOptionalDataViewID(config.Variables{
							"name": config.StringVariable(sloName),
						}),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index-"+sloName),
							dataviewCheckFunc("metric_custom_indicator"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.name", "A"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.aggregation", "sum"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.field", "processor.processed"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.name", "B"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.aggregation", "sum"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.field", "processor.processed"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.equation", "A + B"),

							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.name", "A"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.aggregation", "sum"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.field", "processor.accepted"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.name", "B"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.aggregation", "sum"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.field", "processor.accepted"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.equation", "A + B"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.#", "1"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.0", "some.field"),
						),
					},
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("metric_custom_indicator_tags"),
						ConfigVariables: withOptionalDataViewID(config.Variables{
							"name": config.StringVariable(sloName),
						}),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "tags.0", "tag-1"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "tags.1", "another_tag"),
						),
					},
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionIsUnsupported(sloTimesliceMetricsMinVersion)()
							}

							return versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("timeslice_metric_indicator"),
						ConfigVariables: withOptionalDataViewID(config.Variables{
							"name": config.StringVariable(sloName),
						}),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.index", "my-index-"+sloName),
							dataviewCheckFunc("timeslice_metric_indicator"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.name", "A"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.aggregation", "sum"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.equation", "A"),
						),
					},
				},
			})
		})
	}
}

//go:embed testdata/TestAccResourceSloGroupBy/single_element/main.tf
var singleElementConfig string

func TestAccResourceSloGroupBy(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	// The empty group_by step test exposes a bug in Kibana present in 8.11.x
	slo8_10Constraints, err := version.NewConstraint(">=8.10.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				// Create the SLO with the last provider version enforcing single element group_by
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.11",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsMultipleGroupByMinVersion),
				Config:   singleElementConfig,
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.field", "processor.processed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.field", "processor.processed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.equation", "A + B"),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.field", "processor.accepted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.field", "processor.accepted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.equation", "A + B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by", "some.field"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(slo.SLOSupportsMultipleGroupByMinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multiple_elements"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.field", "processor.processed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.field", "processor.processed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.equation", "A + B"),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.field", "processor.accepted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.field", "processor.accepted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.equation", "A + B"),
					// resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.0", "some.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.1", "some.other.field"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_list"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("star"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by.0", "*"),
				),
			},
		},
	})
}

func TestAccResourceSloPreventInitialBackfill(t *testing.T) {
	versionutils.SkipIfUnsupported(t, slo.SLOSupportsPreventInitialBackfillMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index-"+sloName),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.prevent_initial_backfill", "true"),
				),
			},
		},
	})
}

func TestAccResourceSlo_timeslice_metric_indicator_basic(t *testing.T) {
	versionutils.SkipIfUnsupported(t, sloTimesliceMetricsMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.filter", "status_code: 200"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.field", "latency"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.equation", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.comparator", "GT"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.threshold", "100"),
				),
			},
		},
	})
}

func TestAccResourceSlo_timeslice_metric_indicator_percentile(t *testing.T) {
	versionutils.SkipIfUnsupported(t, sloTimesliceMetricsMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.aggregation", "percentile"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.percentile", "99"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.equation", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.comparator", "LT"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.threshold", "200"),
				),
			},
		},
	})
}

func TestAccResourceSlo_timeslice_metric_indicator_doc_count(t *testing.T) {
	versionutils.SkipIfUnsupported(t, sloTimesliceMetricsMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.name", "C"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.aggregation", "doc_count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.filter", "field: value"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.equation", "C"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.comparator", "GTE"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.threshold", "10"),
				),
			},
		},
	})
}

func TestAccResourceSlo_timeslice_metric_indicator_multiple_mixed_metrics(t *testing.T) {
	versionutils.SkipIfUnsupported(t, sloTimesliceMetricsMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.aggregation", "avg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.0.field", "bops"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.1.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.1.aggregation", "percentile"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.1.percentile", "99"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.2.name", "C"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.metrics.2.aggregation", "doc_count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.equation", "A + B + C"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.comparator", "GT"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.metric.0.threshold", "100"),
				),
			},
		},
	})
}

func TestAccResourceSlo_metric_custom_indicator_doc_count(t *testing.T) {
	versionutils.SkipIfUnsupported(t, sloTimesliceMetricsMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.aggregation", "doc_count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.filter", "status: 200"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.equation", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.aggregation", "doc_count"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.filter"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.equation", "B"),
				),
			},
		},
	})
}

func TestAccResourceSloErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multiple_indicators"),
				ExpectError: regexp.MustCompile(
					"(?s)Invalid Attribute Combination.*?Exactly one of these attributes must be configured:\\s+" +
						regexp.QuoteMeta(`[metric_custom_indicator,histogram_custom_indicator,apm_latency_indicator,apm_availability_indicator,kql_custom_indicator,timeslice_metric_indicator]`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.10.0-SNAPSHOT"))),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("agg_fail"),
				ExpectError: regexp.MustCompile(
					regexp.QuoteMeta(`Attribute histogram_custom_indicator[0].good[0].aggregation value must be one`) +
						"\\s+" +
						regexp.QuoteMeta(`of: ["value_count" "range"], got: "supdawg"`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("budget_fail"),
				ExpectError:              regexp.MustCompile(regexp.QuoteMeta(`Attribute budgeting_method value must be one of: ["occurrences"`) + "\\s+" + regexp.QuoteMeta(`"timeslices"], got: "supdawg"`)),
			},
		},
	})
}

func TestAccResourceSloValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("short"),
				ConfigVariables: config.Variables{
					"name":   config.StringVariable("short"),
					"slo_id": config.StringVariable("sh"),
				},
				ExpectError: regexp.MustCompile(`Attribute slo_id string length must be between 8 and 48, got: 2`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("toolongid"),
				ConfigVariables: config.Variables{
					"name":   config.StringVariable("toolongid"),
					"slo_id": config.StringVariable("this-id-is-way-too-long-and-exceeds-the-48-character-limit-for-slo-ids"),
				},
				ExpectError: regexp.MustCompile(`Attribute slo_id string length must be between 8 and 48, got: 70`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalidchars"),
				ConfigVariables: config.Variables{
					"name":   config.StringVariable("invalidchars"),
					"slo_id": config.StringVariable("invalid@id$"),
				},
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`Attribute slo_id must contain only letters, numbers, hyphens, and`) + "\\s+" + regexp.QuoteMeta(`underscores, got: invalid@id$`)),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("kql_good_and_good_kql"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("kql-dup"),
				},
				// Fires on either the string or object arm of the KQL exclusive pair.
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Combination.*good`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("time_window_invalid_type"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("tw-invalid"),
				},
				ExpectError: regexp.MustCompile(`(?s)time_window\[[0-9]\]\.type.*not_a_valid_type`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("time_window_invalid_duration_rolling"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("tw-dur-rolling"),
				},
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*duration`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("time_window_invalid_duration_calendar"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("tw-dur-calendar"),
				},
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*duration`),
			},
		},
	})
}

// TestAccResourceSlo_kql_object_form_and_settings_enabled exercises:
//   - plan-only: object-form filter_kql / good_kql and settings.sync_field parse (no Kibana create);
//   - apply (>=8.18 only): string-form KQL with sync_field and toggling `enabled` (Kibana accepts
//     body.settings.syncField from 8.18+; older stacks 400 on apply).
//
// The apply path uses string KQL, not a full create with object-form *_kql only: some Kibana/Stack
// versions respond with HTTP 400 for object-only indicator payloads (e.g. expecting a string
// for /indicator/good). Plan-only still validates that *_kql attributes parse and plan. Full
// object-form create+read is covered by unit tests in this package.
func TestAccResourceSlo_kql_object_form_and_settings_enabled(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	skipKqlSLOStack := versionutils.CheckIfVersionMeetsConstraints(slo.SLOKqlAccTestConstraints)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKqlSLOStack,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("kql_object_form_planonly"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKqlSLOOrSettingsSyncFieldUnsupported,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_sync_string_kql"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(sloName),
					"enabled": config.BoolVariable(false),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.filter", "service.name: test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.good", "latency < 300"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.sync_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "enabled", "false"),
					checkSloAPIEnabled(false),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKqlSLOOrSettingsSyncFieldUnsupported,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_sync_string_kql"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(sloName),
					"enabled": config.BoolVariable(true),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.sync_field", "@timestamp"),
					checkSloAPIEnabled(true),
				),
			},
		},
	})
}

// TestAccResourceSlo_kql_custom_indicator_basic uses string KQL only (no timeslice indicator).
// Step 1 skips with SLOKqlAccTestConstraints (8.9+, excluding 8.11.x). Step 2 (Fleet-style config
// with group_by) requires SLOKqlFleetAccTestConstraints (8.10+, same 8.11 exclusions), not 8.12
// timeslice.
func TestAccResourceSlo_kql_custom_indicator_basic(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	skipKqlSLO := versionutils.CheckIfVersionMeetsConstraints(slo.SLOKqlAccTestConstraints)
	skipKqlSLOFleetStep := versionutils.CheckIfVersionMeetsConstraints(slo.SLOKqlFleetAccTestConstraints)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKqlSLO,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.filter", "*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.good", "latency < 300"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.total", "*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.timestamp_field", "custom_timestamp"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipKqlSLOFleetStep,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("fleetctl_test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
					"tags": config.ListVariable(config.StringVariable("test-tag")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.fleetctl_api_pod_readiness", "kql_custom_indicator.0.index", "metrics-*,serverless-metrics-*:metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.fleetctl_api_pod_readiness", "kql_custom_indicator.0.good", "kubernetes.pod.status.ready: true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.fleetctl_api_pod_readiness", "kql_custom_indicator.0.total", ""),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_slo.fleetctl_api_pod_readiness",
						"kql_custom_indicator.0.filter",
						"kubernetes.deployment.name: \"fleetctl-api\" and kubernetes.pod.status.ready : * ",
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.fleetctl_api_pod_readiness", "kql_custom_indicator.0.timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.fleetctl_api_pod_readiness", "settings.sync_delay", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.fleetctl_api_pod_readiness", "settings.frequency", "1m"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceSloFromSDK/create/main.tf
var sloFromSDKCreateConfig string

func TestAccResourceSloFromSDK(t *testing.T) {
	// This test exposes a bug in Kibana present in 8.11.x
	sloConstraints, err := version.NewConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
	require.NoError(t, err)

	versionutils.SkipIfUnsupportedConstraints(t, sloConstraints, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				// Create the SLO with the last provider version where the SLO resource was built on the SDK.
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.13.1",
					},
				},
				Config: sloFromSDKCreateConfig,
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "slo_id", "id-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "budgeting_method", "occurrences"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.#", "1"),
				),
			},
			{
				// Verify the current (Framework) implementation can read and manage the SDK-created state.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "slo_id", "id-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "budgeting_method", "occurrences"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceSloRangeFromZero(t *testing.T) {
	constraints, err := version.NewConstraint(">=8.12.0")
	require.NoError(t, err)

	versionutils.SkipIfUnsupportedConstraints(t, constraints, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "name", "[Crossplane] Managed Resource External API Request Duration "+suffix),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_slo.xp_upjet_ext_api_duration",
						"description",
						"Tests that the SLO can be created with a range from 0.",
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "slo_id", "id-"+suffix),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "budgeting_method", "occurrences"),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.index", "metrics-*:metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.filter", "prometheus.upjet_resource_ext_api_duration.histogram: *"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.timestamp_field", "@timestamp"),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.good.0.field", "prometheus.upjet_resource_ext_api_duration.histogram"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.good.0.aggregation", "range"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.good.0.from", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.good.0.to", "10"),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.total.0.field", "prometheus.upjet_resource_ext_api_duration.histogram"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.total.0.aggregation", "range"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.total.0.from", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "histogram_custom_indicator.0.total.0.to", "999999"),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "time_window.0.duration", "30d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "time_window.0.type", "rolling"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "objective.0.target", "0.99"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "group_by.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "group_by.0", "orchestrator.cluster.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "tags.0", "crossplane"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.xp_upjet_ext_api_duration", "tags.1", "infra-mki"),
				),
			},
		},
	})
}

// TestAccResourceSloFloatPrecision verifies that objective fields (target,
// timeslice_target) round-trip through the provider without precision loss.
// Prior to the fix in https://github.com/elastic/terraform-provider-elasticstack/issues/2396,
// the generated client used float32 for these fields. Values like 0.999 are not
// exactly representable in float32, so reading them back produced different bits
// (e.g. float64(float32(0.999)) = 0.9990000128746033), causing a "provider
// produced inconsistent result after apply" error.
func TestAccResourceSloFloatPrecision(t *testing.T) {
	versionutils.SkipIfUnsupported(t, sloTimesliceMetricsMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.target", "0.999"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_target", "0.95"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_window", "5m"),
				),
			},
		},
	})
}

// TestAccResourceSloHistogramFloatPrecision verifies that histogram_custom_indicator
// range fields (from, to) round-trip without precision loss.
// See https://github.com/elastic/terraform-provider-elasticstack/issues/2400:
// float64(float32(0.001)) = 0.0010000000474974513, causing a "provider produced
// inconsistent result after apply" error when those fields were float32 in the client.
func TestAccResourceSloHistogramFloatPrecision(t *testing.T) {
	versionutils.SkipIfUnsupported(t, sloTimesliceMetricsMinVersion, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.from", "0.001"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.to", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.target", "0.999"),
				),
			},
		},
	})
}

func TestAccResourceSlo_long_slo_id(t *testing.T) {
	// 48-character slo_id is only supported server-side from 8.16.0 onwards.
	slo8_16Constraints, err := version.NewConstraint(">=8.16.0")
	require.NoError(t, err)
	versionutils.SkipIfUnsupportedConstraints(t, slo8_16Constraints, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	longSloID := "my-slo-id-that-is-exactly-48-characters-long-now"
	require.Len(t, longSloID, 48, "slo_id must be exactly 48 characters to test the boundary")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name":   config.StringVariable(sloName),
					"slo_id": config.StringVariable(longSloID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "slo_id", longSloID),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "description", "fully sick SLO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.service", "my-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_type", "request"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_name", "GET /sup/dawg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.threshold", "500"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.duration", "7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.type", "rolling"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "budgeting_method", "timeslices"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.target", "0.999"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_target", "0.95"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_window", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceSlo_36_char_slo_id(t *testing.T) {
	// 36-character slo_id is the historic server-side limit on versions before 8.16.0.
	slo36CharConstraints, err := version.NewConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4,<8.16.0")
	require.NoError(t, err)
	versionutils.SkipIfUnsupportedConstraints(t, slo36CharConstraints, versionutils.FlavorAny)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	sloID36 := "slo-id-that-is-exactly-36-characters"
	require.Len(t, sloID36, 36, "slo_id must be exactly 36 characters to test the boundary")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name":   config.StringVariable(sloName),
					"slo_id": config.StringVariable(sloID36),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "slo_id", sloID36),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "description", "fully sick SLO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.service", "my-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_type", "request"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_name", "GET /sup/dawg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.threshold", "500"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.duration", "7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.type", "rolling"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "budgeting_method", "timeslices"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.target", "0.999"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_target", "0.95"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_window", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
				),
			},
		},
	})
}

// checkSloAPIEnabled asserts the SLO get API reports the same enabled flag as Terraform state
// (after the provider's enable/disable reconciliation).
func checkSloAPIEnabled(want bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["elasticstack_kibana_slo.test_slo"]
		if !ok {
			return fmt.Errorf("resource elasticstack_kibana_slo.test_slo not found in state")
		}
		compID, diags := clients.CompositeIDFromStr(rs.Primary.ID)
		if diags.HasError() {
			return fmt.Errorf("parse composite id: %v", diags)
		}
		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		oapi, err := client.GetKibanaOapiClient()
		if err != nil {
			return err
		}
		apiSlo, getDiags := kibanaoapi.GetSlo(context.Background(), oapi, compID.ClusterID, compID.ResourceID)
		if getDiags.HasError() {
			return fmt.Errorf("get SLO: %v", getDiags)
		}
		if apiSlo == nil {
			return fmt.Errorf("SLO %q not found in Kibana for API enabled check", compID.ResourceID)
		}
		if apiSlo.Enabled != want {
			return fmt.Errorf("Kibana GetSLO enabled=%v, want %v (space=%q, sloId=%q)", apiSlo.Enabled, want, compID.ClusterID, compID.ResourceID)
		}
		return nil
	}
}

// checkResourceSloDestroy verifies all SLO resources have been destroyed.
// The composite ID stores spaceID as ClusterID and sloID as ResourceID.
var checkResourceSloDestroy = checks.KibanaResourceDestroyCheckCompositeID(
	"elasticstack_kibana_slo",
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, sloID string) (bool, error) {
		res, diags := kibanaoapi.GetSlo(ctx, client, spaceID, sloID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to check if SLO was destroyed: %v", diags)
		}
		return res != nil, nil
	},
)
