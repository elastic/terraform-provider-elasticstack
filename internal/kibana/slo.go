package kibana

import (
	"context"
	"fmt"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	SLOSupportsMultipleGroupByMinVersion        = version.Must(version.NewVersion("8.14.0"))
	SLOSupportsPreventInitialBackfillMinVersion = version.Must(version.NewVersion("8.15.0"))
	SLOSupportsDataViewIDMinVersion             = version.Must(version.NewVersion("8.15.0"))
)

func ResourceSlo() *schema.Resource {
	return &schema.Resource{
		Description: `Creates or updates a Kibana SLO. See the [Kibana SLO docs](https://www.elastic.co/guide/en/observability/current/slo.html) and [dev docs](https://github.com/elastic/kibana/blob/main/x-pack/plugins/observability/dev_docs/slo.md) for more information.`,

		CreateContext: resourceSloCreate,
		UpdateContext: resourceSloUpdate,
		ReadContext:   resourceSloRead,
		DeleteContext: resourceSloDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        getSchema(),
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    getResourceSchemaV0().CoreConfigSchema().ImpliedType(),
				Upgrade: func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
					groupBy, ok := rawState["group_by"]
					if !ok {
						return rawState, nil
					}

					groupByStr, ok := groupBy.(string)
					if !ok {
						return rawState, nil
					}

					if len(groupByStr) == 0 {
						return rawState, nil
					}

					rawState["group_by"] = []string{groupByStr}
					return rawState, nil
				},
			},
		},
	}
}

func getResourceSchemaV0() *schema.Resource {
	s := getSchema()
	s["group_by"] = &schema.Schema{
		Description: "Optional group by field to use to generate an SLO per distinct value.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    false,
	}

	return &schema.Resource{
		Schema: s,
	}
}

