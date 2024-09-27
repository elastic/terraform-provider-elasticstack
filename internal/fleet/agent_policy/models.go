package agent_policy

import (
	"slices"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
}

func (model *agentPolicyModel) populateFromAPI(data *fleetapi.AgentPolicy) {
	if data == nil {
		return
	}

	model.ID = types.StringValue(data.Id)
	model.PolicyID = types.StringValue(data.Id)
	model.DataOutputId = types.StringPointerValue(data.DataOutputId)
	model.Description = types.StringPointerValue(data.Description)
	model.DownloadSourceId = types.StringPointerValue(data.DownloadSourceId)
	model.FleetServerHostId = types.StringPointerValue(data.FleetServerHostId)

	if data.MonitoringEnabled != nil {
		if slices.Contains(*data.MonitoringEnabled, fleetapi.AgentPolicyMonitoringEnabledLogs) {
			model.MonitorLogs = types.BoolValue(true)
		}
		if slices.Contains(*data.MonitoringEnabled, fleetapi.AgentPolicyMonitoringEnabledMetrics) {
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
}

func (model agentPolicyModel) toAPICreateModel() fleetapi.AgentPolicyCreateRequest {
	monitoring := make([]fleetapi.AgentPolicyCreateRequestMonitoringEnabled, 0, 2)
	if model.MonitorLogs.ValueBool() {
		monitoring = append(monitoring, fleetapi.AgentPolicyCreateRequestMonitoringEnabledLogs)
	}
	if model.MonitorMetrics.ValueBool() {
		monitoring = append(monitoring, fleetapi.AgentPolicyCreateRequestMonitoringEnabledMetrics)
	}

	body := fleetapi.AgentPolicyCreateRequest{
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

	return body
}

func (model agentPolicyModel) toAPIUpdateModel() fleetapi.AgentPolicyUpdateRequest {
	monitoring := make([]fleetapi.AgentPolicyUpdateRequestMonitoringEnabled, 0, 2)
	if model.MonitorLogs.ValueBool() {
		monitoring = append(monitoring, fleetapi.Logs)
	}
	if model.MonitorMetrics.ValueBool() {
		monitoring = append(monitoring, fleetapi.Metrics)
	}

	body := fleetapi.AgentPolicyUpdateRequest{
		DataOutputId:       model.DataOutputId.ValueStringPointer(),
		Description:        model.Description.ValueStringPointer(),
		DownloadSourceId:   model.DownloadSourceId.ValueStringPointer(),
		FleetServerHostId:  model.FleetServerHostId.ValueStringPointer(),
		MonitoringEnabled:  &monitoring,
		MonitoringOutputId: model.MonitoringOutputId.ValueStringPointer(),
		Name:               model.Name.ValueString(),
		Namespace:          model.Namespace.ValueString(),
	}

	return body
}
