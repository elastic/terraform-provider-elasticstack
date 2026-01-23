package kibana_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	kibanaresource "github.com/elastic/terraform-provider-elasticstack/internal/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var sloTimesliceMetricsMinVersion = version.Must(version.NewVersion("8.12.0"))

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
					return func(s *terraform.State) error {
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

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
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
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.sync_delay", "1m"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.frequency", "1m"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
						),
					},
					{
						//check that name can be updated
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("update_name"),
						ConfigVariables: config.Variables{
							"name": config.StringVariable(fmt.Sprintf("updated-%s", sloName)),
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", fmt.Sprintf("updated-%s", sloName)),
						),
					},
					{ //check that settings can be updated from api-computed defaults
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
						},
						ConfigDirectory: acctest.NamedTestCaseDirectory("update_settings"),
						ConfigVariables: config.Variables{
							"name": config.StringVariable(sloName),
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.sync_delay", "5m"),
							resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.frequency", "5m"),
						),
					},
					{
						ProtoV6ProviderFactories: acctest.Providers,
						SkipFunc: func() (bool, error) {
							if !testWithDataViewID {
								return versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints)()
							}

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
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

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
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

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
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
							resource.TestCheckResourceAttrSet("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.from"),
							resource.TestCheckResourceAttrSet("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.to"),
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

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
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

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
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

							return versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsDataViewIDMinVersion)()
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

//go:embed testdata/TestAccResourceSloGroupBy/single_element/test.tf
var singleElementConfig string

func TestAccResourceSloGroupBy(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsMultipleGroupByMinVersion),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsMultipleGroupByMinVersion),
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
		},
	})
}

func TestAccResourceSloPreventInitialBackfill(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(kibanaresource.SLOSupportsPreventInitialBackfillMinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index-"+sloName),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.prevent_initial_backfill", "true"),
				),
			},
		},
	})
}

func TestAccResourceSlo_timeslice_metric_indicator_basic(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(sloTimesliceMetricsMinVersion),
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
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(sloTimesliceMetricsMinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "timeslice_metric_indicator.0.filter", ""),
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
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(sloTimesliceMetricsMinVersion),
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
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(sloTimesliceMetricsMinVersion),
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

func TestAccResourceSloErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multiple_indicators"),
				ExpectError:              regexp.MustCompile("Invalid combination of arguments"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.10.0-SNAPSHOT"))),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("agg_fail"),
				ExpectError:              regexp.MustCompile(`expected histogram_custom_indicator.0.good.0.aggregation to be one of \["?value_count"? "?range"?\], got supdawg`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("budget_fail"),
				ExpectError:              regexp.MustCompile(`expected budgeting_method to be one of \["occurrences" "timeslices"\], got supdawg`),
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
				ExpectError: regexp.MustCompile(`expected length of slo_id to be in the range \(8 - 48\)`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("toolongid"),
				ConfigVariables: config.Variables{
					"name":   config.StringVariable("toolongid"),
					"slo_id": config.StringVariable("this-id-is-way-too-long-and-exceeds-the-48-character-limit-for-slo-ids"),
				},
				ExpectError: regexp.MustCompile(`expected length of slo_id to be in the range \(8 - 48\)`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalidchars"),
				ConfigVariables: config.Variables{
					"name":   config.StringVariable("invalidchars"),
					"slo_id": config.StringVariable("invalid@id$"),
				},
				ExpectError: regexp.MustCompile(`must contain only letters, numbers, hyphens, and underscores`),
			},
		},
	})
}

func TestAccResourceSlo_kql_custom_indicator_basic(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSloDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(sloTimesliceMetricsMinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("test"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sloName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.index", "my-index-"+sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.good", "latency < 300"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.total", "*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.timestamp_field", "custom_timestamp"),
				),
			},
		},
	})
}

func TestSloIdValidation(t *testing.T) {
	resource := kibanaresource.ResourceSlo()
	sloIdSchema := resource.Schema["slo_id"]

	// Test valid slo_id values
	validIds := []string{
		"valid_id",   // 8 chars with underscore
		"valid-id",   // 8 chars with hyphen
		"validId123", // 11 chars with mixed case and numbers
		"a1234567",   // exactly 8 chars
		"this-is-a-very-long-but-valid-slo-id-12345678", // exactly 48 chars
	}

	for _, id := range validIds {
		warnings, errors := sloIdSchema.ValidateFunc(id, "slo_id")
		if len(errors) > 0 {
			t.Errorf("Expected valid ID %q to pass validation, but got errors: %v", id, errors)
		}
		if len(warnings) > 0 {
			t.Errorf("Expected valid ID %q to have no warnings, but got: %v", id, warnings)
		}
	}

	// Test invalid slo_id values
	invalidTests := []struct {
		id          string
		expectedErr string
	}{
		{"short", "expected length of slo_id to be in the range (8 - 48)"},
		{"1234567", "expected length of slo_id to be in the range (8 - 48)"},                                                                 // 7 chars
		{"this-is-a-very-long-slo-id-that-exceeds-the-48-character-limit-for-sure", "expected length of slo_id to be in the range (8 - 48)"}, // > 48 chars
		{"invalid@id", "must contain only letters, numbers, hyphens, and underscores"},
		{"invalid$id", "must contain only letters, numbers, hyphens, and underscores"},
		{"invalid id", "must contain only letters, numbers, hyphens, and underscores"}, // space
		{"invalid.id", "must contain only letters, numbers, hyphens, and underscores"}, // period
	}

	for _, test := range invalidTests {
		_, errors := sloIdSchema.ValidateFunc(test.id, "slo_id")
		if len(errors) == 0 {
			t.Errorf("Expected invalid ID %q to fail validation", test.id)
		} else {
			found := false
			for _, err := range errors {
				if strings.Contains(err.Error(), test.expectedErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error for ID %q to contain %q, but got: %v", test.id, test.expectedErr, errors)
			}
		}
	}
}

func checkResourceSloDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_slo" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		slo, diags := kibana.GetSlo(context.Background(), client, compId.ResourceId, compId.ClusterId)
		if diags.HasError() {
			if len(diags) > 1 || diags[0].Summary != "404 Not Found" {
				return fmt.Errorf("Failed to check if SLO was destroyed: %v", diags)
			}
		}

		if slo != nil {
			return fmt.Errorf("SLO (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