func getSchema() map[string]*schema.Schema {
	var indicatorAddresses []string
	for i := range indicatorAddressToType {
		indicatorAddresses = append(indicatorAddresses, i)
	}

	return map[string]*schema.Schema{
		"slo_id": {
			Description: "An ID (8 to 48 characters) that contains only letters, numbers, hyphens, and underscores. If omitted, a UUIDv1 will be generated server-side.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(8, 48),
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must contain only letters, numbers, hyphens, and underscores"),
			),
		},
		"name": {
			Description: "The name of the SLO.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "A description for the SLO.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"metric_custom_indicator": {
			Type:         schema.TypeList,
			MinItems:     1,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: indicatorAddresses,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Type:     schema.TypeString,
						Required: true,
					},
					"data_view_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Optional data view id to use for this indicator.",
					},
					"filter": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"timestamp_field": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "@timestamp",
					},
					"good": {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"metrics": {
									Type:     schema.TypeList,
									Required: true,
									MinItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Type:     schema.TypeString,
												Required: true,
											},
											"aggregation": {
												Type:     schema.TypeString,
												Required: true,
											},
											"field": {
												Type:     schema.TypeString,
												Required: true,
											},
											"filter": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
								},
								"equation": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"total": {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"metrics": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Type:     schema.TypeString,
												Required: true,
											},
											"aggregation": {
												Type:     schema.TypeString,
												Required: true,
											},
											"field": {
												Type:     schema.TypeString,
												Required: true,
											},
											"filter": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
								},
								"equation": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"histogram_custom_indicator": {
			Type:         schema.TypeList,
			MinItems:     1,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: indicatorAddresses,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Type:     schema.TypeString,
						Required: true,
					},
					"data_view_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Optional data view id to use for this indicator.",
					},
					"filter": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"timestamp_field": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "@timestamp",
					},
					"good": {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field": {
									Type:     schema.TypeString,
									Required: true,
								},
								"aggregation": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice([]string{"value_count", "range"}, false),
								},
								"filter": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"from": {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								"to": {
									Type:     schema.TypeFloat,
									Optional: true,
								},
							},
						},
					},
					"total": {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"aggregation": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice([]string{"value_count", "range"}, false),
								},
								"field": {
									Type:     schema.TypeString,
									Required: true,
								},
								"filter": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"from": {
									Type:     schema.TypeFloat,
									Optional: true,
								},
								"to": {
									Type:     schema.TypeFloat,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},
		"apm_latency_indicator": {
			Type:         schema.TypeList,
			MinItems:     1,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: indicatorAddresses,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Type:     schema.TypeString,
						Required: true,
					},
					"filter": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"service": {
						Type:     schema.TypeString,
						Required: true,
					},
					"environment": {
						Type:     schema.TypeString,
						Required: true,
					},
					"transaction_type": {
						Type:     schema.TypeString,
						Required: true,
					},
					"transaction_name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"threshold": {
						Type:     schema.TypeInt,
						Required: true,
					},
				},
			},
		},
		"apm_availability_indicator": {
			Type:         schema.TypeList,
			MinItems:     1,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: indicatorAddresses,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Type:     schema.TypeString,
						Required: true,
					},
					"filter": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"service": {
						Type:     schema.TypeString,
						Required: true,
					},
					"environment": {
						Type:     schema.TypeString,
						Required: true,
					},
					"transaction_type": {
						Type:     schema.TypeString,
						Required: true,
					},
					"transaction_name": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
		"kql_custom_indicator": {
			Type:         schema.TypeList,
			MinItems:     1,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: indicatorAddresses,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Type:     schema.TypeString,
						Required: true,
					},
					"data_view_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Optional data view id to use for this indicator.",
					},
					"filter": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"good": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"total": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"timestamp_field": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "@timestamp",
					},
				},
			},
		},
		"timeslice_metric_indicator": {
			Description:  "Defines a timeslice metric indicator for SLO.",
			Type:         schema.TypeList,
			MinItems:     1,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: indicatorAddresses,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Type:     schema.TypeString,
						Required: true,
					},
					"data_view_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Optional data view id to use for this indicator.",
					},
					"timestamp_field": {
						Type:     schema.TypeString,
						Required: true,
					},
					"filter": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"metric": {
						Type:     schema.TypeList,
						Required: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"metrics": {
									Type:     schema.TypeList,
									Required: true,
									MinItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "The unique name for this metric. Used as a variable in the equation field.",
											},
											"aggregation": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "The aggregation type for this metric. One of: sum, avg, min, max, value_count, percentile, doc_count. Determines which other fields are required:",
											},
											"field": {
												Type:        schema.TypeString,
												Optional:    true,
												Description: "Field to aggregate. Required for aggregations: sum, avg, min, max, value_count, percentile. Must NOT be set for doc_count.",
											},
											"percentile": {
												Type:        schema.TypeFloat,
												Optional:    true,
												Description: "Percentile value (e.g., 99). Required if aggregation is 'percentile'. Must NOT be set for other aggregations.",
											},
											"filter": {
												Type:        schema.TypeString,
												Optional:    true,
												Description: "Optional KQL filter for this metric. Supported for all aggregations except doc_count.",
											},
										},
									},
								},
								"equation": {
									Type:     schema.TypeString,
									Required: true,
								},
								"comparator": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice([]string{"GT", "GTE", "LT", "LTE"}, false),
								},
								"threshold": {
									Type:     schema.TypeFloat,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"time_window": {
			Description: "Currently support `calendarAligned` and `rolling` time windows. Any duration greater than 1 day can be used: days, weeks, months, quarters, years. Rolling time window requires a duration, e.g. `1w` for one week, and type: `rolling`. SLOs defined with such time window, will only consider the SLI data from the last duration period as a moving window. Calendar aligned time window requires a duration, limited to `1M` for monthly or `1w` for weekly, and type: `calendarAligned`.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"duration": {
						Type:     schema.TypeString,
						Required: true,
					},
					"type": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
		"budgeting_method": {
			Description:  "An `occurrences` budgeting method uses the number of good and total events during the time window. A `timeslices` budgeting method uses the number of good slices and total slices during the time window. A slice is an arbitrary time window (smaller than the overall SLO time window) that is either considered good or bad, calculated from the timeslice threshold and the ratio of good over total events that happened during the slice window. A budgeting method is required and must be either occurrences or timeslices.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"occurrences", "timeslices"}, false),
		},
		"objective": {
			Description: "The target objective is the value the SLO needs to meet during the time window. If a timeslices budgeting method is used, we also need to define the timesliceTarget which can be different than the overall SLO target.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"target": {
						Type:     schema.TypeFloat,
						Required: true,
					},
					"timeslice_target": {
						Type:     schema.TypeFloat,
						Optional: true,
					},
					"timeslice_window": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"settings": {
			Description: "The default settings should be sufficient for most users, but if needed, these properties can be overwritten.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"sync_delay": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"frequency": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"prevent_initial_backfill": {
						Description: "Prevents the underlying ES transform from attempting to backfill data on start, which can sometimes be resource-intensive or time-consuming and unnecessary",
						Type:        schema.TypeBool,
						Optional:    true,
					},
				},
			},
		},
		"space_id": {
			Description: "An identifier for the space. If space_id is not provided, the default space is used.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
			ForceNew:    true,
		},
		"group_by": {
			Description: "Optional group by fields to use to generate an SLO per distinct value.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    false,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			DefaultFunc: func() (interface{}, error) {
				return []string{"*"}, nil
			},
		},
		"tags": {
			Description: "The tags for the SLO.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    false,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func getOrNil[T any](path string, d *schema.ResourceData) *T {
	return transformOrNil[T](path, d, func(v interface{}) T {
		return v.(T)
	})
}

