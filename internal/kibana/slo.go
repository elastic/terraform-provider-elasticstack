package kibana

import (
	"context"
	"fmt"

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
			Description: "An ID (8 and 36 characters). If omitted, a UUIDv1 will be generated server-side.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
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

	var indicator slo.SloResponseIndicator
	var indicatorType string
	for key := range indicatorAddressToType {
		_, exists := d.GetOk(key)
		if exists {
			indicatorType = key
		}
	}

	switch indicatorType {
	case "kql_custom_indicator":
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesCustomKql: &slo.IndicatorPropertiesCustomKql{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesCustomKqlParams{
					Index:          d.Get(indicatorType + ".0.index").(string),
					Filter:         getOrNilString(indicatorType+".0.filter", d),
					Good:           d.Get(indicatorType + ".0.good").(string),
					Total:          d.Get(indicatorType + ".0.total").(string),
					TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
				},
			},
		}

	case "apm_availability_indicator":
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesApmAvailabilityParams{
					Service:         d.Get(indicatorType + ".0.service").(string),
					Environment:     d.Get(indicatorType + ".0.environment").(string),
					TransactionType: d.Get(indicatorType + ".0.transaction_type").(string),
					TransactionName: d.Get(indicatorType + ".0.transaction_name").(string),
					Filter:          getOrNilString(indicatorType+".0.filter", d),
					Index:           d.Get(indicatorType + ".0.index").(string),
				},
			},
		}

	case "apm_latency_indicator":
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesApmLatency: &slo.IndicatorPropertiesApmLatency{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesApmLatencyParams{
					Service:         d.Get(indicatorType + ".0.service").(string),
					Environment:     d.Get(indicatorType + ".0.environment").(string),
					TransactionType: d.Get(indicatorType + ".0.transaction_type").(string),
					TransactionName: d.Get(indicatorType + ".0.transaction_name").(string),
					Filter:          getOrNilString(indicatorType+".0.filter", d),
					Index:           d.Get(indicatorType + ".0.index").(string),
					Threshold:       float64(d.Get(indicatorType + ".0.threshold").(int)),
				},
			},
		}

	case "histogram_custom_indicator":
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesHistogram: &slo.IndicatorPropertiesHistogram{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesHistogramParams{
					Filter:         getOrNilString(indicatorType+".0.filter", d),
					Index:          d.Get(indicatorType + ".0.index").(string),
					TimestampField: d.Get(indicatorType + ".0.timestamp_field").(string),
					Good: slo.IndicatorPropertiesHistogramParamsGood{
						Field:       d.Get(indicatorType + ".0.good.0.field").(string),
						Aggregation: d.Get(indicatorType + ".0.good.0.aggregation").(string),
						Filter:      getOrNilString(indicatorType+".0.good.0.filter", d),
						From:        getOrNilFloat(indicatorType+".0.good.0.from", d),
						To:          getOrNilFloat(indicatorType+".0.good.0.to", d),
					},
					Total: slo.IndicatorPropertiesHistogramParamsTotal{
						Field:       d.Get(indicatorType + ".0.total.0.field").(string),
						Aggregation: d.Get(indicatorType + ".0.total.0.aggregation").(string),
						Filter:      getOrNilString(indicatorType+".0.total.0.filter", d),
						From:        getOrNilFloat(indicatorType+".0.total.0.from", d),
						To:          getOrNilFloat(indicatorType+".0.total.0.to", d),
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
				Filter:      getOrNilString(indicatorType+".0.good.0.metrics."+idx+".filter", d),
			})
		}
		totalMetricsRaw := d.Get(indicatorType + ".0.total.0.metrics").([]interface{})
		var totalMetrics []slo.IndicatorPropertiesCustomMetricParamsTotalMetricsInner
		for n := range totalMetricsRaw {
			idx := fmt.Sprint(n)
			totalMetrics = append(totalMetrics, slo.IndicatorPropertiesCustomMetricParamsTotalMetricsInner{
				Name:        d.Get(indicatorType + ".0.total.0.metrics." + idx + ".name").(string),
				Field:       d.Get(indicatorType + ".0.total.0.metrics." + idx + ".field").(string),
				Aggregation: d.Get(indicatorType + ".0.total.0.metrics." + idx + ".aggregation").(string),
				Filter:      getOrNilString(indicatorType+".0.total.0.metrics."+idx+".filter", d),
			})
		}
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesCustomMetric: &slo.IndicatorPropertiesCustomMetric{
				Type: indicatorAddressToType[indicatorType],
				Params: slo.IndicatorPropertiesCustomMetricParams{
					Filter:         getOrNilString(indicatorType+".0.filter", d),
					Index:          d.Get(indicatorType + ".0.index").(string),
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

	default:
		return models.Slo{}, diag.Errorf("unknown indicator type %s", indicatorType)
	}

	timeWindow := slo.TimeWindow{
		Type:     d.Get("time_window.0.type").(string),
		Duration: d.Get("time_window.0.duration").(string),
	}

	objective := slo.Objective{
		Target:          d.Get("objective.0.target").(float64),
		TimesliceTarget: getOrNilFloat("objective.0.timeslice_target", d),
		TimesliceWindow: getOrNilString("objective.0.timeslice_window", d),
	}

	settings := slo.Settings{
		SyncDelay: getOrNilString("settings.0.sync_delay", d),
		Frequency: getOrNilString("settings.0.frequency", d),
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
		indicator = append(indicator, map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"good":            params.Good,
			"total":           params.Total,
			"timestamp_field": params.TimestampField,
		})

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
		indicator = append(indicator, map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"timestamp_field": params.TimestampField,
			"good":            good,
			"total":           total,
		})

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
		indicator = append(indicator, map[string]interface{}{
			"index":           params.Index,
			"filter":          params.Filter,
			"timestamp_field": params.TimestampField,
			"good":            good,
			"total":           total,
		})

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
}

var indicatorTypeToAddress = utils.FlipMap(indicatorAddressToType)
