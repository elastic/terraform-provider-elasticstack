package kibana

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validateMinMaintenanceWindowServerVersion(serverVersion *version.Version) diag.Diagnostics {
	var maintenanceWindowPublicAPIMinSupportedVersion = version.Must(version.NewVersion("8.1.0"))
	var diags diag.Diagnostics

	if serverVersion.LessThan(maintenanceWindowPublicAPIMinSupportedVersion) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Maintenance window API not supported",
			Detail:   fmt.Sprintf(`The maintenance Window public API feature requires a minimum Elasticsearch version of "%s"`, maintenanceWindowPublicAPIMinSupportedVersion),
		})
		return diags
	}
	return nil
}

func ResourceMaintenanceWindow() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"space_id": {
			Description: "An identifier for the space. If space_id is not provided, the default space is used.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
			ForceNew:    true,
		},
		"title": {
			Description: "The name of the maintenance window.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"enabled": {
			Description: "Whether the current maintenance window is enabled.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"custom_schedule": {
			Description: "A set schedule over which the maintenance window applies.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"start": {
						Description: "The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-03-12T12:00:00.000Z`.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"duration": {
						Description:  "The duration of the schedule. It allows values in `<integer><unit>` format. `<unit>` is one of `d`, `h`, `m`, or `s` for hours, minutes, seconds. For example: `1d`, `5h`, `30m`, `5000s`.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: utils.StringIsAlertingDuration(),
					},
					"timezone": {
						Description: "The timezone of the schedule. The default timezone is UTC.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"recurring": {
						Type:     schema.TypeList,
						MinItems: 0,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"end": {
									Description: "The end date of a recurring schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-04-01T00:00:00.000Z`.",
									Type:        schema.TypeString,
									Optional:    true,
								},
								"every": {
									Description:  "The interval and frequency of a recurring schedule. It allows values in `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`.",
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: utils.StringIsMaintenanceWindowIntervalFrequency(),
								},
								"on_week_day": {
									Description: "The specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`) for a recurring schedule.",
									Type:        schema.TypeList,
									Optional:    true,
									Elem: &schema.Schema{
										Type:         schema.TypeString,
										ValidateFunc: utils.StringIsMaintenanceWindowOnWeekDay(),
									},
								},
								"on_month_day": {
									Description: "The specific days of the month for a recurring schedule. Valid values are 1-31.",
									Type:        schema.TypeList,
									Optional:    true,
									Elem: &schema.Schema{
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(1, 31),
									},
								},
								"on_month": {
									Description: "The specific months for a recurring schedule. Valid values are 1-12.",
									Type:        schema.TypeList,
									Optional:    true,
									Elem: &schema.Schema{
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(1, 12),
									},
								},
								"occurrences": {
									Description:  "The total number of recurrences of the schedule.",
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntAtLeast(0),
								},
							},
						},
					},
				},
			},
		},
		"scope": {
			Description: "An object that narrows the scope of what is affected by this maintenance window.",
			Type:        schema.TypeList,
			MinItems:    0,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"alerting": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"kql": {
									Description: "A filter written in Kibana Query Language (KQL).",
									Type:        schema.TypeString,
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}

	return &schema.Resource{
		Description: "Creates a Kibana Maintenance Window.",

		CreateContext: resourceMaintenanceWindowCreate,
		UpdateContext: resourceMaintenanceWindowUpdate,
		ReadContext:   resourceMaintenanceWindowRead,
		DeleteContext: resourceMaintenanceWindowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: apikeySchema,
	}
}

func getMaintenanceWindowFromResourceData(d *schema.ResourceData) (models.MaintenanceWindow, diag.Diagnostics) {
	var diags diag.Diagnostics
	maintenanceWindow := models.MaintenanceWindow{
		SpaceId: d.Get("space_id").(string),
		Title:   d.Get("title").(string),
		Enabled: d.Get("enabled").(bool),
	}

	if _, ok := d.GetOk("scope.alerting"); ok {

		alerting := &models.MaintenanceWindowAlertingScope{
			Kql: d.Get("scope.alerting.kql").(string),
		}

		maintenanceWindow.Scope = &models.MaintenanceWindowScope{
			Alerting: alerting,
		}
	}

	schedule, diags := getScheduleFromResourceData(d)

	if diags.HasError() {
		return models.MaintenanceWindow{}, diags
	}

	maintenanceWindow.CustomSchedule = schedule

	return maintenanceWindow, diags
}

