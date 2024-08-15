package kibana

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSloV0() *schema.Resource {
	var indicatorAddresses []string
	for i := range indicatorAddressToType {
		indicatorAddresses = append(indicatorAddresses, i)
	}

	sloSchema := map[string]*schema.Schema{
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
			Description: "Optional group by field to use to generate an SLO per distinct value.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    false,
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

	return &schema.Resource{
		Schema: sloSchema,
	}
}
