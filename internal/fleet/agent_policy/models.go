package agent_policy

import (
	"context"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	GlobalDataTags     types.Map    `tfsdk:"global_data_tags"`
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
	if utils.Deref(data.GlobalDataTags) != nil {
		diags := diag.Diagnostics{}
		var map0 = make(map[string]any)
		for _, v := range utils.Deref(data.GlobalDataTags) {
			maybeFloat, error := v.Value.AsAgentPolicyGlobalDataTagsItemValue1()
			if error != nil {
				maybeString, error := v.Value.AsAgentPolicyGlobalDataTagsItemValue0()
				if error != nil {
					diags.AddError("Failed to unmarshal global data tags", error.Error())
				}
				map0[v.Name] = map[string]string{
					"string_value": string(maybeString),
				}
			} else {
				map0[v.Name] = map[string]float32{
					"number_value": float32(maybeFloat),
				}
			}
		}
		gdt := utils.MapValueFrom(ctx, map0, getGlobalDataTagsAttrType(), path.Root("global_data_tags"), &diags)
		if diags.HasError() {
			return diags
		}
		model.GlobalDataTags = gdt
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

	if len(model.GlobalDataTags.Elements()) > 0 {
		var diags diag.Diagnostics
		if serverVersion.LessThan(MinVersionGlobalDataTags) {
			diags.AddError("global_data_tags ES version error", "Global data tags are only supported in Elastic Stack 8.15.0 and above")
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}

		var items []kbapi.AgentPolicyGlobalDataTagsItem
		itemsMap := utils.MapTypeAs[struct {
			string_value *string
			number_value *float32
		}](ctx, model.GlobalDataTags, path.Root("global_data_tags"), &diags)
		if diags.HasError() {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}
		for k, v := range itemsMap {
			if (v.string_value != nil && v.number_value != nil) || (v.string_value == nil && v.number_value == nil) {
				diags.AddError("global_data_tags ES version error", "Global data tags must have exactly one of string_value or number_value")
				return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
			}
			var value kbapi.AgentPolicyGlobalDataTagsItem_Value
			var err error
			if v.string_value != nil {
				err = value.FromAgentPolicyGlobalDataTagsItemValue0(*v.string_value)
			} else {
				err = value.FromAgentPolicyGlobalDataTagsItemValue1(*v.number_value)
			}
			if err != nil {
				diags.AddError("global_data_tags ES version error", "could not convert global data tags value")
				return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
			}
			items = append(items, kbapi.AgentPolicyGlobalDataTagsItem{
				Name:  k,
				Value: value,
			})
		}

		body.GlobalDataTags = &items
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

	if len(model.GlobalDataTags.Elements()) > 0 {
		var diags diag.Diagnostics
		if serverVersion.LessThan(MinVersionGlobalDataTags) {
			diags.AddError("global_data_tags ES version error", "Global data tags are only supported in Elastic Stack 8.15.0 and above")
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}

		var items []kbapi.AgentPolicyGlobalDataTagsItem
		itemsMap := utils.MapTypeAs[struct {
			string_value *string
			number_value *float32
		}](ctx, model.GlobalDataTags, path.Root("global_data_tags"), &diags)
		if diags.HasError() {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}
		for k, v := range itemsMap {
			if (v.string_value != nil && v.number_value != nil) || (v.string_value == nil && v.number_value == nil) {
				diags.AddError("global_data_tags ES version error", "Global data tags must have exactly one of string_value or number_value")
				return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
			}
			var value kbapi.AgentPolicyGlobalDataTagsItem_Value
			var err error
			if v.string_value != nil {
				// s := *v.string_value
				err = value.FromAgentPolicyGlobalDataTagsItemValue0(*v.string_value)
			} else {
				err = value.FromAgentPolicyGlobalDataTagsItemValue1(*v.number_value)
			}
			if err != nil {
				diags.AddError("global_data_tags ES version error", "could not convert global data tags value")
				return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
			}
			items = append(items, kbapi.AgentPolicyGlobalDataTagsItem{
				Name:  k,
				Value: value,
			})
		}

		body.GlobalDataTags = &items
	}

	return body, nil
}
