package agent_policy

import (
	"context"
	"fmt"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type globalDataTagsItemModel struct {
	StringValue types.String  `tfsdk:"string_value"`
	NumberValue types.Float32 `tfsdk:"number_value"`
}

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
	UnenrollTimeout    types.Int64  `tfsdk:"unenroll_timeout"`
	GlobalDataTags     types.Map    `tfsdk:"global_data_tags"` //> globalDataTagsModel
}

func (model *agentPolicyModel) populateFromAPI(ctx context.Context, data *kbapi.AgentPolicy) diag.Diagnostics {
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
	if !utils.IsKnown(model.MonitorMetrics) {
		model.MonitorMetrics = types.BoolValue(false)
	}

	model.MonitoringOutputId = types.StringPointerValue(data.MonitoringOutputId)
	model.Name = types.StringValue(data.Name)
	model.Namespace = types.StringValue(data.Namespace)

	if data.UnenrollTimeout != nil {
		model.UnenrollTimeout = types.Int64Value(int64(*data.UnenrollTimeout))
	}

	if utils.Deref(data.GlobalDataTags) != nil {
		diags := diag.Diagnostics{}
		var map0 = make(map[string]globalDataTagsItemModel)
		for _, v := range utils.Deref(data.GlobalDataTags) {
			maybeFloat, error := v.Value.AsAgentPolicyGlobalDataTagsItemValue1()
			if error != nil {
				maybeString, error := v.Value.AsAgentPolicyGlobalDataTagsItemValue0()
				if error != nil {
					diags.AddError("Failed to unmarshal global data tags", error.Error())
				}
				map0[v.Name] = globalDataTagsItemModel{
					StringValue: types.StringValue(maybeString),
				}
			} else {
				map0[v.Name] = globalDataTagsItemModel{
					NumberValue: types.Float32Value(float32(maybeFloat)),
				}
			}
		}

		model.GlobalDataTags = utils.MapValueFrom(ctx, map0, getGlobalDataTagsAttrTypes().(attr.TypeWithElementType).ElementType(), path.Root("global_data_tags"), &diags)
		if diags.HasError() {
			return diags
		}

	}

	return nil
}

// convertGlobalDataTags converts the global data tags from terraform model to API model
// and performs version validation
func (model *agentPolicyModel) convertGlobalDataTags(ctx context.Context, serverVersion *version.Version) (*[]kbapi.AgentPolicyGlobalDataTagsItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(model.GlobalDataTags.Elements()) == 0 {
		if serverVersion.GreaterThanOrEqual(MinVersionGlobalDataTags) {
			emptyList := make([]kbapi.AgentPolicyGlobalDataTagsItem, 0)
			return &emptyList, diags
		}
		return nil, diags
	}

	if serverVersion.LessThan(MinVersionGlobalDataTags) {
		diags.AddError("global_data_tags ES version error", fmt.Sprintf("Global data tags are only supported in Elastic Stack %s and above", MinVersionGlobalDataTags))
		return nil, diags
	}

	items := utils.MapTypeToMap(ctx, model.GlobalDataTags, path.Root("global_data_tags"), &diags,
		func(item globalDataTagsItemModel, meta utils.MapMeta) kbapi.AgentPolicyGlobalDataTagsItem {
			var value kbapi.AgentPolicyGlobalDataTagsItem_Value
			var err error
			if item.StringValue.ValueStringPointer() != nil {
				err = value.FromAgentPolicyGlobalDataTagsItemValue0(*item.StringValue.ValueStringPointer())
			} else {
				err = value.FromAgentPolicyGlobalDataTagsItemValue1(*item.NumberValue.ValueFloat32Pointer())
			}
			if err != nil {
				diags.AddError("global_data_tags validation_error_converting_values", err.Error())
				return kbapi.AgentPolicyGlobalDataTagsItem{}
			}
			return kbapi.AgentPolicyGlobalDataTagsItem{
				Name:  meta.Key,
				Value: value,
			}
		})

	if diags.HasError() {
		return nil, diags
	}

	itemsList := make([]kbapi.AgentPolicyGlobalDataTagsItem, 0, len(items))
	for _, v := range items {
		itemsList = append(itemsList, v)
	}

	return &itemsList, diags
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

	if utils.IsKnown(model.UnenrollTimeout) {
		body.UnenrollTimeout = utils.Pointer(float32(model.UnenrollTimeout.ValueInt64()))
	}

	tags, diags := model.convertGlobalDataTags(ctx, serverVersion)
	if diags.HasError() {
		return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
	}
	body.GlobalDataTags = tags

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

	if utils.IsKnown(model.UnenrollTimeout) {
		body.UnenrollTimeout = utils.Pointer(float32(model.UnenrollTimeout.ValueInt64()))
	}

	tags, diags := model.convertGlobalDataTags(ctx, serverVersion)
	if diags.HasError() {
		return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
	}
	body.GlobalDataTags = tags

	return body, nil
}
