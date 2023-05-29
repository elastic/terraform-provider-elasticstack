package kibana

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/alerting"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ruleResponseToModel(spaceID string, res *alerting.RuleResponseProperties) *models.AlertingRule {
	if res == nil {
		return nil
	}

	actions := []models.AlertingRuleAction{}
	for _, action := range res.Actions {
		actions = append(actions, models.AlertingRuleAction{
			Group:  *action.Group,
			ID:     *action.Id,
			Params: action.Params,
		})
	}

	return &models.AlertingRule{
		ID:         res.Id,
		SpaceID:    spaceID,
		Name:       res.Name,
		Consumer:   res.Consumer,
		NotifyWhen: string(unwrapOptionalField(res.NotifyWhen)),
		Params:     res.Params,
		RuleTypeID: res.RuleTypeId,
		Schedule: models.AlertingRuleSchedule{
			Interval: unwrapOptionalField(res.Schedule.Interval),
		},
		Enabled:         &res.Enabled,
		Tags:            res.Tags,
		Throttle:        res.Throttle.Get(),
		ScheduledTaskID: res.ScheduledTaskId,
		ExecutionStatus: models.AlertingRuleExecutionStatus{
			LastExecutionDate: res.ExecutionStatus.LastExecutionDate,
			Status:            res.ExecutionStatus.Status,
		},
		Actions: actions,
	}
}

// Maps the rule actions to the struct required by the request model (ActionsInner)
func ruleActionsToActionsInner(ruleActions []models.AlertingRuleAction) []alerting.ActionsInner {
	actions := []alerting.ActionsInner{}
	for index := range ruleActions {
		action := ruleActions[index]
		actions = append(actions, alerting.ActionsInner{
			Group:  &action.Group,
			Id:     &action.ID,
			Params: action.Params,
		})
	}
	return actions
}

func CreateAlertingRule(ctx context.Context, apiClient *clients.ApiClient, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)

	reqModel := alerting.CreateRuleRequest{
		Consumer:   rule.Consumer,
		Actions:    ruleActionsToActionsInner(rule.Actions),
		Enabled:    rule.Enabled,
		Name:       rule.Name,
		NotifyWhen: (*alerting.NotifyWhen)(&rule.NotifyWhen),
		Params:     rule.Params,
		RuleTypeId: rule.RuleTypeID,
		Schedule: alerting.Schedule{
			Interval: &rule.Schedule.Interval,
		},
		Tags:     rule.Tags,
		Throttle: *alerting.NewNullableString(rule.Throttle),
	}
	req := client.CreateRule(ctxWithAuth, rule.SpaceID, "").KbnXsrf("true").CreateRuleRequest(reqModel)
	ruleRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	return ruleResponseToModel(rule.SpaceID, ruleRes), utils.CheckHttpError(res, "Unabled to create alerting rule")
}

func UpdateAlertingRule(ctx context.Context, apiClient *clients.ApiClient, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)

	reqModel := alerting.UpdateRuleRequest{
		Actions:    ruleActionsToActionsInner((rule.Actions)),
		Name:       rule.Name,
		NotifyWhen: (*alerting.NotifyWhen)(&rule.NotifyWhen),
		Params:     rule.Params,
		Schedule: alerting.Schedule{
			Interval: &rule.Schedule.Interval,
		},
		Tags:     rule.Tags,
		Throttle: *alerting.NewNullableString(rule.Throttle),
	}
	req := client.UpdateRule(ctxWithAuth, rule.ID, rule.SpaceID).KbnXsrf("true").UpdateRuleRequest(reqModel)
	ruleRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	if diags := utils.CheckHttpError(res, "Unable to update alerting rule"); diags.HasError() {
		return nil, diags
	}

	shouldBeEnabled := rule.Enabled != nil && *rule.Enabled
	if shouldBeEnabled && !ruleRes.Enabled {
		res, err := client.EnableRule(ctxWithAuth, rule.ID, rule.SpaceID).KbnXsrf("true").Execute()
		if err != nil && res == nil {
			return nil, diag.FromErr(err)
		}

		if diags := utils.CheckHttpError(res, "Unable to enable alerting rule"); diags.HasError() {
			return nil, diag.FromErr(err)
		}
	}

	if !shouldBeEnabled && ruleRes.Enabled {
		res, err := client.DisableRule(ctxWithAuth, rule.ID, rule.SpaceID).KbnXsrf("true").Execute()
		if err != nil && res == nil {
			return nil, diag.FromErr(err)
		}

		if diags := utils.CheckHttpError(res, "Unable to disable alerting rule"); diags.HasError() {
			return nil, diag.FromErr(err)
		}
	}

	return ruleResponseToModel(rule.SpaceID, ruleRes), diag.Diagnostics{}
}

func GetAlertingRule(ctx context.Context, apiClient *clients.ApiClient, id, spaceID string) (*models.AlertingRule, diag.Diagnostics) {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)
	req := client.GetRule(ctxWithAuth, id, spaceID)
	ruleRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return ruleResponseToModel(spaceID, ruleRes), utils.CheckHttpError(res, "Unabled to get alerting rule")
}

func DeleteAlertingRule(ctx context.Context, apiClient *clients.ApiClient, ruleId string, spaceId string) diag.Diagnostics {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)
	req := client.DeleteRule(ctxWithAuth, ruleId, spaceId).KbnXsrf("true")
	res, err := req.Execute()
	if err != nil && res == nil {
		return diag.FromErr(err)
	}

	defer res.Body.Close()
	return utils.CheckHttpError(res, "Unabled to delete alerting rule")
}
