package kibana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var alertDelayMinSupportedVersion = version.Must(version.NewVersion("8.13.0"))

// when notify_when and throttle became optional
var frequencyMinSupportedVersion = version.Must(version.NewVersion("8.6.0"))

func ResourceAlertingRule() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"rule_id": {
			Description: "A UUID v1 or v4 to use instead of a randomly generated ID.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"space_id": {
			Description: "An identifier for the space. If space_id is not provided, the default space is used.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"consumer": {
			Description: "The name of the application or feature that owns the rule.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"notify_when": {
			Description:  "Required until v8.6.0. Deprecated in v8.13.0. Use the `notify_when` property in the action `frequency` object instead. Defines how often alerts generate actions. Valid values include: `onActionGroupChange`: Actions run when the alert status changes; `onActiveAlert`: Actions run when the alert becomes active and at each check interval while the rule conditions are met; `onThrottleInterval`: Actions run when the alert becomes active and at the interval specified in the throttle property while the rule conditions are met. NOTE: This is a rule level property; if you update the rule in Kibana, it is automatically changed to use action-specific `notify_when` values.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"onActionGroupChange", "onActiveAlert", "onThrottleInterval"}, false),
		},
		"params": {
			Description:      "The rule parameters, which differ for each rule type.",
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
		},
		"rule_type_id": {
			Description: "The ID of the rule type that you want to call when the rule is scheduled to run. For more information about the valid values, list the rule types using [Get rule types API](https://www.elastic.co/guide/en/kibana/master/list-rule-types-api.html) or refer to the [Rule types documentation](https://www.elastic.co/guide/en/kibana/master/rule-types.html).",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"interval": {
			Description:  "The check interval, which specifies how frequently the rule conditions are checked. The interval must be specified in seconds, minutes, hours or days.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: utils.StringIsDuration,
		},
		"actions": {
			Description: "An action that runs under defined conditions.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"group": {
						Description: "The group name, which affects when the action runs (for example, when the threshold is met or when the alert is recovered). Each rule type has a list of valid action group names.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "default",
					},
					"id": {
						Description: "The identifier for the connector saved object.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"params": {
						Description:      "The parameters for the action, which are sent to the connector.",
						Type:             schema.TypeString,
						Required:         true,
						ValidateFunc:     validation.StringIsJSON,
						DiffSuppressFunc: utils.DiffJsonSuppress,
					},
					"frequency": {
						Description: "The properties that affect how often actions are generated. If the rule type supports setting summary to true, the action can be a summary of alerts at the specified notification interval. Otherwise, an action runs for each alert at the specified notification interval. NOTE: You cannot specify these parameters when `notify_when` or `throttle` are defined at the rule level.",
						Type:        schema.TypeList,
						MinItems:    0,
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"summary": {
									Description: "Indicates whether the action is a summary.",
									Type:        schema.TypeBool,
									Required:    true,
								},
								"notify_when": {
									Description:  "Defines how often alerts generate actions. Valid values include: `onActionGroupChange`: Actions run when the alert status changes; `onActiveAlert`: Actions run when the alert becomes active and at each check interval while the rule conditions are met; `onThrottleInterval`: Actions run when the alert becomes active and at the interval specified in the throttle property while the rule conditions are met. NOTE: This is a rule level property; if you update the rule in Kibana, it is automatically changed to use action-specific `notify_when` values.",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice([]string{"onActionGroupChange", "onActiveAlert", "onThrottleInterval"}, false),
								},
								"throttle": {
									Description:  "Defines how often an alert generates repeated actions. This custom action interval must be specified in seconds, minutes, hours, or days. For example, 10m or 1h. This property is applicable only if `notify_when` is `onThrottleInterval`. NOTE: This is a rule level property; if you update the rule in Kibana, it is automatically changed to use action-specific `throttle` values.",
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: utils.StringIsDuration,
								},
							},
						},
					},
				},
			},
		},
		"enabled": {
			Description: "Indicates if you want to run the rule on an interval basis.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"tags": {
			Description: "A list of tag names that are applied to the rule.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"throttle": {
			Description:  "Deprecated in 8.13.0. Defines how often an alert generates repeated actions. This custom action interval must be specified in seconds, minutes, hours, or days. For example, 10m or 1h. This property is applicable only if `notify_when` is `onThrottleInterval`. NOTE: This is a rule level property; if you update the rule in Kibana, it is automatically changed to use action-specific `throttle` values.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: utils.StringIsDuration,
		},
		"scheduled_task_id": {
			Description: "ID of the scheduled task that will execute the alert.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"last_execution_status": {
			Description: "Status of the last execution of this rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"last_execution_date": {
			Description: "Date of the last execution of this rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"alert_delay": {
			Description: "A number that indicates how many consecutive runs need to meet the rule conditions for an alert to occur.",
			Type:        schema.TypeFloat,
			Optional:    true,
		},
	}

	return &schema.Resource{
		Description: "Creates a Kibana rule. See https://www.elastic.co/guide/en/kibana/master/create-rule-api.html",

		CreateContext: resourceRuleCreate,
		UpdateContext: resourceRuleUpdate,
		ReadContext:   resourceRuleRead,
		DeleteContext: resourceRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: apikeySchema,
	}
}