func getScheduleFromResourceData(d *schema.ResourceData) (models.MaintenanceWindowSchedule, diag.Diagnostics) {
	schedule := models.MaintenanceWindowSchedule{
		Start:    d.Get("custom_schedule.0.start").(string),
		Duration: d.Get("custom_schedule.0.duration").(string),
	}

	// Explicitly set timezone if provided
	if timezone := getOrNilString("custom_schedule.0.timezone", d); timezone != nil && *timezone != "" {
		schedule.Timezone = timezone
	}

	if _, ok := d.GetOk("custom_schedule.0.recurring"); ok {
		recurring := models.MaintenanceWindowScheduleRecurring{}

		if v, ok := d.GetOk("custom_schedule.0.recurring.0.end"); ok {
			end := v.(string)
			recurring.End = &end
		}

		if v, ok := d.GetOk("custom_schedule.0.recurring.0.every"); ok {
			every := v.(string)
			recurring.Every = &every
		}

		if onWeekDay, ok := d.GetOk("custom_schedule.0.recurring.0.on_week_day"); ok {
			weekDayArray := []string{}
			for _, weekDay := range onWeekDay.([]interface{}) {
				weekDayArray = append(weekDayArray, weekDay.(string))
			}
			recurring.OnWeekDay = &weekDayArray
		}

		if onMonthDay, ok := d.GetOk("custom_schedule.0.recurring.0.on_month_day"); ok {
			monthDayArray := []float32{}
			for _, monthDay := range onMonthDay.([]interface{}) {
				monthDayArray = append(monthDayArray, float32(monthDay.(int)))
			}
			recurring.OnMonthDay = &monthDayArray
		}

		if onMonth, ok := d.GetOk("custom_schedule.0.recurring.0.on_month"); ok {
			monthArray := []float32{}
			for _, month := range onMonth.([]interface{}) {
				monthArray = append(monthArray, float32(month.(int)))
			}
			recurring.OnMonth = &monthArray

		}

		if v, ok := d.GetOk("custom_schedule.0.recurring.0.occurrences"); ok {
			occurrences := v.(int)
			recurring.Occurrences = utils.Pointer(float32(occurrences))
		}

		schedule.Recurring = &recurring
	}

	return schedule, nil
}

func resourceMaintenanceWindowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	diags = validateMinMaintenanceWindowServerVersion(serverVersion)
	if diags.HasError() {
		return diags
	}

	maintenanceWindow, diags := getMaintenanceWindowFromResourceData(d)

	if diags.HasError() {
		return diags
	}

	res, diags := kibana.CreateMaintenanceWindow(ctx, client, maintenanceWindow)

	if diags.HasError() {
		return diags
	}

	compositeID := &clients.CompositeId{ClusterId: res.SpaceId, ResourceId: res.MaintenanceWindowId}
	d.SetId(compositeID.String())

	return resourceMaintenanceWindowRead(ctx, d, meta)
}

func resourceMaintenanceWindowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	diags = validateMinMaintenanceWindowServerVersion(serverVersion)
	if diags.HasError() {
		return diags
	}

	maintenanceWindow, diags := getMaintenanceWindowFromResourceData(d)
	if diags.HasError() {
		return diags
	}

	compId, diags := clients.CompositeIdFromStr(d.Id())

	if diags.HasError() {
		return diags
	}

	maintenanceWindow.MaintenanceWindowId = compId.ResourceId

	_, diags = kibana.UpdateMaintenanceWindow(ctx, client, maintenanceWindow)

	if diags.HasError() {
		return diags
	}

	return resourceMaintenanceWindowRead(ctx, d, meta)
}

func resourceMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	diags = validateMinMaintenanceWindowServerVersion(serverVersion)
	if diags.HasError() {
		return diags
	}

	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	id := compId.ResourceId
	spaceId := compId.ClusterId

	maintenanceWindow, diags := kibana.GetMaintenanceWindow(ctx, client, id, spaceId)

	if maintenanceWindow == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}
	if err := d.Set("space_id", maintenanceWindow.SpaceId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("title", maintenanceWindow.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", maintenanceWindow.Enabled); err != nil {
		return diag.FromErr(err)
	}

	if maintenanceWindow.Scope != nil && maintenanceWindow.Scope.Alerting != nil {
		alertingScope := []interface{}{}
		alertingScope = append(alertingScope, map[string]interface{}{
			"kql": maintenanceWindow.Scope.Alerting.Kql,
		})

		scope := []interface{}{}
		scope = append(scope, map[string]interface{}{
			"alerting": alertingScope,
		})

		if err := d.Set("scope", scope); err != nil {
			return diag.FromErr(err)
		}
	}

	schedule := []interface{}{}
	recurring := []interface{}{}

	if maintenanceWindow.CustomSchedule.Recurring != nil {
		recurring = append(recurring, map[string]interface{}{
			"end":          maintenanceWindow.CustomSchedule.Recurring.End,
			"every":        maintenanceWindow.CustomSchedule.Recurring.Every,
			"on_week_day":  maintenanceWindow.CustomSchedule.Recurring.OnWeekDay,
			"on_month_day": maintenanceWindow.CustomSchedule.Recurring.OnMonthDay,
			"on_month":     maintenanceWindow.CustomSchedule.Recurring.OnMonth,
			"occurrences":  maintenanceWindow.CustomSchedule.Recurring.Occurrences,
		})
	} else {
		recurring = nil
	}

	schedule = append(schedule, map[string]interface{}{
		"start":     maintenanceWindow.CustomSchedule.Start,
		"duration":  maintenanceWindow.CustomSchedule.Duration,
		"timezone":  maintenanceWindow.CustomSchedule.Timezone,
		"recurring": recurring,
	})

	if err := d.Set("custom_schedule", schedule); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	diags = validateMinMaintenanceWindowServerVersion(serverVersion)
	if diags.HasError() {
		return diags
	}

	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	id := compId.ResourceId
	spaceId := compId.ClusterId

	if diags = kibana.DeleteMaintenanceWindow(ctx, client, id, spaceId); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}
