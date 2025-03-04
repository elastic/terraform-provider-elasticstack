package agent_policy

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// type globalDataTagModel struct {
// 	Name  types.String `tfsdk:"name"`
// 	Value types.String `tfsdk:"value"`
// }

// func newGlobalDataTagModel(data struct {
// 	Name  string                                 "json:\"name\""
// 	Value kbapi.AgentPolicy_GlobalDataTags_Value "json:\"value\""
// }) globalDataTagModel {
// 	val, err := data.Value.AsAgentPolicyGlobalDataTagsValue0()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return globalDataTagModel{
// 		Name:  types.StringValue(data.Name),
// 		Value: types.StringValue(val),
// 	}
// }

type agentPolicyModel struct {
	ID                 types.String `tfsdk:"id"`
	PolicyID           types.String `tfsdk:"policy_id"`
	Name               types.String `tfsdk:"name"`
	Namespace          types.String `tfsdk:"namespace"`
	Description        types.String `tfsdk:"description"`
	DataOutputId       types.String `tfsdk:"data_output_id"`
	MonitoringOutputId types.String `tfsdk:"monitoring_output_id"`
	FleetServerHostId  types.String `tfsdk:"fleet_server_host_id"`
	DownloadSourceId   types.String `tfsdk:"download_source_id"`
	MonitorLogs        types.Bool   `tfsdk:"monitor_logs"`
	MonitorMetrics     types.Bool   `tfsdk:"monitor_metrics"`
	SysMonitoring      types.Bool   `tfsdk:"sys_monitoring"`
	SkipDestroy        types.Bool   `tfsdk:"skip_destroy"`
	GlobalDataTags     types.String `tfsdk:"global_data_tags"`
}

func (model *agentPolicyModel) populateFromAPI(ctx context.Context, data *kbapi.AgentPolicy, serverVersion *version.Version) diag.Diagnostics {
	if data == nil {
		return nil
	}

	model.ID = types.StringValue(data.Id)
	model.PolicyID = types.StringValue(data.Id)
	model.DataOutputId = types.StringPointerValue(data.DataOutputId)
	model.Description = types.StringPointerValue(data.Description)
	model.DownloadSourceId = types.StringPointerValue(data.DownloadSourceId)
	model.FleetServerHostId = types.StringPointerValue(data.FleetServerHostId)

	if data.MonitoringEnabled != nil {
		if slices.Contains(*data.MonitoringEnabled, kbapi.AgentPolicyMonitoringEnabledLogs) {
			model.MonitorLogs = types.BoolValue(true)
		}
		if slices.Contains(*data.MonitoringEnabled, kbapi.AgentPolicyMonitoringEnabledMetrics) {
			model.MonitorMetrics = types.BoolValue(true)
		}
	}
	if !utils.IsKnown(model.MonitorLogs) {
		model.MonitorLogs = types.BoolValue(false)
	}
	if !utils.IsKnown(model.MonitorLogs) {
		model.MonitorLogs = types.BoolValue(false)
	}

	model.MonitoringOutputId = types.StringPointerValue(data.MonitoringOutputId)
	model.Name = types.StringValue(data.Name)
	model.Namespace = types.StringValue(data.Namespace)
	if serverVersion.GreaterThanOrEqual(MinVersionGlobalDataTags) && utils.Deref(data.GlobalDataTags) != nil {
		diags := diag.Diagnostics{}
		d, err := json.Marshal(data.GlobalDataTags)
		if err != nil {
			diags.AddError("Failed to marshal global data tags", err.Error())
			return diags
		}
		model.GlobalDataTags = types.StringValue(string(d))
	}

	return nil
}

func (model *agentPolicyModel) toAPICreateModel(ctx context.Context, serverVersion *version.Version) (kbapi.PostFleetAgentPoliciesJSONRequestBody, diag.Diagnostics) {
	monitoring := make([]kbapi.PostFleetAgentPoliciesJSONBodyMonitoringEnabled, 0, 2)

	if model.MonitorLogs.ValueBool() {
		monitoring = append(monitoring, kbapi.PostFleetAgentPoliciesJSONBodyMonitoringEnabledLogs)
	}
	if model.MonitorMetrics.ValueBool() {
		monitoring = append(monitoring, kbapi.PostFleetAgentPoliciesJSONBodyMonitoringEnabledMetrics)
	}

	body := kbapi.PostFleetAgentPoliciesJSONRequestBody{
		DataOutputId:       model.DataOutputId.ValueStringPointer(),
		Description:        model.Description.ValueStringPointer(),
		DownloadSourceId:   model.DownloadSourceId.ValueStringPointer(),
		FleetServerHostId:  model.FleetServerHostId.ValueStringPointer(),
		Id:                 model.PolicyID.ValueStringPointer(),
		MonitoringEnabled:  &monitoring,
		MonitoringOutputId: model.MonitoringOutputId.ValueStringPointer(),
		Name:               model.Name.ValueString(),
		Namespace:          model.Namespace.ValueString(),
	}

	if len(model.GlobalDataTags.ValueString()) > 0 {
		var diags diag.Diagnostics
		if serverVersion.LessThan(MinVersionGlobalDataTags) {
			diags.AddError("global_data_tags ES version error", "Global data tags are only supported in Elastic Stack 8.15.0 and above")
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}

		str := model.GlobalDataTags.ValueStringPointer()
		var items []struct {
			Name  string                                                    `json:"name"`
			Value kbapi.PostFleetAgentPoliciesJSONBody_GlobalDataTags_Value `json:"value"`
		}

		err := json.Unmarshal([]byte(utils.Deref(str)), &items)
		if err != nil {
			diags.AddError(err.Error(), "")
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}
		*body.GlobalDataTags = items
	}
	return body, nil
}

func (model *agentPolicyModel) toAPIUpdateModel(ctx context.Context, serverVersion *version.Version) (kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody, diag.Diagnostics) {
	monitoring := make([]kbapi.PutFleetAgentPoliciesAgentpolicyidJSONBodyMonitoringEnabled, 0, 2)
	if model.MonitorLogs.ValueBool() {
		monitoring = append(monitoring, kbapi.Logs)
	}
	if model.MonitorMetrics.ValueBool() {
		monitoring = append(monitoring, kbapi.Metrics)
	}

	body := kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{
		DataOutputId:       model.DataOutputId.ValueStringPointer(),
		Description:        model.Description.ValueStringPointer(),
		DownloadSourceId:   model.DownloadSourceId.ValueStringPointer(),
		FleetServerHostId:  model.FleetServerHostId.ValueStringPointer(),
		MonitoringEnabled:  &monitoring,
		MonitoringOutputId: model.MonitoringOutputId.ValueStringPointer(),
		Name:               model.Name.ValueString(),
		Namespace:          model.Namespace.ValueString(),
	}

	if len(model.GlobalDataTags.ValueString()) > 0 {
		var diags diag.Diagnostics
		if serverVersion.LessThan(MinVersionGlobalDataTags) {
			diags.AddError("global_data_tags ES version error", "Global data tags are only supported in Elastic Stack 8.15.0 and above")
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}
		str := model.GlobalDataTags.ValueStringPointer()
		var items []struct {
			Name  string                                                                `json:"name"`
			Value kbapi.PutFleetAgentPoliciesAgentpolicyidJSONBody_GlobalDataTags_Value `json:"value"`
		}
		err := json.Unmarshal([]byte(utils.Deref(str)), &items)
		if err != nil {
			diags.AddError(err.Error(), "")
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}
		*body.GlobalDataTags = items
	}

	return body, nil
}