func transformOrNil[T any](path string, d *schema.ResourceData, transform func(interface{}) T) *T {
	if v, ok := d.GetOk(path); ok {
		val := transform(v)
		return &val
	}
	return nil
}

func getSloFromResourceData(d *schema.ResourceData) (models.Slo, diag.Diagnostics) {
	var diags diag.Diagnostics

	var indicator slo.SloWithSummaryResponseIndicator
	var indicatorType string
	for key := range indicatorAddressToType {
		_, exists := d.GetOk(key)
		if exists {
			indicatorType = key
		}
	}

	switch indicatorType {
	case "kql_custom_indicator":
		indicator = slo.SloWithSummaryResponseIndicator{
			IndicatorPropertiesCustomKql: &slo.IndicatorPropertiesCustomKql{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesCustomKqlParams{
					Index:      d.Get(indicatorType + ".0.index").(string),
					DataViewId: getOrNil[string](indicatorType+".0.data_view_id", d),
					Filter: transformOrNil[slo.KqlWithFilters](
						indicatorType+".0.filter", d,
						func(v interface{}) slo.KqlWithFilters {
							return slo.KqlWithFilters{
								String: utils.Pointer(v.(string)),
							}
						}),
					Good: slo.KqlWithFiltersGood{
						String: utils.Pointer(d.Get(indicatorType + ".0.good").(string)),
					},
					Total: slo.KqlWithFiltersTotal{
						String: utils.Pointer(d.Get(indicatorType + ".0.total").(string)),
					},
					TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
				},
			},
		}

	case "apm_availability_indicator":
		indicator = slo.SloWithSummaryResponseIndicator{
			IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesApmAvailabilityParams{
					Service:         d.Get(indicatorType + ".0.service").(string),
					Environment:     d.Get(indicatorType + ".0.environment").(string),
					TransactionType: d.Get(indicatorType + ".0.transaction_type").(string),
					TransactionName: d.Get(indicatorType + ".0.transaction_name").(string),
					Filter:          getOrNil[string](indicatorType+".0.filter", d),
					Index:           d.Get(indicatorType + ".0.index").(string),
				},
			},
		}

	case "apm_latency_indicator":
		indicator = slo.SloWithSummaryResponseIndicator{
			IndicatorPropertiesApmLatency: &slo.IndicatorPropertiesApmLatency{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesApmLatencyParams{
					Service:         d.Get(indicatorType + ".0.service").(string),
					Environment:     d.Get(indicatorType + ".0.environment").(string),
					TransactionType: d.Get(indicatorType + ".0.transaction_type").(string),
					TransactionName: d.Get(indicatorType + ".0.transaction_name").(string),
					Filter:          getOrNil[string](indicatorType+".0.filter", d),
					Index:           d.Get(indicatorType + ".0.index").(string),
					Threshold:       float64(d.Get(indicatorType + ".0.threshold").(int)),
				},
			},
		}

	case "histogram_custom_indicator":
		indicator = slo.SloWithSummaryResponseIndicator{
			IndicatorPropertiesHistogram: &slo.IndicatorPropertiesHistogram{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesHistogramParams{
					Filter:         getOrNil[string](indicatorType+".0.filter", d),
					Index:          d.Get(indicatorType + ".0.index").(string),
					DataViewId:     getOrNil[string](indicatorType+".0.data_view_id", d),
					TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
					Good: slo.IndicatorPropertiesHistogramParamsGood{
						Field:       d.Get(indicatorType + ".0.good.0.field").(string),
						Aggregation: d.Get(indicatorType + ".0.good.0.aggregation").(string),
						Filter:      getOrNil[string](indicatorType+".0.good.0.filter", d),
						From:        getOrNil[float64](indicatorType+".0.good.0.from", d),
						To:          getOrNil[float64](indicatorType+".0.good.0.to", d),
					},
					Total: slo.IndicatorPropertiesHistogramParamsTotal{
						Field:       d.Get(indicatorType + ".0.total.0.field").(string),
						Aggregation: d.Get(indicatorType + ".0.total.0.aggregation").(string),
						Filter:      getOrNil[string](indicatorType+".0.total.0.filter", d),
						From:        getOrNil[float64](indicatorType+".0.total.0.from", d),
						To:          getOrNil[float64](indicatorType+".0.total.0.to", d),
					},
				},
			},
		}

	case "metric_custom_indicator":
		goodMetricsRaw := d.Get(indicatorType + ".0.good.0.metrics").([]interface{})
		var goodMetrics []slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner
		for n := range goodMetricsRaw {
			idx := fmt.Sprint(n)
			goodMetrics = append(goodMetrics, slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{
				Name:        d.Get(indicatorType + ".0.good.0.metrics." + idx + ".name").(string),
				Field:       d.Get(indicatorType + ".0.good.0.metrics." + idx + ".field").(string),
				Aggregation: d.Get(indicatorType + ".0.good.0.metrics." + idx + ".aggregation").(string),
				Filter:      getOrNil[string](indicatorType+".0.good.0.metrics."+idx+".filter", d),
			})
		}
		totalMetricsRaw := d.Get(indicatorType + ".0.total.0.metrics").([]interface{})
		var totalMetrics []slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner
		for n := range totalMetricsRaw {
			idx := fmt.Sprint(n)
			totalMetrics = append(totalMetrics, slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{
				Name:        d.Get(indicatorType + ".0.total.0.metrics." + idx + ".name").(string),
				Field:       d.Get(indicatorType + ".0.total.0.metrics." + idx + ".field").(string),
				Aggregation: d.Get(indicatorType + ".0.total.0.metrics." + idx + ".aggregation").(string),
				Filter:      getOrNil[string](indicatorType+".0.total.0.metrics."+idx+".filter", d),
			})
		}
		indicator = slo.SloWithSummaryResponseIndicator{
			IndicatorPropertiesCustomMetric: &slo.IndicatorPropertiesCustomMetric{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesCustomMetricParams{
					Filter:         getOrNil[string](indicatorType+".0.filter", d),
					Index:          d.Get(indicatorType + ".0.index").(string),
					DataViewId:     getOrNil[string](indicatorType+".0.data_view_id", d),
					TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
					Good: slo.IndicatorPropertiesCustomMetricParamsGood{
						Equation: d.Get(indicatorType + ".0.good.0.equation").(string),
						Metrics:  goodMetrics,
					},
					Total: slo.IndicatorPropertiesCustomMetricParamsTotal{
						Equation: d.Get(indicatorType + ".0.total.0.equation").(string),
						Metrics:  totalMetrics,
					},
				},
			},
		}

	case "timeslice_metric_indicator":
		params := d.Get("timeslice_metric_indicator.0").(map[string]interface{})
		metricBlock := params["metric"].([]interface{})[0].(map[string]interface{})
		metricsIface := metricBlock["metrics"].([]interface{})
		metrics := make([]slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner, len(metricsIface))
		for i, m := range metricsIface {
			metric := m.(map[string]interface{})
			agg := metric["aggregation"].(string)
			switch agg {
			case "sum", "avg", "min", "max", "value_count":
				metrics[i] = slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
					TimesliceMetricBasicMetricWithField: &slo.TimesliceMetricBasicMetricWithField{
						Name:        metric["name"].(string),
						Aggregation: agg,
						Field:       metric["field"].(string),
					},
				}
			case "percentile":
				metrics[i] = slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
					TimesliceMetricPercentileMetric: &slo.TimesliceMetricPercentileMetric{
						Name:        metric["name"].(string),
						Aggregation: agg,
						Field:       metric["field"].(string),
						Percentile:  metric["percentile"].(float64),
					},
				}
			case "doc_count":
				metrics[i] = slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
					TimesliceMetricDocCountMetric: &slo.TimesliceMetricDocCountMetric{
						Name:        metric["name"].(string),
						Aggregation: agg,
					},
				}
			default:
				return models.Slo{}, diag.Errorf("metrics[%d]: unsupported aggregation '%s'", i, agg)
			}
		}
		indicator = slo.SloWithSummaryResponseIndicator{
			IndicatorPropertiesTimesliceMetric: &slo.IndicatorPropertiesTimesliceMetric{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesTimesliceMetricParams{
					Index:          params["index"].(string),
					DataViewId:     getOrNil[string]("timeslice_metric_indicator.0.data_view_id", d),
					TimestampField: params["timestamp_field"].(string),
					Filter:         getOrNil[string]("timeslice_metric_indicator.0.filter", d),
					Metric: slo.IndicatorPropertiesTimesliceMetricParamsMetric{
						Metrics:    metrics,
						Equation:   metricBlock["equation"].(string),
						Comparator: metricBlock["comparator"].(string),
						Threshold:  metricBlock["threshold"].(float64),
					},
				},
			},
		}

	default:
		return models.Slo{}, diag.Errorf("unknown indicator type %s", indicatorType)
	}

	timeWindow := slo.TimeWindow{
		Type:     d.Get("time_window.0.type").(string),
		Duration: d.Get("time_window.0.duration").(string),
	}

	objective := slo.Objective{
		Target:          d.Get("objective.0.target").(float64),
		TimesliceTarget: getOrNil[float64]("objective.0.timeslice_target", d),
		TimesliceWindow: getOrNil[string]("objective.0.timeslice_window", d),
	}

	settings := slo.Settings{
		SyncDelay:              getOrNil[string]("settings.0.sync_delay", d),
		Frequency:              getOrNil[string]("settings.0.frequency", d),
		PreventInitialBackfill: getOrNil[bool]("settings.0.prevent_initial_backfill", d),
	}

	budgetingMethod := slo.BudgetingMethod(d.Get("budgeting_method").(string))

	slo := models.Slo{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Indicator:       indicator,
		TimeWindow:      timeWindow,
		BudgetingMethod: budgetingMethod,
		Objective:       objective,
		Settings:        &settings,
		SpaceID:         d.Get("space_id").(string),
	}

	// Explicitly set SLO object id if provided, otherwise we'll use the autogenerated ID from the Kibana API response
	if sloID := getOrNil[string]("slo_id", d); sloID != nil && *sloID != "" {
		slo.SloID = *sloID
	}

	if groupBy, ok := d.GetOk("group_by"); ok {
		for _, g := range groupBy.([]interface{}) {
			slo.GroupBy = append(slo.GroupBy, g.(string))
		}
	}

	if tags, ok := d.GetOk("tags"); ok {
		for _, t := range tags.([]interface{}) {
			slo.Tags = append(slo.Tags, t.(string))
		}
	}

	return slo, diags
}

func resourceSloCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	slo, diags := getSloFromResourceData(d)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	// Version check for prevent_initial_backfill
	if slo.Settings.PreventInitialBackfill != nil {
		if !serverVersion.GreaterThanOrEqual(SLOSupportsPreventInitialBackfillMinVersion) {
			return diag.Errorf("The 'prevent_initial_backfill' setting requires Elastic Stack version %s or higher.", SLOSupportsPreventInitialBackfillMinVersion)
		}
	}

	// Version check for data_view_id support
	if !serverVersion.GreaterThanOrEqual(SLOSupportsDataViewIDMinVersion) {
		// Check all indicator types that support data_view_id
		for _, indicatorType := range []string{"metric_custom_indicator", "histogram_custom_indicator", "kql_custom_indicator", "timeslice_metric_indicator"} {
			if v, ok := d.GetOk(indicatorType + ".0.data_view_id"); ok && v != "" {
				return diag.Errorf("data_view_id is not supported for %s on Elastic Stack versions < %s", indicatorType, SLOSupportsDataViewIDMinVersion)
			}
		}
	}

	supportsMultipleGroupBy := serverVersion.GreaterThanOrEqual(SLOSupportsMultipleGroupByMinVersion)
	if len(slo.GroupBy) > 1 && !supportsMultipleGroupBy {
		return diag.Errorf("multiple group_by fields are not supported in this version of the Elastic Stack. Multiple group_by fields requires %s", SLOSupportsMultipleGroupByMinVersion)
	}

	res, diags := kibana.CreateSlo(ctx, client, slo, supportsMultipleGroupBy)
	if diags.HasError() {
		return diags
	}

	compositeID := &clients.CompositeId{ClusterId: slo.SpaceID, ResourceId: res.SloID}
	d.SetId(compositeID.String())

	return resourceSloRead(ctx, d, meta)
}

func resourceSloUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	slo, diags := getSloFromResourceData(d)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	// Version check for prevent_initial_backfill
	if slo.Settings.PreventInitialBackfill != nil {
		if !serverVersion.GreaterThanOrEqual(SLOSupportsPreventInitialBackfillMinVersion) {
			return diag.Errorf("The 'prevent_initial_backfill' setting requires Elastic Stack version %s or higher.", SLOSupportsPreventInitialBackfillMinVersion)
		}
	}

	// Version check for data_view_id support
	if !serverVersion.GreaterThanOrEqual(SLOSupportsDataViewIDMinVersion) {
		for _, indicatorType := range []string{"metric_custom_indicator", "histogram_custom_indicator", "kql_custom_indicator", "timeslice_metric_indicator"} {
			if v, ok := d.GetOk(indicatorType + ".0.data_view_id"); ok && v != "" {
				return diag.Errorf("data_view_id is not supported for %s on Elastic Stack versions < %s", indicatorType, SLOSupportsDataViewIDMinVersion)
			}
		}
	}

	supportsMultipleGroupBy := serverVersion.GreaterThanOrEqual(SLOSupportsMultipleGroupByMinVersion)
	if len(slo.GroupBy) > 1 && !supportsMultipleGroupBy {
		return diag.Errorf("multiple group_by fields are not supported in this version of the Elastic Stack. Multiple group_by fields requires %s", SLOSupportsMultipleGroupByMinVersion)
	}

	res, diags := kibana.UpdateSlo(ctx, client, slo, supportsMultipleGroupBy)
	if diags.HasError() {
		return diags
	}

	compositeID := &clients.CompositeId{ClusterId: slo.SpaceID, ResourceId: res.SloID}
	d.SetId(compositeID.String())

	return resourceSloRead(ctx, d, meta)
}

func resourceSloRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	id := compId.ResourceId
	spaceId := compId.ClusterId

	s, diags := kibana.GetSlo(ctx, client, id, spaceId)
	if s == nil && diags == nil {
		d.SetId("")
		return nil
	}
	if diags.HasError() {
		return diags
	}

	indicator := []interface{}{}
	var indicatorAddress string
	switch {
	case s.Indicator.IndicatorPropertiesApmAvailability != nil:
		indicatorAddress = indicatorTypeToAddress[s.Indicator.IndicatorPropertiesApmAvailability.Type]
		params := s.Indicator.IndicatorPropertiesApmAvailability.Params
		indicator = append(indicator, map[string]interface{}{
			"environment":      params.Environment,
			"service":          params.Service,
			"transaction_type": params.TransactionType,
			"transaction_name": params.TransactionName,
			"index":            params.Index,
			"filter":           params.Filter,
		})

	case s.Indicator.IndicatorPropertiesApmLatency != nil:
		indicatorAddress = indicatorTypeToAddress[s.Indicator.IndicatorPropertiesApmLatency.Type]
		params := s.Indicator.IndicatorPropertiesApmLatency.Params
		indicator = append(indicator, map[string]interface{}{
			"environment":      params.Environment,
			"service":          params.Service,
			"transaction_type": params.TransactionType,
			"transaction_name": params.TransactionName,
			"index":            params.Index,
			"filter":           params.Filter,
			"threshold":        params.Threshold,
		})

	case s.Indicator.IndicatorPropertiesCustomKql != nil:
		indicatorAddress = indicatorTypeToAddress[s.Indicator.IndicatorPropertiesCustomKql.Type]
		params := s.Indicator.IndicatorPropertiesCustomKql.Params
		indicatorMap := map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter.String,
			"good":            params.Good.String,
			"total":           params.Total.String,
			"timestamp_field": params.TimestampField,
		}
		if params.DataViewId != nil {
			indicatorMap["data_view_id"] = *params.DataViewId
		}
		indicator = append(indicator, indicatorMap)

	case s.Indicator.IndicatorPropertiesHistogram != nil:
		indicatorAddress = indicatorTypeToAddress[s.Indicator.IndicatorPropertiesHistogram.Type]
		params := s.Indicator.IndicatorPropertiesHistogram.Params
		good := []map[string]interface{}{{
			"field":       params.Good.Field,
			"aggregation": params.Good.Aggregation,
			"filter":      params.Good.Filter,
			"from":        params.Good.From,
			"to":          params.Good.To,
		}}
		total := []map[string]interface{}{{
			"field":       params.Total.Field,
			"aggregation": params.Total.Aggregation,
			"filter":      params.Total.Filter,
			"from":        params.Total.From,
			"to":          params.Total.To,
		}}
		indicatorMap := map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"timestamp_field": params.TimestampField,
			"good":            good,
			"total":           total,
		}
		if params.DataViewId != nil {
			indicatorMap["data_view_id"] = *params.DataViewId
		}
		indicator = append(indicator, indicatorMap)

	case s.Indicator.IndicatorPropertiesCustomMetric != nil:
		indicatorAddress = indicatorTypeToAddress[s.Indicator.IndicatorPropertiesCustomMetric.Type]
		params := s.Indicator.IndicatorPropertiesCustomMetric.Params
		goodMetrics := []map[string]interface{}{}
		for _, m := range params.Good.Metrics {
			goodMetrics = append(goodMetrics, map[string]interface{}{
				"name":        m.Name,
				"aggregation": m.Aggregation,
				"field":       m.Field,
				"filter":      m.Filter,
			})
		}
		good := []map[string]interface{}{{
			"equation": params.Good.Equation,
			"metrics":  goodMetrics,
		}}
		totalMetrics := []map[string]interface{}{}
		for _, m := range params.Total.Metrics {
			totalMetrics = append(totalMetrics, map[string]interface{}{
				"name":        m.Name,
				"aggregation": m.Aggregation,
				"field":       m.Field,
				"filter":      m.Filter,
			})
		}
		total := []map[string]interface{}{{
			"equation": params.Total.Equation,
			"metrics":  totalMetrics,
		}}
		indicatorMap := map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"timestamp_field": params.TimestampField,
			"good":            good,
			"total":           total,
		}
		if params.DataViewId != nil {
			indicatorMap["data_view_id"] = *params.DataViewId
		}
		indicator = append(indicator, indicatorMap)

	case s.Indicator.IndicatorPropertiesTimesliceMetric != nil:
		indicatorAddress = indicatorTypeToAddress[s.Indicator.IndicatorPropertiesTimesliceMetric.Type]
		params := s.Indicator.IndicatorPropertiesTimesliceMetric.Params
		metrics := []map[string]interface{}{}
		for _, m := range params.Metric.Metrics {
			metric := map[string]interface{}{}
			if m.TimesliceMetricBasicMetricWithField != nil {
				metric["name"] = m.TimesliceMetricBasicMetricWithField.Name
				metric["aggregation"] = m.TimesliceMetricBasicMetricWithField.Aggregation
				metric["field"] = m.TimesliceMetricBasicMetricWithField.Field
			}
			if m.TimesliceMetricPercentileMetric != nil {
				metric["name"] = m.TimesliceMetricPercentileMetric.Name
				metric["aggregation"] = m.TimesliceMetricPercentileMetric.Aggregation
				metric["field"] = m.TimesliceMetricPercentileMetric.Field
				metric["percentile"] = m.TimesliceMetricPercentileMetric.Percentile
			}
			if m.TimesliceMetricDocCountMetric != nil {
				metric["name"] = m.TimesliceMetricDocCountMetric.Name
				metric["aggregation"] = m.TimesliceMetricDocCountMetric.Aggregation
			}
			metrics = append(metrics, metric)
		}
		metricBlock := map[string]interface{}{
			"metrics":    metrics,
			"equation":   params.Metric.Equation,
			"comparator": params.Metric.Comparator,
			"threshold":  params.Metric.Threshold,
		}
		indicatorMap := map[string]interface{}{
			"index":           params.Index,
			"timestamp_field": params.TimestampField,
			"filter":          params.Filter,
			"metric":          []interface{}{metricBlock},
		}
		if params.DataViewId != nil {
			indicatorMap["data_view_id"] = *params.DataViewId
		}
		indicator = append(indicator, indicatorMap)

	default:
		return diag.Errorf("indicator not set")
	}

	if err := d.Set(indicatorAddress, indicator); err != nil {
		return diag.FromErr(err)
	}

	time_window := []interface{}{
		map[string]interface{}{
			"duration": s.TimeWindow.Duration,
			"type":     s.TimeWindow.Type,
		},
	}
	if err := d.Set("time_window", time_window); err != nil {
		return diag.FromErr(err)
	}

	objective := []interface{}{
		map[string]interface{}{
			"target":           s.Objective.Target,
			"timeslice_target": s.Objective.TimesliceTarget,
			"timeslice_window": s.Objective.TimesliceWindow,
		},
	}
	if err := d.Set("objective", objective); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("settings", []interface{}{
		map[string]interface{}{
			"sync_delay":               s.Settings.SyncDelay,
			"frequency":                s.Settings.Frequency,
			"prevent_initial_backfill": s.Settings.PreventInitialBackfill,
		},
	}); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("group_by", s.GroupBy); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("slo_id", s.SloID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("space_id", s.SpaceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", s.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", s.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("budgeting_method", s.BudgetingMethod); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", s.Tags); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSloDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	spaceId := d.Get("space_id").(string)

	if diags = kibana.DeleteSlo(ctx, client, spaceId, compId.ResourceId); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}

// indicatorAddressToType is a mapping between the terraform resource address and the internal indicator type name used by the API
var indicatorAddressToType = map[string]string{
	"apm_latency_indicator":      "sli.apm.transactionDuration",
	"apm_availability_indicator": "sli.apm.transactionErrorRate",
	"kql_custom_indicator":       "sli.kql.custom",
	"metric_custom_indicator":    "sli.metric.custom",
	"histogram_custom_indicator": "sli.histogram.custom",
	"timeslice_metric_indicator": "sli.metric.timeslice",
}

var indicatorTypeToAddress = utils.FlipMap(indicatorAddressToType)