func getAlertingRuleFromResourceData(d *schema.ResourceData, serverVersion *version.Version) (models.AlertingRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	rule := models.AlertingRule{
		SpaceID:    d.Get("space_id").(string),
		Name:       d.Get("name").(string),
		Consumer:   d.Get("consumer").(string),
		RuleTypeID: d.Get("rule_type_id").(string),
		Schedule: models.AlertingRuleSchedule{
			Interval: d.Get("interval").(string),
		},
	}

	// Explicitly set rule id if provided, otherwise we'll use the autogenerated ID from the Kibana API response
	if ruleID := getOrNilString("rule_id", d); ruleID != nil && *ruleID != "" {
		rule.RuleID = *ruleID
	}

	paramsStr := d.Get("params")
	params := map[string]interface{}{}
	if err := json.NewDecoder(strings.NewReader(paramsStr.(string))).Decode(&params); err != nil {
		return models.AlertingRule{}, diag.FromErr(err)
	}
	rule.Params = params

	if v, ok := d.GetOk("enabled"); ok {
		e := v.(bool)
		rule.Enabled = &e
	}

	if v, ok := d.GetOk("throttle"); ok {
		t := v.(string)
		rule.Throttle = &t
	}

	if v, ok := d.GetOk("notify_when"); ok {
		rule.NotifyWhen = utils.Pointer(v.(string))
	} else {
		if serverVersion.LessThan(frequencyMinSupportedVersion) {
			return models.AlertingRule{}, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "notify_when is required until v8.6",
					Detail:   "notify_when is required until v8.6",
				},
			}
		}
	}

	if v, ok := d.GetOk("alert_delay"); ok {
		if serverVersion.LessThan(alertDelayMinSupportedVersion) {
			return models.AlertingRule{}, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "alert_delay is only supported for Elasticsearch v8.13 or higher",
					Detail:   "alert_delay is only supported for Elasticsearch v8.13 or higher",
				},
			}
		}

		rule.AlertDelay = utils.Pointer(float32(v.(float64)))
	}

	actions, diags := getActionsFromResourceData(d, serverVersion)
	if diags.HasError() {
		return models.AlertingRule{}, diags
	}
	rule.Actions = actions

	if tags, ok := d.GetOk("tags"); ok {
		for _, t := range tags.([]interface{}) {
			rule.Tags = append(rule.Tags, t.(string))
		}
	}

	return rule, diags
}

