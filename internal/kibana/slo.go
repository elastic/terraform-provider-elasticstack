package kibana

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceSlo() *schema.Resource {
	sloSchema := map[string]*schema.Schema{
		"id": {
			Description: "An ID (8 and 36 characters). If omitted, a UUIDv1 will be generated server-side.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
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
		"indicator": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"sli.kql.custom", "sli.apm.transactionErrorRate", "sli.apm.transactionDuration"}, false),
					},
					"params": {
						Type:     schema.TypeList,
						Required: true,
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
								"service": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"environment": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"transaction_type": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"transaction_name": {
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
								},
								"threshold": {
									Type:     schema.TypeInt,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},
		"time_window": {
			Description: "Currently support calendar aligned and rolling time windows. Any duration greater than 1 day can be used: days, weeks, months, quarters, years. Rolling time window requires a duration, e.g. 1w for one week, and isRolling: true. SLOs defined with such time window, will only consider the SLI data from the last duration period as a moving window. Calendar aligned time window requires a duration, limited to 1M for monthly or 1w for weekly, and isCalendar: true.",
			Type:        schema.TypeList,
			Required:    true,
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
			Description:  "An occurrences budgeting method uses the number of good and total events during the time window. A timeslices budgeting method uses the number of good slices and total slices during the time window. A slice is an arbitrary time window (smaller than the overall SLO time window) that is either considered good or bad, calculated from the timeslice threshold and the ratio of good over total events that happened during the slice window. A budgeting method is required and must be either occurrences or timeslices.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"occurrences", "timeslices"}, false),
		},
		"objective": {
			Description: "The target objective is the value the SLO needs to meet during the time window. If a timeslices budgeting method is used, we also need to define the timesliceTarget which can be different than the overall SLO target.",
			Type:        schema.TypeList,
			Required:    true,
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
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"sync_delay": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"frequency": {
						Type:     schema.TypeString,
						Optional: true,
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
	}

	return &schema.Resource{
		Description: "Creates an SLO.",

		CreateContext: resourceSloCreate,
		UpdateContext: resourceSloUpdate,
		ReadContext:   resourceSloRead,
		DeleteContext: resourceSloDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sloSchema,
	}
}

// surely this kind of thing exists in the SDK?
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
	indicatorType := d.Get("indicator.0.type").(string)

	switch indicatorType {
	case "sli.kql.custom":
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesCustomKql: &slo.IndicatorPropertiesCustomKql{
				Type: indicatorType,
				Params: slo.IndicatorPropertiesCustomKqlParams{
					Index:          d.Get("indicator.0.params.0.index").(string),
					Filter:         getOrNilString("indicator.0.params.0.filter", d),
					Good:           getOrNilString("indicator.0.params.0.good", d),
					Total:          getOrNilString("indicator.0.params.0.total", d),
					TimestampField: d.Get("indicator.0.params.0.timestamp_field").(string),
				},
			},
		}

	case "sli.apm.transactionErrorRate":
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
				Type: indicatorType,
				Params: slo.IndicatorPropertiesApmAvailabilityParams{
					Service:         d.Get("indicator.0.params.0.service").(string),
					Environment:     d.Get("indicator.0.params.0.environment").(string),
					TransactionType: d.Get("indicator.0.params.0.transaction_type").(string),
					TransactionName: d.Get("indicator.0.params.0.transaction_name").(string),
					Filter:          getOrNilString("indicator.0.params.0.filter", d),
					Index:           d.Get("indicator.0.params.0.index").(string),
				},
			},
		}

	case "sli.apm.transactionDuration":
		indicator = slo.SloResponseIndicator{
			IndicatorPropertiesApmLatency: &slo.IndicatorPropertiesApmLatency{
				Type: indicatorType,
				Params: slo.IndicatorPropertiesApmLatencyParams{
					Service:         d.Get("indicator.0.params.0.service").(string),
					Environment:     d.Get("indicator.0.params.0.environment").(string),
					TransactionType: d.Get("indicator.0.params.0.transaction_type").(string),
					TransactionName: d.Get("indicator.0.params.0.transaction_name").(string),
					Filter:          getOrNilString("indicator.0.params.0.filter", d),
					Index:           d.Get("indicator.0.params.0.index").(string),
					Threshold:       float64(d.Get("indicator.0.params.0.threshold").(int)),
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

	var settings slo.Settings
	if _, ok := d.GetOk("settings"); ok {
		settings = slo.Settings{
			SyncDelay: getOrNilString("settings.0.sync_delay", d),
			Frequency: getOrNilString("settings.0.frequency", d),
		}
	}

	slo := models.Slo{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Indicator:       indicator,
		TimeWindow:      timeWindow,
		BudgetingMethod: d.Get("budgeting_method").(string),
		Objective:       objective,
		Settings:        &settings,
		SpaceID:         d.Get("space_id").(string),
	}

	return slo, diags
}

func resourceSloCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	slo, diags := getSloFromResourceData(d)
	if diags.HasError() {
		return diags
	}

	res, diags := kibana.CreateSlo(ctx, client, slo)

	if diags.HasError() {
		return diags
	}

	id, diags := client.ID(ctx, res.ID)
	if diags.HasError() {
		return diags
	}

	id.ResourceId = res.ID
	d.SetId(id.String())

	return resourceSloRead(ctx, d, meta)
}

func resourceSloUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	slo, diags := getSloFromResourceData(d)
	if diags.HasError() {
		return diags
	}

	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	slo.ID = compId.ResourceId

	res, diags := kibana.UpdateSlo(ctx, client, slo)

	if diags.HasError() {
		return diags
	}

	id := &clients.CompositeId{ClusterId: slo.SpaceID, ResourceId: res.ID}
	d.SetId(id.String())

	return resourceSloRead(ctx, d, meta)
}

func resourceSloRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
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
		return diags
	}
	if diags.HasError() {
		return diags
	}

	indicators := []interface{}{}
	if s.Indicator.IndicatorPropertiesApmAvailability != nil {
		params := s.Indicator.IndicatorPropertiesApmAvailability.Params
		indicators = append(indicators, map[string]interface{}{
			"type": s.Indicator.IndicatorPropertiesApmAvailability.Type,
			"params": []map[string]interface{}{{
				"environment":      params.Environment,
				"service":          params.Service,
				"transaction_type": params.TransactionType,
				"transaction_name": params.TransactionName,
				"index":            params.Index,
				"filter":           params.Filter,
			}},
		})
	} else if s.Indicator.IndicatorPropertiesApmLatency != nil {
		params := s.Indicator.IndicatorPropertiesApmLatency.Params
		indicators = append(indicators, map[string]interface{}{
			"type": s.Indicator.IndicatorPropertiesApmLatency.Type,
			"params": []map[string]interface{}{{
				"environment":      params.Environment,
				"service":          params.Service,
				"transaction_type": params.TransactionType,
				"transaction_name": params.TransactionName,
				"index":            params.Index,
				"filter":           params.Filter,
				"threshold":        params.Threshold,
			}},
		})
	} else if s.Indicator.IndicatorPropertiesCustomKql != nil {
		params := s.Indicator.IndicatorPropertiesCustomKql.Params
		indicators = append(indicators, map[string]interface{}{
			"type": s.Indicator.IndicatorPropertiesCustomKql.Type,
			"params": []map[string]interface{}{{
				"index":           params.Index,
				"filter":          params.Filter,
				"good":            params.Filter,
				"total":           params.Total,
				"timestamp_field": params.TimestampField,
			}},
		})
	} else {
		return diag.Errorf("unknown indicator type")
	}
	if err := d.Set("indicator", indicators); err != nil {
		return diag.FromErr(err)
	}

	time_window := []interface{}{}
	time_window = append(time_window, map[string]interface{}{
		"duration": s.TimeWindow.Duration,
		"type":     s.TimeWindow.Type,
	})
	if err := d.Set("time_window", time_window); err != nil {
		return diag.FromErr(err)
	}

	objective := []interface{}{}
	objective = append(objective, map[string]interface{}{
		"target":           s.Objective.Target,
		"timeslice_target": s.Objective.TimesliceTarget,
		"timeslice_window": s.Objective.TimesliceWindow,
	})
	if err := d.Set("objective", objective); err != nil {
		return diag.FromErr(err)
	}

	settings := []interface{}{}
	settings = append(settings, map[string]interface{}{
		"sync_delay": s.Settings.SyncDelay,
		"frequency":  s.Settings.Frequency,
	})
	if err := d.Set("settings", settings); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("id", s.ID); err != nil {
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

	return diags
}

func resourceSloDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
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
