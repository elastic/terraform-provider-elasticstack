package kibana

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/alerting"
)

type FakeAlertingAPI struct {
	RuleResponseProperties *alerting.RuleResponseProperties
	HttpResponse           *http.Response
	Error                  error
}

// The method stubs are generated initially by the VS Code Quick Fix for the below Blank Identifier definition.
// https://stackoverflow.com/a/77393824
var _ alerting.AlertingAPI = (*FakeAlertingAPI)(nil)

// CreateRule implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) CreateRule(ctx context.Context, spaceId string, ruleId string) alerting.ApiCreateRuleRequest {
	return alerting.ApiCreateRuleRequest{ApiService: f}
}

// CreateRuleExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) CreateRuleExecute(r alerting.ApiCreateRuleRequest) (*alerting.RuleResponseProperties, *http.Response, error) {
	return f.RuleResponseProperties, f.HttpResponse, f.Error
}

// DeleteRule implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) DeleteRule(ctx context.Context, ruleId string, spaceId string) alerting.ApiDeleteRuleRequest {
	panic("unimplemented")
}

// DeleteRuleExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) DeleteRuleExecute(r alerting.ApiDeleteRuleRequest) (*http.Response, error) {
	panic("unimplemented")
}

// DisableRule implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) DisableRule(ctx context.Context, ruleId string, spaceId string) alerting.ApiDisableRuleRequest {
	panic("unimplemented")
}

// DisableRuleExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) DisableRuleExecute(r alerting.ApiDisableRuleRequest) (*http.Response, error) {
	panic("unimplemented")
}

// EnableRule implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) EnableRule(ctx context.Context, ruleId string, spaceId string) alerting.ApiEnableRuleRequest {
	panic("unimplemented")
}

// EnableRuleExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) EnableRuleExecute(r alerting.ApiEnableRuleRequest) (*http.Response, error) {
	panic("unimplemented")
}

// FindRules implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) FindRules(ctx context.Context, spaceId string) alerting.ApiFindRulesRequest {
	panic("unimplemented")
}

// FindRulesExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) FindRulesExecute(r alerting.ApiFindRulesRequest) (*alerting.FindRules200Response, *http.Response, error) {
	panic("unimplemented")
}

// GetAlertingHealth implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) GetAlertingHealth(ctx context.Context, spaceId string) alerting.ApiGetAlertingHealthRequest {
	panic("unimplemented")
}

// GetAlertingHealthExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) GetAlertingHealthExecute(r alerting.ApiGetAlertingHealthRequest) (*alerting.GetAlertingHealth200Response, *http.Response, error) {
	panic("unimplemented")
}

// GetRule implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) GetRule(ctx context.Context, ruleId string, spaceId string) alerting.ApiGetRuleRequest {
	panic("unimplemented")
}

// GetRuleExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) GetRuleExecute(r alerting.ApiGetRuleRequest) (*alerting.RuleResponseProperties, *http.Response, error) {
	panic("unimplemented")
}

// GetRuleTypes implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) GetRuleTypes(ctx context.Context, spaceId string) alerting.ApiGetRuleTypesRequest {
	panic("unimplemented")
}

// GetRuleTypesExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) GetRuleTypesExecute(r alerting.ApiGetRuleTypesRequest) ([]alerting.GetRuleTypes200ResponseInner, *http.Response, error) {
	panic("unimplemented")
}

// LegacyCreateAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyCreateAlert(ctx context.Context, alertId string, spaceId string) alerting.ApiLegacyCreateAlertRequest {
	panic("unimplemented")
}

// LegacyCreateAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyCreateAlertExecute(r alerting.ApiLegacyCreateAlertRequest) (*alerting.AlertResponseProperties, *http.Response, error) {
	panic("unimplemented")
}

// LegacyDisableAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyDisableAlert(ctx context.Context, spaceId string, alertId string) alerting.ApiLegacyDisableAlertRequest {
	panic("unimplemented")
}

// LegacyDisableAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyDisableAlertExecute(r alerting.ApiLegacyDisableAlertRequest) (*http.Response, error) {
	panic("unimplemented")
}

// LegacyEnableAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyEnableAlert(ctx context.Context, spaceId string, alertId string) alerting.ApiLegacyEnableAlertRequest {
	panic("unimplemented")
}

// LegacyEnableAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyEnableAlertExecute(r alerting.ApiLegacyEnableAlertRequest) (*http.Response, error) {
	panic("unimplemented")
}

// LegacyFindAlerts implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyFindAlerts(ctx context.Context, spaceId string) alerting.ApiLegacyFindAlertsRequest {
	panic("unimplemented")
}

// LegacyFindAlertsExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyFindAlertsExecute(r alerting.ApiLegacyFindAlertsRequest) (*alerting.LegacyFindAlerts200Response, *http.Response, error) {
	panic("unimplemented")
}

// LegacyGetAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyGetAlert(ctx context.Context, spaceId string, alertId string) alerting.ApiLegacyGetAlertRequest {
	panic("unimplemented")
}

// LegacyGetAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyGetAlertExecute(r alerting.ApiLegacyGetAlertRequest) (*alerting.AlertResponseProperties, *http.Response, error) {
	panic("unimplemented")
}

// LegacyGetAlertTypes implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyGetAlertTypes(ctx context.Context, spaceId string) alerting.ApiLegacyGetAlertTypesRequest {
	panic("unimplemented")
}

// LegacyGetAlertTypesExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyGetAlertTypesExecute(r alerting.ApiLegacyGetAlertTypesRequest) ([]alerting.LegacyGetAlertTypes200ResponseInner, *http.Response, error) {
	panic("unimplemented")
}

// LegacyGetAlertingHealth implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyGetAlertingHealth(ctx context.Context, spaceId string) alerting.ApiLegacyGetAlertingHealthRequest {
	panic("unimplemented")
}

// LegacyGetAlertingHealthExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyGetAlertingHealthExecute(r alerting.ApiLegacyGetAlertingHealthRequest) (*alerting.LegacyGetAlertingHealth200Response, *http.Response, error) {
	panic("unimplemented")
}

// LegacyMuteAlertInstance implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyMuteAlertInstance(ctx context.Context, spaceId string, alertId string, alertInstanceId string) alerting.ApiLegacyMuteAlertInstanceRequest {
	panic("unimplemented")
}

// LegacyMuteAlertInstanceExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyMuteAlertInstanceExecute(r alerting.ApiLegacyMuteAlertInstanceRequest) (*http.Response, error) {
	panic("unimplemented")
}

// LegacyMuteAllAlertInstances implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyMuteAllAlertInstances(ctx context.Context, spaceId string, alertId string) alerting.ApiLegacyMuteAllAlertInstancesRequest {
	panic("unimplemented")
}

// LegacyMuteAllAlertInstancesExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyMuteAllAlertInstancesExecute(r alerting.ApiLegacyMuteAllAlertInstancesRequest) (*http.Response, error) {
	panic("unimplemented")
}

// LegacyUnmuteAlertInstance implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyUnmuteAlertInstance(ctx context.Context, spaceId string, alertId string, alertInstanceId string) alerting.ApiLegacyUnmuteAlertInstanceRequest {
	panic("unimplemented")
}

// LegacyUnmuteAlertInstanceExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyUnmuteAlertInstanceExecute(r alerting.ApiLegacyUnmuteAlertInstanceRequest) (*http.Response, error) {
	panic("unimplemented")
}

// LegacyUnmuteAllAlertInstances implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyUnmuteAllAlertInstances(ctx context.Context, spaceId string, alertId string) alerting.ApiLegacyUnmuteAllAlertInstancesRequest {
	panic("unimplemented")
}

// LegacyUnmuteAllAlertInstancesExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyUnmuteAllAlertInstancesExecute(r alerting.ApiLegacyUnmuteAllAlertInstancesRequest) (*http.Response, error) {
	panic("unimplemented")
}

// LegacyUpdateAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyUpdateAlert(ctx context.Context, spaceId string, alertId string) alerting.ApiLegacyUpdateAlertRequest {
	panic("unimplemented")
}

// LegacyUpdateAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegacyUpdateAlertExecute(r alerting.ApiLegacyUpdateAlertRequest) (*alerting.AlertResponseProperties, *http.Response, error) {
	panic("unimplemented")
}

// LegaryDeleteAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegaryDeleteAlert(ctx context.Context, spaceId string, alertId string) alerting.ApiLegaryDeleteAlertRequest {
	panic("unimplemented")
}

// LegaryDeleteAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) LegaryDeleteAlertExecute(r alerting.ApiLegaryDeleteAlertRequest) (*http.Response, error) {
	panic("unimplemented")
}

// MuteAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) MuteAlert(ctx context.Context, alertId string, ruleId string, spaceId string) alerting.ApiMuteAlertRequest {
	panic("unimplemented")
}

// MuteAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) MuteAlertExecute(r alerting.ApiMuteAlertRequest) (*http.Response, error) {
	panic("unimplemented")
}

// MuteAllAlerts implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) MuteAllAlerts(ctx context.Context, ruleId string, spaceId string) alerting.ApiMuteAllAlertsRequest {
	panic("unimplemented")
}

// MuteAllAlertsExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) MuteAllAlertsExecute(r alerting.ApiMuteAllAlertsRequest) (*http.Response, error) {
	panic("unimplemented")
}

// UnmuteAlert implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) UnmuteAlert(ctx context.Context, alertId string, ruleId string, spaceId string) alerting.ApiUnmuteAlertRequest {
	panic("unimplemented")
}

// UnmuteAlertExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) UnmuteAlertExecute(r alerting.ApiUnmuteAlertRequest) (*http.Response, error) {
	panic("unimplemented")
}

// UnmuteAllAlerts implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) UnmuteAllAlerts(ctx context.Context, ruleId string, spaceId string) alerting.ApiUnmuteAllAlertsRequest {
	panic("unimplemented")
}

// UnmuteAllAlertsExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) UnmuteAllAlertsExecute(r alerting.ApiUnmuteAllAlertsRequest) (*http.Response, error) {
	panic("unimplemented")
}

// UpdateRule implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) UpdateRule(ctx context.Context, ruleId string, spaceId string) alerting.ApiUpdateRuleRequest {
	return alerting.ApiUpdateRuleRequest{ApiService: f}
}

// UpdateRuleExecute implements alerting.AlertingAPI.
func (f *FakeAlertingAPI) UpdateRuleExecute(r alerting.ApiUpdateRuleRequest) (*alerting.RuleResponseProperties, *http.Response, error) {
	return f.RuleResponseProperties, f.HttpResponse, f.Error
}