func getActionsFromResourceData(d *schema.ResourceData, serverVersion *version.Version) ([]models.AlertingRuleAction, diag.Diagnostics) {
	actions := []models.AlertingRuleAction{}
	if v, ok := d.GetOk("actions"); ok {
		resourceActions := v.([]interface{})
		for i, a := range resourceActions {
			action := a.(map[string]interface{})
			paramsStr := action["params"].(string)
			var params map[string]interface{}
			err := json.Unmarshal([]byte(paramsStr), &params)
			if err != nil {
				return []models.AlertingRuleAction{}, diag.FromErr(err)
			}

			a := models.AlertingRuleAction{
				Group:  action["group"].(string),
				ID:     action["id"].(string),
				Params: params,
			}

			currentAction := fmt.Sprintf("actions.%d", i)

			if _, ok := d.GetOk(currentAction + ".frequency"); ok {
				if serverVersion.LessThan(frequencyMinSupportedVersion) {
					return []models.AlertingRuleAction{}, diag.Errorf("actions.frequency is only supported for Elasticsearch v8.6 or higher")
				}

				frequency := models.AlertingRuleActionFrequency{
					Summary:    d.Get(currentAction + ".frequency.0.summary").(bool),
					NotifyWhen: d.Get(currentAction + ".frequency.0.notify_when").(string),
				}

				if throttle := getOrNilString(currentAction+".frequency.0.throttle", d); throttle != nil && *throttle != "" {
					frequency.Throttle = throttle
				}

				a.Frequency = &frequency
			}

			actions = append(actions, a)
		}
	}

	return actions, nil
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	rule, diags := getAlertingRuleFromResourceData(d, serverVersion)
	if diags.HasError() {
		return diags
	}

	res, diags := kibana.CreateAlertingRule(ctx, client, rule)

	if diags.HasError() {
		return diags
	}

	compositeID := &clients.CompositeId{ClusterId: rule.SpaceID, ResourceId: res.RuleID}
	d.SetId(compositeID.String())

	return resourceRuleRead(ctx, d, meta)
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	rule, diags := getAlertingRuleFromResourceData(d, serverVersion)
	if diags.HasError() {
		return diags
	}

	res, diags := kibana.UpdateAlertingRule(ctx, client, rule)

	if diags.HasError() {
		return diags
	}

	compositeID := &clients.CompositeId{ClusterId: rule.SpaceID, ResourceId: res.RuleID}
	d.SetId(compositeID.String())

	return resourceRuleRead(ctx, d, meta)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	rule, diags := kibana.GetAlertingRule(ctx, client, id, spaceId)
	if rule == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	// set the fields
	if err := d.Set("rule_id", rule.RuleID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("space_id", rule.SpaceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", rule.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("consumer", rule.Consumer); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notify_when", rule.NotifyWhen); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rule_type_id", rule.RuleTypeID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interval", rule.Schedule.Interval); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", rule.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", rule.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("throttle", rule.Throttle); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("scheduled_task_id", rule.ScheduledTaskID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_execution_status", rule.ExecutionStatus.Status); err != nil {
		return diag.FromErr(err)
	}
	if rule.ExecutionStatus.LastExecutionDate != nil {
		if err := d.Set("last_execution_date", rule.ExecutionStatus.LastExecutionDate.Format("2006-01-02 15:04:05.999 -0700 MST")); err != nil {
			return diag.FromErr(err)
		}
	}

	params, err := json.Marshal(rule.Params)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("params", string(params)); err != nil {
		return diag.FromErr(err)
	}

	actions := []interface{}{}
	for _, action := range rule.Actions {
		params, err := json.Marshal(action.Params)
		if err != nil {
			return diag.FromErr(err)
		}

		frequency := []interface{}{}

		if action.Frequency != nil {
			frequency = append(frequency, map[string]interface{}{
				"summary":     action.Frequency.Summary,
				"notify_when": action.Frequency.NotifyWhen,
				"throttle":    action.Frequency.Throttle,
			})
		} else {
			frequency = nil
		}

		actions = append(actions, map[string]interface{}{
			"group":     action.Group,
			"id":        action.ID,
			"params":    string(params),
			"frequency": frequency,
		})
	}

	if err := d.Set("actions", actions); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	spaceId := d.Get("space_id").(string)

	if diags = kibana.DeleteAlertingRule(ctx, client, compId.ResourceId, spaceId); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}
