package kibana

import (
	"context"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var SLOSupportsMultipleGroupByMinVersion = version.Must(version.NewVersion("8.14.0"))

func ResourceSlo() *schema.Resource {
	return &schema.Resource{
		Description: "Creates an SLO.",

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

func getOrNilString(path string, d *schema.ResourceData) *string {
	if v, ok := d.GetOk(path); ok {
		str := v.(string)
		return &str
	}
	return nil
}

func getOrNilFloat(path string, d *schema.ResourceData) *float64 {
	if v, ok := d.GetOk(path); ok {
		f := v.(float64)
		return &f
	}
	return nil
}

func getSloFromResourceData(d *schema.ResourceData) (models.Slo, diag.Diagnostics) {
	var diags diag.Diagnostics

	var indicator kbapi.SLOsSloDefinitionResponse_Indicator
	var indicatorType string
	for key := range indicatorAddressToType {
		_, exists := d.GetOk(key)
		if exists {
			indicatorType = key
		}
	}

	switch indicatorType {
	case "kql_custom_indicator":
		var goodFilter kbapi.SLOsKqlWithFiltersGood
		goodFilter.FromSLOsKqlWithFiltersGood0(d.Get(indicatorType + ".0.good").(string))

		var totalFilter kbapi.SLOsKqlWithFiltersTotal
		totalFilter.FromSLOsKqlWithFiltersTotal0(d.Get(indicatorType + ".0.total").(string))

		customKql := kbapi.SLOsIndicatorPropertiesCustomKql{
			Type: indicatorAddressToType[indicatorType],
			Params: struct {
				DataViewId     *string                       `json:"dataViewId,omitempty"`
				Filter         *kbapi.SLOsKqlWithFilters     `json:"filter,omitempty"`
				Good           kbapi.SLOsKqlWithFiltersGood  `json:"good"`
				Index          string                        `json:"index"`
				TimestampField string                        `json:"timestampField"`
				Total          kbapi.SLOsKqlWithFiltersTotal `json:"total"`
			}{
				Index:          d.Get(indicatorType + ".0.index").(string),
				Good:           goodFilter,
				Total:          totalFilter,
				TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
			},
		}
		indicator.FromSLOsIndicatorPropertiesCustomKql(customKql)

	case "apm_availability_indicator":
		apmAvailability := kbapi.SLOsIndicatorPropertiesApmAvailability{
			Type: indicatorAddressToType[indicatorType],
			Params: struct {
				Environment     string  `json:"environment"`
				Filter          *string `json:"filter,omitempty"`
				Index           string  `json:"index"`
				Service         string  `json:"service"`
				TransactionName string  `json:"transactionName"`
				TransactionType string  `json:"transactionType"`
			}{
				Service:         d.Get(indicatorType + ".0.service").(string),
				Environment:     d.Get(indicatorType + ".0.environment").(string),
				TransactionType: d.Get(indicatorType + ".0.transaction_type").(string),
				TransactionName: d.Get(indicatorType + ".0.transaction_name").(string),
				Filter:          getOrNilString(indicatorType+".0.filter", d),
				Index:           d.Get(indicatorType + ".0.index").(string),
			},
		}
		indicator.FromSLOsIndicatorPropertiesApmAvailability(apmAvailability)

	case "apm_latency_indicator":
		apmLatency := kbapi.SLOsIndicatorPropertiesApmLatency{
			Type: indicatorAddressToType[indicatorType],
			Params: struct {
				Environment     string  `json:"environment"`
				Filter          *string `json:"filter,omitempty"`
				Index           string  `json:"index"`
				Service         string  `json:"service"`
				Threshold       float32 `json:"threshold"`
				TransactionName string  `json:"transactionName"`
				TransactionType string  `json:"transactionType"`
			}{
				Service:         d.Get(indicatorType + ".0.service").(string),
				Environment:     d.Get(indicatorType + ".0.environment").(string),
				TransactionType: d.Get(indicatorType + ".0.transaction_type").(string),
				TransactionName: d.Get(indicatorType + ".0.transaction_name").(string),
				Filter:          getOrNilString(indicatorType+".0.filter", d),
				Index:           d.Get(indicatorType + ".0.index").(string),
				Threshold:       float32(d.Get(indicatorType + ".0.threshold").(int)),
			},
		}
		indicator.FromSLOsIndicatorPropertiesApmLatency(apmLatency)

	case "histogram_custom_indicator":
		histogram := kbapi.SLOsIndicatorPropertiesHistogram{
			Type: indicatorAddressToType[indicatorType],
			Params: struct {
				DataViewId *string `json:"dataViewId,omitempty"`
				Filter     *string `json:"filter,omitempty"`
				Good       struct {
					Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsGoodAggregation `json:"aggregation"`
					Field       string                                                      `json:"field"`
					Filter      *string                                                     `json:"filter,omitempty"`
					From        *float32                                                    `json:"from,omitempty"`
					To          *float32                                                    `json:"to,omitempty"`
				} `json:"good"`
				Index          string `json:"index"`
				TimestampField string `json:"timestampField"`
				Total          struct {
					Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsTotalAggregation `json:"aggregation"`
					Field       string                                                       `json:"field"`
					Filter      *string                                                      `json:"filter,omitempty"`
					From        *float32                                                     `json:"from,omitempty"`
					To          *float32                                                     `json:"to,omitempty"`
				} `json:"total"`
			}{
				Filter:         getOrNilString(indicatorType+".0.filter", d),
				Index:          d.Get(indicatorType + ".0.index").(string),
				TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
				Good: struct {
					Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsGoodAggregation `json:"aggregation"`
					Field       string                                                      `json:"field"`
					Filter      *string                                                     `json:"filter,omitempty"`
					From        *float32                                                    `json:"from,omitempty"`
					To          *float32                                                    `json:"to,omitempty"`
				}{
					Field:       d.Get(indicatorType + ".0.good.0.field").(string),
					Aggregation: kbapi.SLOsIndicatorPropertiesHistogramParamsGoodAggregation(d.Get(indicatorType + ".0.good.0.aggregation").(string)),
					Filter:      getOrNilString(indicatorType+".0.good.0.filter", d),
					From:        convertFloat64PtrToFloat32Ptr(getOrNilFloat(indicatorType+".0.good.0.from", d)),
					To:          convertFloat64PtrToFloat32Ptr(getOrNilFloat(indicatorType+".0.good.0.to", d)),
				},
				Total: struct {
					Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsTotalAggregation `json:"aggregation"`
					Field       string                                                       `json:"field"`
					Filter      *string                                                      `json:"filter,omitempty"`
					From        *float32                                                     `json:"from,omitempty"`
					To          *float32                                                     `json:"to,omitempty"`
				}{
					Field:       d.Get(indicatorType + ".0.total.0.field").(string),
					Aggregation: kbapi.SLOsIndicatorPropertiesHistogramParamsTotalAggregation(d.Get(indicatorType + ".0.total.0.aggregation").(string)),
					Filter:      getOrNilString(indicatorType+".0.total.0.filter", d),
					From:        convertFloat64PtrToFloat32Ptr(getOrNilFloat(indicatorType+".0.total.0.from", d)),
					To:          convertFloat64PtrToFloat32Ptr(getOrNilFloat(indicatorType+".0.total.0.to", d)),
				},
			},
		}
		indicator.FromSLOsIndicatorPropertiesHistogram(histogram)

	case "metric_custom_indicator":
		params := d.Get("metric_custom_indicator.0").(map[string]interface{})

		// Parse good and total queries
		goodBlock := params["good"].([]interface{})[0].(map[string]interface{})
		totalBlock := params["total"].([]interface{})[0].(map[string]interface{})

		// Parse good metrics
		goodMetricsIface := goodBlock["metrics"].([]interface{})
		goodMetrics := make([]kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item, len(goodMetricsIface))
		for i, m := range goodMetricsIface {
			metric := m.(map[string]interface{})
			agg := metric["aggregation"].(string)

			var metricItem kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item
			switch agg {
			case "sum":
				metricWithField := kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0{
					Name:        metric["name"].(string),
					Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0AggregationSum,
					Field:       metric["field"].(string),
				}
				if filter, ok := metric["filter"].(string); ok && filter != "" {
					metricWithField.Filter = &filter
				}
				metricItem.FromSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0(metricWithField)
			case "doc_count":
				docCountMetric := kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1{
					Name:        metric["name"].(string),
					Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1AggregationDocCount,
				}
				if filter, ok := metric["filter"].(string); ok && filter != "" {
					docCountMetric.Filter = &filter
				}
				metricItem.FromSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1(docCountMetric)
			default:
				return models.Slo{}, diag.Errorf("good.metrics[%d]: unsupported aggregation '%s', only 'sum' and 'doc_count' are supported", i, agg)
			}
			goodMetrics[i] = metricItem
		}

		// Parse total metrics
		totalMetricsIface := totalBlock["metrics"].([]interface{})
		totalMetrics := make([]kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item, len(totalMetricsIface))
		for i, m := range totalMetricsIface {
			metric := m.(map[string]interface{})
			agg := metric["aggregation"].(string)

			var metricItem kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item
			switch agg {
			case "sum":
				metricWithField := kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0{
					Name:        metric["name"].(string),
					Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0AggregationSum,
					Field:       metric["field"].(string),
				}
				if filter, ok := metric["filter"].(string); ok && filter != "" {
					metricWithField.Filter = &filter
				}
				metricItem.FromSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0(metricWithField)
			case "doc_count":
				docCountMetric := kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1{
					Name:        metric["name"].(string),
					Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1AggregationDocCount,
				}
				if filter, ok := metric["filter"].(string); ok && filter != "" {
					docCountMetric.Filter = &filter
				}
				metricItem.FromSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1(docCountMetric)
			default:
				return models.Slo{}, diag.Errorf("total.metrics[%d]: unsupported aggregation '%s', only 'sum' and 'doc_count' are supported", i, agg)
			}
			totalMetrics[i] = metricItem
		}

		customMetric := kbapi.SLOsIndicatorPropertiesCustomMetric{
			Type: indicatorAddressToType[indicatorType],
			Params: struct {
				DataViewId *string `json:"dataViewId,omitempty"`
				Filter     *string `json:"filter,omitempty"`
				Good       struct {
					Equation string                                                               `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item `json:"metrics"`
				} `json:"good"`
				Index          string `json:"index"`
				TimestampField string `json:"timestampField"`
				Total          struct {
					Equation string                                                                `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item `json:"metrics"`
				} `json:"total"`
			}{
				Filter:         getOrNilString(indicatorType+".0.filter", d),
				Index:          d.Get(indicatorType + ".0.index").(string),
				TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
				Good: struct {
					Equation string                                                               `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item `json:"metrics"`
				}{
					Equation: goodBlock["equation"].(string),
					Metrics:  goodMetrics,
				},
				Total: struct {
					Equation string                                                                `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item `json:"metrics"`
				}{
					Equation: totalBlock["equation"].(string),
					Metrics:  totalMetrics,
				},
			},
		}
		indicator.FromSLOsIndicatorPropertiesCustomMetric(customMetric)

	case "timeslice_metric_indicator":
		params := d.Get("timeslice_metric_indicator.0").(map[string]interface{})
		metricBlock := params["metric"].([]interface{})[0].(map[string]interface{})
		metricsIface := metricBlock["metrics"].([]interface{})

		metrics := make([]kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item, len(metricsIface))
		for i, m := range metricsIface {
			metric := m.(map[string]interface{})
			agg := metric["aggregation"].(string)

			var metricItem kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item
			switch agg {
			case "sum", "avg", "min", "max", "value_count":
				basicMetric := kbapi.SLOsTimesliceMetricBasicMetricWithField{
					Name:        metric["name"].(string),
					Aggregation: kbapi.SLOsTimesliceMetricBasicMetricWithFieldAggregation(agg),
					Field:       metric["field"].(string),
				}
				if filter, ok := metric["filter"].(string); ok && filter != "" {
					basicMetric.Filter = &filter
				}
				metricItem.FromSLOsTimesliceMetricBasicMetricWithField(basicMetric)
			case "percentile":
				percentileMetric := kbapi.SLOsTimesliceMetricPercentileMetric{
					Name:        metric["name"].(string),
					Aggregation: kbapi.SLOsTimesliceMetricPercentileMetricAggregation(agg),
					Field:       metric["field"].(string),
					Percentile:  float32(metric["percentile"].(float64)),
				}
				if filter, ok := metric["filter"].(string); ok && filter != "" {
					percentileMetric.Filter = &filter
				}
				metricItem.FromSLOsTimesliceMetricPercentileMetric(percentileMetric)
			case "doc_count":
				docCountMetric := kbapi.SLOsTimesliceMetricDocCountMetric{
					Name:        metric["name"].(string),
					Aggregation: kbapi.SLOsTimesliceMetricDocCountMetricAggregation(agg),
				}
				if filter, ok := metric["filter"].(string); ok && filter != "" {
					docCountMetric.Filter = &filter
				}
				metricItem.FromSLOsTimesliceMetricDocCountMetric(docCountMetric)
			default:
				return models.Slo{}, diag.Errorf("metrics[%d]: unsupported aggregation '%s'", i, agg)
			}
			metrics[i] = metricItem
		}

		threshold := float32(metricBlock["threshold"].(float64))
		equation := metricBlock["equation"].(string)
		comparator := kbapi.SLOsIndicatorPropertiesTimesliceMetricParamsMetricComparator(metricBlock["comparator"].(string))

		timesliceMetric := kbapi.SLOsIndicatorPropertiesTimesliceMetric{
			Type: indicatorAddressToType[indicatorType],
			Params: struct {
				DataViewId *string `json:"dataViewId,omitempty"`
				Filter     *string `json:"filter,omitempty"`
				Index      string  `json:"index"`
				Metric     struct {
					Comparator kbapi.SLOsIndicatorPropertiesTimesliceMetricParamsMetricComparator        `json:"comparator"`
					Equation   string                                                                    `json:"equation"`
					Metrics    []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item `json:"metrics"`
					Threshold  float32                                                                   `json:"threshold"`
				} `json:"metric"`
				TimestampField string `json:"timestampField"`
			}{
				Filter:         getOrNilString(indicatorType+".0.filter", d),
				Index:          d.Get(indicatorType + ".0.index").(string),
				TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
				Metric: struct {
					Comparator kbapi.SLOsIndicatorPropertiesTimesliceMetricParamsMetricComparator        `json:"comparator"`
					Equation   string                                                                    `json:"equation"`
					Metrics    []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item `json:"metrics"`
					Threshold  float32                                                                   `json:"threshold"`
				}{
					Metrics:    metrics,
					Equation:   equation,
					Threshold:  threshold,
					Comparator: comparator,
				},
			},
		}
		indicator.FromSLOsIndicatorPropertiesTimesliceMetric(timesliceMetric)

	default:
		return models.Slo{}, diag.Errorf("unknown indicator type %s", indicatorType)
	}

	timeWindow := kbapi.SLOsTimeWindow{
		Type:     kbapi.SLOsTimeWindowType(d.Get("time_window.0.type").(string)),
		Duration: d.Get("time_window.0.duration").(string),
	}

	objective := kbapi.SLOsObjective{
		Target:          d.Get("objective.0.target").(float64),
		TimesliceTarget: getOrNilFloat("objective.0.timeslice_target", d),
		TimesliceWindow: getOrNilString("objective.0.timeslice_window", d),
	}

	settings := kbapi.SLOsSettings{
		SyncDelay: getOrNilString("settings.0.sync_delay", d),
		Frequency: getOrNilString("settings.0.frequency", d),
	}

	budgetingMethod := kbapi.SLOsBudgetingMethod(d.Get("budgeting_method").(string))

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
	if sloID := getOrNilString("slo_id", d); sloID != nil && *sloID != "" {
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

	supportsMultipleGroupBy := serverVersion.GreaterThanOrEqual(SLOSupportsMultipleGroupByMinVersion)
	if len(slo.GroupBy) > 1 && !supportsMultipleGroupBy {
		return diag.Errorf("multiple group_by fields are not supported in this version of the Elastic Stack. Multiple group_by fields requires %s", SLOSupportsMultipleGroupByMinVersion)
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	res, diags := kibana_oapi.CreateSlo(ctx, oapiClient, slo)
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

	supportsMultipleGroupBy := serverVersion.GreaterThanOrEqual(SLOSupportsMultipleGroupByMinVersion)
	if len(slo.GroupBy) > 1 && !supportsMultipleGroupBy {
		return diag.Errorf("multiple group_by fields are not supported in this version of the Elastic Stack. Multiple group_by fields requires %s", SLOSupportsMultipleGroupByMinVersion)
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	res, diags := kibana_oapi.UpdateSlo(ctx, oapiClient, slo)
	if diags.HasError() {
		return diags
	}

	compositeID := &clients.CompositeId{ClusterId: slo.SpaceID, ResourceId: res.SloID}
	d.SetId(compositeID.String())

	return resourceSloRead(ctx, d, meta)
}

func setIndicatorInResourceData(s *models.Slo, d *schema.ResourceData) diag.Diagnostics {
	indicator := []interface{}{}
	var indicatorAddress string

	value, err := s.Indicator.ValueByDiscriminator()
	if err != nil {
		return diag.Errorf("failed to get discriminator value: %v", err)
	}

	switch indicatorValue := value.(type) {
	case kbapi.SLOsIndicatorPropertiesApmAvailability:
		indicatorAddress = indicatorTypeToAddress[indicatorValue.Type]
		params := indicatorValue.Params
		indicator = append(indicator, map[string]interface{}{
			"environment":      params.Environment,
			"service":          params.Service,
			"transaction_type": params.TransactionType,
			"transaction_name": params.TransactionName,
			"index":            params.Index,
			"filter":           params.Filter,
		})

	case kbapi.SLOsIndicatorPropertiesApmLatency:
		indicatorAddress = indicatorTypeToAddress[indicatorValue.Type]
		params := indicatorValue.Params
		indicator = append(indicator, map[string]interface{}{
			"environment":      params.Environment,
			"service":          params.Service,
			"transaction_type": params.TransactionType,
			"transaction_name": params.TransactionName,
			"index":            params.Index,
			"filter":           params.Filter,
			"threshold":        params.Threshold,
		})

	case kbapi.SLOsIndicatorPropertiesCustomKql:
		indicatorAddress = indicatorTypeToAddress[indicatorValue.Type]
		params := indicatorValue.Params
		indicator = append(indicator, map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"good":            params.Good,
			"total":           params.Total,
			"timestamp_field": params.TimestampField,
		})

	case kbapi.SLOsIndicatorPropertiesHistogram:
		indicatorAddress = indicatorTypeToAddress[indicatorValue.Type]
		params := indicatorValue.Params
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
		indicator = append(indicator, map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"timestamp_field": params.TimestampField,
			"good":            good,
			"total":           total,
		})

	case kbapi.SLOsIndicatorPropertiesCustomMetric:
		indicatorAddress = indicatorTypeToAddress[indicatorValue.Type]
		params := indicatorValue.Params

		// Convert good metrics
		goodMetrics := []map[string]interface{}{}
		for _, item := range params.Good.Metrics {
			// Try to extract metric from union type
			if basic, err := item.AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0(); err == nil {
				goodMetrics = append(goodMetrics, map[string]interface{}{
					"name":        basic.Name,
					"aggregation": basic.Aggregation,
					"field":       basic.Field,
					"filter":      basic.Filter,
				})
			} else if docCount, err := item.AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1(); err == nil {
				goodMetrics = append(goodMetrics, map[string]interface{}{
					"name":        docCount.Name,
					"aggregation": docCount.Aggregation,
					"filter":      docCount.Filter,
				})
			}
		}
		good := []map[string]interface{}{{
			"equation": params.Good.Equation,
			"metrics":  goodMetrics,
		}}

		// Convert total metrics
		totalMetrics := []map[string]interface{}{}
		for _, item := range params.Total.Metrics {
			// Try to extract metric from union type
			if basic, err := item.AsSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0(); err == nil {
				totalMetrics = append(totalMetrics, map[string]interface{}{
					"name":        basic.Name,
					"aggregation": basic.Aggregation,
					"field":       basic.Field,
					"filter":      basic.Filter,
				})
			} else if docCount, err := item.AsSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1(); err == nil {
				totalMetrics = append(totalMetrics, map[string]interface{}{
					"name":        docCount.Name,
					"aggregation": docCount.Aggregation,
					"filter":      docCount.Filter,
				})
			}
		}
		total := []map[string]interface{}{{
			"equation": params.Total.Equation,
			"metrics":  totalMetrics,
		}}
		indicator = append(indicator, map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"timestamp_field": params.TimestampField,
			"good":            good,
			"total":           total,
		})

	case kbapi.SLOsIndicatorPropertiesTimesliceMetric:
		indicatorAddress = indicatorTypeToAddress[indicatorValue.Type]
		params := indicatorValue.Params

		// Convert metrics from union types
		metrics := []map[string]interface{}{}
		for _, item := range params.Metric.Metrics {
			metric := map[string]interface{}{}

			value, err := item.ValueByDiscriminator()
			if err != nil {
				continue // Skip invalid metrics
			}
			switch metricValue := value.(type) {
			case kbapi.SLOsTimesliceMetricBasicMetricWithField:
				metric["name"] = metricValue.Name
				metric["aggregation"] = metricValue.Aggregation
				metric["field"] = metricValue.Field
				if metricValue.Filter != nil {
					metric["filter"] = *metricValue.Filter
				}
			case kbapi.SLOsTimesliceMetricPercentileMetric:
				metric["name"] = metricValue.Name
				metric["aggregation"] = metricValue.Aggregation
				metric["field"] = metricValue.Field
				metric["percentile"] = metricValue.Percentile
				if metricValue.Filter != nil {
					metric["filter"] = *metricValue.Filter
				}
			case kbapi.SLOsTimesliceMetricDocCountMetric:
				metric["name"] = metricValue.Name
				metric["aggregation"] = metricValue.Aggregation
				if metricValue.Filter != nil {
					metric["filter"] = *metricValue.Filter
				}
			}
			metrics = append(metrics, metric)
		}
		metricBlock := map[string]interface{}{
			"metrics":    metrics,
			"equation":   params.Metric.Equation,
			"comparator": params.Metric.Comparator,
			"threshold":  params.Metric.Threshold,
		}
		indicator = append(indicator, map[string]interface{}{
			"index":           params.Index,
			"timestamp_field": params.TimestampField,
			"filter":          params.Filter,
			"metric":          []interface{}{metricBlock},
		})

	default:
		return diag.Errorf("unsupported indicator type: %T", indicatorValue)
	}

	if err := d.Set(indicatorAddress, indicator); err != nil {
		return diag.FromErr(err)
	}

	return nil
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

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	s, diags := kibana_oapi.GetSlo(ctx, oapiClient, spaceId, id)
	if s == nil && diags == nil {
		d.SetId("")
		return nil
	}
	if diags.HasError() {
		return diags
	}

	// Set indicator data in resource
	if diags := setIndicatorInResourceData(s, d); diags.HasError() {
		return diags
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
			"sync_delay": s.Settings.SyncDelay,
			"frequency":  s.Settings.Frequency,
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

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	spaceId := d.Get("space_id").(string)

	if diags = kibana_oapi.DeleteSlo(ctx, oapiClient, spaceId, compId.ResourceId); diags.HasError() {
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

func convertFloat64PtrToFloat32Ptr(f64ptr *float64) *float32 {
	if f64ptr == nil {
		return nil
	}
	f32 := float32(*f64ptr)
	return &f32
}
