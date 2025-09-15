package agent_policy

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type features struct {
	SupportsGlobalDataTags      bool
	SupportsSupportsAgentless   bool
	SupportsInactivityTimeout   bool
	SupportsUnenrollmentTimeout bool
}

type globalDataTagsItemModel struct {
	StringValue types.String  `tfsdk:"string_value"`
	NumberValue types.Float32 `tfsdk:"number_value"`
}

type agentPolicyModel struct {
	ID                  types.String         `tfsdk:"id"`
	PolicyID            types.String         `tfsdk:"policy_id"`
	Name                types.String         `tfsdk:"name"`
	Namespace           types.String         `tfsdk:"namespace"`
	Description         types.String         `tfsdk:"description"`
	DataOutputId        types.String         `tfsdk:"data_output_id"`
	MonitoringOutputId  types.String         `tfsdk:"monitoring_output_id"`
	FleetServerHostId   types.String         `tfsdk:"fleet_server_host_id"`
	DownloadSourceId    types.String         `tfsdk:"download_source_id"`
	MonitorLogs         types.Bool           `tfsdk:"monitor_logs"`
	MonitorMetrics      types.Bool           `tfsdk:"monitor_metrics"`
	SysMonitoring       types.Bool           `tfsdk:"sys_monitoring"`
	SkipDestroy         types.Bool           `tfsdk:"skip_destroy"`
	SupportsAgentless   types.Bool           `tfsdk:"supports_agentless"`
	InactivityTimeout   customtypes.Duration `tfsdk:"inactivity_timeout"`
	UnenrollmentTimeout customtypes.Duration `tfsdk:"unenrollment_timeout"`
	GlobalDataTags      types.Map            `tfsdk:"global_data_tags"` //> globalDataTagsModel
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
	model.SupportsAgentless = types.BoolPointerValue(data.SupportsAgentless)
	if data.InactivityTimeout != nil {
		// Convert seconds to duration string
		seconds := int64(*data.InactivityTimeout)
		d := time.Duration(seconds) * time.Second
		model.InactivityTimeout = customtypes.NewDurationValue(d.String())
	} else {
		model.InactivityTimeout = customtypes.NewDurationNull()
	}
	if data.UnenrollTimeout != nil {
		// Convert seconds to duration string
		seconds := int64(*data.UnenrollTimeout)
		d := time.Duration(seconds) * time.Second
		model.UnenrollmentTimeout = customtypes.NewDurationValue(d.String())
	} else {
		model.UnenrollmentTimeout = customtypes.NewDurationNull()
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
func (model *agentPolicyModel) convertGlobalDataTags(ctx context.Context, feat features) (*[]kbapi.AgentPolicyGlobalDataTagsItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(model.GlobalDataTags.Elements()) == 0 {
		if feat.SupportsGlobalDataTags {
			emptyList := make([]kbapi.AgentPolicyGlobalDataTagsItem, 0)
			return &emptyList, diags
		}
		return nil, diags
	}

	if !feat.SupportsGlobalDataTags {
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

func (model *agentPolicyModel) toAPICreateModel(ctx context.Context, feat features) (kbapi.PostFleetAgentPoliciesJSONRequestBody, diag.Diagnostics) {
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

	if utils.IsKnown(model.SupportsAgentless) {
		if !feat.SupportsSupportsAgentless {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("supports_agentless"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Supports agentless is only supported in Elastic Stack %s and above", MinSupportsAgentlessVersion),
				),
			}
		}
		body.SupportsAgentless = model.SupportsAgentless.ValueBoolPointer()
	}

	if utils.IsKnown(model.InactivityTimeout) {
		if !feat.SupportsInactivityTimeout {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("inactivity_timeout"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Inactivity timeout is only supported in Elastic Stack %s and above", MinVersionInactivityTimeout),
				),
			}
		}
		duration, diags := model.InactivityTimeout.Parse()
		if diags.HasError() {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}
		seconds := float32(duration.Seconds())
		body.InactivityTimeout = &seconds
	}

	if utils.IsKnown(model.UnenrollmentTimeout) {
		if !feat.SupportsUnenrollmentTimeout {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("unenrollment_timeout"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Unenrollment timeout is only supported in Elastic Stack %s and above", MinVersionUnenrollmentTimeout),
				),
			}
		}
		duration, diags := model.UnenrollmentTimeout.Parse()
		if diags.HasError() {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}
		seconds := float32(duration.Seconds())
		body.UnenrollTimeout = &seconds
	}

	tags, diags := model.convertGlobalDataTags(ctx, feat)
	if diags.HasError() {
		return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
	}
	body.GlobalDataTags = tags

	return body, nil
}

func (model *agentPolicyModel) toAPIUpdateModel(ctx context.Context, feat features) (kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody, diag.Diagnostics) {
	monitoring := make([]kbapi.PutFleetAgentPoliciesAgentpolicyidJSONBodyMonitoringEnabled, 0, 2)
	if model.MonitorLogs.ValueBool() {
		monitoring = append(monitoring, kbapi.PutFleetAgentPoliciesAgentpolicyidJSONBodyMonitoringEnabledLogs)
	}
	if model.MonitorMetrics.ValueBool() {
		monitoring = append(monitoring, kbapi.PutFleetAgentPoliciesAgentpolicyidJSONBodyMonitoringEnabledMetrics)
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

	if utils.IsKnown(model.SupportsAgentless) {
		if !feat.SupportsSupportsAgentless {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("supports_agentless"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Supports agentless is only supported in Elastic Stack %s and above", MinSupportsAgentlessVersion),
				),
			}
		}
		body.SupportsAgentless = model.SupportsAgentless.ValueBoolPointer()
	}

	if utils.IsKnown(model.InactivityTimeout) {
		if !feat.SupportsInactivityTimeout {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("inactivity_timeout"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Inactivity timeout is only supported in Elastic Stack %s and above", MinVersionInactivityTimeout),
				),
			}
		}
		duration, diags := model.InactivityTimeout.Parse()
		if diags.HasError() {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}
		seconds := float32(duration.Seconds())
		body.InactivityTimeout = &seconds
	}

	if utils.IsKnown(model.UnenrollmentTimeout) {
		if !feat.SupportsUnenrollmentTimeout {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("unenrollment_timeout"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Unenrollment timeout is only supported in Elastic Stack %s and above", MinVersionUnenrollmentTimeout),
				),
			}
		}
		duration, diags := model.UnenrollmentTimeout.Parse()
		if diags.HasError() {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}
		seconds := float32(duration.Seconds())
		body.UnenrollTimeout = &seconds
	}

	tags, diags := model.convertGlobalDataTags(ctx, feat)
	if diags.HasError() {
		return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
	}
	body.GlobalDataTags = tags

	return body, nil
}
