package agent_policy

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	// HostNameFormatHostname represents the short hostname format (e.g., "myhost")
	HostNameFormatHostname = "hostname"
	// HostNameFormatFQDN represents the fully qualified domain name format (e.g., "myhost.example.com")
	HostNameFormatFQDN = "fqdn"
	// agentFeatureFQDN is the name of the agent feature that enables FQDN host name format
	agentFeatureFQDN = "fqdn"
)

// apiAgentFeature is the type expected by the generated API for agent features
type apiAgentFeature = struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name"`
}

type features struct {
	SupportsGlobalDataTags      bool
	SupportsSupportsAgentless   bool
	SupportsInactivityTimeout   bool
	SupportsUnenrollmentTimeout bool
	SupportsSpaceIds            bool
	SupportsRequiredVersions    bool
	SupportsAgentFeatures       bool
}

type globalDataTagsItemModel struct {
	StringValue types.String  `tfsdk:"string_value"`
	NumberValue types.Float32 `tfsdk:"number_value"`
}

type advancedSettingsModel struct {
	LoggingLevel                  types.String         `tfsdk:"logging_level"`
	LoggingToFiles                types.Bool           `tfsdk:"logging_to_files"`
	LoggingFilesInterval          customtypes.Duration `tfsdk:"logging_files_interval"`
	LoggingFilesKeepfiles         types.Int32          `tfsdk:"logging_files_keepfiles"`
	LoggingFilesRotateeverybytes  types.Int64          `tfsdk:"logging_files_rotateeverybytes"`
	LoggingMetricsPeriod          customtypes.Duration `tfsdk:"logging_metrics_period"`
	GoMaxProcs                    types.Int32          `tfsdk:"go_max_procs"`
	DownloadTimeout               customtypes.Duration `tfsdk:"download_timeout"`
	DownloadTargetDirectory       types.String         `tfsdk:"download_target_directory"`
	MonitoringRuntimeExperimental types.Bool           `tfsdk:"monitoring_runtime_experimental"`
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
	HostNameFormat      types.String         `tfsdk:"host_name_format"`
	SupportsAgentless   types.Bool           `tfsdk:"supports_agentless"`
	InactivityTimeout   customtypes.Duration `tfsdk:"inactivity_timeout"`
	UnenrollmentTimeout customtypes.Duration `tfsdk:"unenrollment_timeout"`
	GlobalDataTags      types.Map            `tfsdk:"global_data_tags"` //> globalDataTagsModel
	SpaceIds            types.Set            `tfsdk:"space_ids"`
	RequiredVersions    types.Map            `tfsdk:"required_versions"`
	AdvancedSettings    types.Object         `tfsdk:"advanced_settings"`
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

	// Determine host_name_format from AgentFeatures
	// If AgentFeatures contains {"enabled": true, "name": "fqdn"}, then host_name_format is "fqdn"
	// Otherwise, it defaults to "hostname"
	model.HostNameFormat = types.StringValue(HostNameFormatHostname)
	if data.AgentFeatures != nil {
		for _, feature := range *data.AgentFeatures {
			if feature.Name == agentFeatureFQDN && feature.Enabled {
				model.HostNameFormat = types.StringValue(HostNameFormatFQDN)
				break
			}
		}
	}

	if data.InactivityTimeout != nil {
		// Convert seconds to duration string
		seconds := int64(*data.InactivityTimeout)
		d := time.Duration(seconds) * time.Second
		model.InactivityTimeout = customtypes.NewDurationValue(d.Truncate(time.Second).String())
	} else {
		model.InactivityTimeout = customtypes.NewDurationNull()
	}
	if data.UnenrollTimeout != nil {
		// Convert seconds to duration string
		seconds := int64(*data.UnenrollTimeout)
		d := time.Duration(seconds) * time.Second
		model.UnenrollmentTimeout = customtypes.NewDurationValue(d.Truncate(time.Second).String())
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

	if data.SpaceIds != nil {
		spaceIds, d := types.SetValueFrom(ctx, types.StringType, *data.SpaceIds)
		if d.HasError() {
			return d
		}
		model.SpaceIds = spaceIds
	} else {
		model.SpaceIds = types.SetNull(types.StringType)
	}

	// Handle required_versions
	if data.RequiredVersions != nil {
		versionMap := make(map[string]attr.Value)

		for _, rv := range *data.RequiredVersions {
			// Round the float32 percentage to nearest integer since we use Int32 in the schema
			percentage := int32(math.Round(float64(rv.Percentage)))
			versionMap[rv.Version] = types.Int32Value(percentage)
		}

		reqVersions, d := types.MapValue(types.Int32Type, versionMap)
		if d.HasError() {
			return d
		}
		model.RequiredVersions = reqVersions
	} else {
		model.RequiredVersions = types.MapNull(types.Int32Type)
	}

	// Handle advanced_settings
	if diags := model.populateAdvancedSettingsFromAPI(ctx, data); diags.HasError() {
		return diags
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

// convertRequiredVersions converts the required versions from terraform model to API model
func (model *agentPolicyModel) convertRequiredVersions(feat features) (*[]struct {
	Percentage float32 `json:"percentage"`
	Version    string  `json:"version"`
}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if model.RequiredVersions.IsNull() || model.RequiredVersions.IsUnknown() {
		return nil, diags
	}

	// Check if required_versions is supported
	if !feat.SupportsRequiredVersions {
		return nil, diag.Diagnostics{
			diag.NewAttributeErrorDiagnostic(
				path.Root("required_versions"),
				"Unsupported Elasticsearch version",
				fmt.Sprintf("Required versions (automatic agent upgrades) are only supported in Elastic Stack %s and above", MinVersionRequiredVersions),
			),
		}
	}

	elements := model.RequiredVersions.Elements()

	// If the map is empty (required_versions = {}), return an empty array to clear upgrades
	if len(elements) == 0 {
		emptyArray := make([]struct {
			Percentage float32 `json:"percentage"`
			Version    string  `json:"version"`
		}, 0)
		return &emptyArray, diags
	}

	result := make([]struct {
		Percentage float32 `json:"percentage"`
		Version    string  `json:"version"`
	}, 0, len(elements))

	for version, percentageVal := range elements {
		percentageInt32, ok := percentageVal.(types.Int32)
		if !ok {
			diags.AddError("required_versions conversion error", fmt.Sprintf("Expected Int32 value, got %T", percentageVal))
			continue
		}

		if percentageInt32.IsNull() || percentageInt32.IsUnknown() {
			diags.AddError("required_versions validation error", "percentage cannot be null or unknown")
			continue
		}

		result = append(result, struct {
			Percentage float32 `json:"percentage"`
			Version    string  `json:"version"`
		}{
			Percentage: float32(percentageInt32.ValueInt32()),
			Version:    version,
		})
	}

	if diags.HasError() {
		return nil, diags
	}

	return &result, diags
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

	if utils.IsKnown(model.SpaceIds) {
		if !feat.SupportsSpaceIds {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("space_ids"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Space IDs are only supported in Elastic Stack %s and above", MinVersionSpaceIds),
				),
			}
		}
		var spaceIds []string
		d := model.SpaceIds.ElementsAs(ctx, &spaceIds, false)
		diags.Append(d...)
		if diags.HasError() {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}
		body.SpaceIds = &spaceIds
	}

	// Handle required_versions
	requiredVersions, d := model.convertRequiredVersions(feat)
	if d.HasError() {
		return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, d
	}
	body.RequiredVersions = requiredVersions

	// Handle host_name_format via AgentFeatures
	if agentFeature := model.convertHostNameFormatToAgentFeature(); agentFeature != nil {
		if !feat.SupportsAgentFeatures {
			// Only error if user explicitly requests FQDN on unsupported version
			// Default "hostname" is fine - just don't send agent_features
			if agentFeature.Enabled {
				return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("host_name_format"),
						"Unsupported Elasticsearch version",
						fmt.Sprintf("host_name_format (agent_features) is only supported in Elastic Stack %s and above", MinVersionAgentFeatures),
					),
				}
			}
			// On unsupported version with default "hostname", don't send agent_features
		} else {
			body.AgentFeatures = &[]apiAgentFeature{*agentFeature}
		}
	}

	// Handle advanced_settings
	body.AdvancedSettings = model.convertAdvancedSettingsToAPI(ctx)

	return body, nil
}

func (model *agentPolicyModel) toAPIUpdateModel(ctx context.Context, feat features, existingFeatures []apiAgentFeature) (kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody, diag.Diagnostics) {
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

	if utils.IsKnown(model.SpaceIds) {
		if !feat.SupportsSpaceIds {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("space_ids"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Space IDs are only supported in Elastic Stack %s and above", MinVersionSpaceIds),
				),
			}
		}
		var spaceIds []string
		d := model.SpaceIds.ElementsAs(ctx, &spaceIds, false)
		diags.Append(d...)
		if diags.HasError() {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}
		body.SpaceIds = &spaceIds
	}

	// Handle required_versions
	requiredVersions, d := model.convertRequiredVersions(feat)
	if d.HasError() {
		return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, d
	}
	body.RequiredVersions = requiredVersions

	// Handle host_name_format via AgentFeatures, preserving other existing features
	if agentFeature := model.convertHostNameFormatToAgentFeature(); agentFeature != nil {
		if !feat.SupportsAgentFeatures {
			// Only error if user explicitly requests FQDN on unsupported version
			// Default "hostname" is fine - just don't send agent_features
			if agentFeature.Enabled {
				return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("host_name_format"),
						"Unsupported Elasticsearch version",
						fmt.Sprintf("host_name_format (agent_features) is only supported in Elastic Stack %s and above", MinVersionAgentFeatures),
					),
				}
			}
			// On unsupported version with default "hostname", don't send agent_features
		} else {
			body.AgentFeatures = mergeAgentFeature(existingFeatures, agentFeature)
		}
	} else if feat.SupportsAgentFeatures && len(existingFeatures) > 0 {
		// Preserve existing features even when host_name_format is not set
		body.AgentFeatures = &existingFeatures
	}

	// Handle advanced_settings
	body.AdvancedSettings = model.convertAdvancedSettingsToAPI(ctx)

	return body, nil
}

// convertHostNameFormatToAgentFeature converts the host_name_format field to a single AgentFeature.
// - When host_name_format is "fqdn": returns {"name": "fqdn", "enabled": true}
// - When host_name_format is "hostname": returns {"name": "fqdn", "enabled": false} to explicitly disable
// - When not set: returns nil (no change to existing features)
func (model *agentPolicyModel) convertHostNameFormatToAgentFeature() *apiAgentFeature {
	// If host_name_format is not set or unknown, don't modify AgentFeatures
	if model.HostNameFormat.IsNull() || model.HostNameFormat.IsUnknown() {
		return nil
	}

	// Explicitly set enabled based on the host_name_format value
	// We need to send enabled: false when hostname is selected to override any existing fqdn setting
	return &apiAgentFeature{
		Enabled: model.HostNameFormat.ValueString() == HostNameFormatFQDN,
		Name:    agentFeatureFQDN,
	}
}

// mergeAgentFeature merges a single feature into existing features, replacing any feature with the same name.
// If newFeature is nil, returns existing features unchanged (nil if existing is empty).
func mergeAgentFeature(existing []apiAgentFeature, newFeature *apiAgentFeature) *[]apiAgentFeature {
	if newFeature == nil {
		if len(existing) == 0 {
			return nil
		}
		return &existing
	}

	// Check if the feature already exists and replace it, otherwise append
	result := make([]apiAgentFeature, 0, len(existing)+1)
	found := false

	for _, f := range existing {
		if f.Name == newFeature.Name {
			result = append(result, *newFeature)
			found = true
		} else {
			result = append(result, f)
		}
	}

	if !found {
		result = append(result, *newFeature)
	}

	return &result
}

// advancedSettingsAttrTypes returns attribute types for advanced_settings pulled from the schema
func advancedSettingsAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["advanced_settings"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// populateAdvancedSettingsFromAPI populates the advanced settings from API response
func (model *agentPolicyModel) populateAdvancedSettingsFromAPI(ctx context.Context, data *kbapi.AgentPolicy) diag.Diagnostics {
	if data.AdvancedSettings == nil {
		model.AdvancedSettings = types.ObjectNull(advancedSettingsAttrTypes())
		return nil
	}

	settings := advancedSettingsModel{}

	// Logging level
	if data.AdvancedSettings.AgentLoggingLevel != nil {
		if str, ok := data.AdvancedSettings.AgentLoggingLevel.(string); ok {
			settings.LoggingLevel = types.StringValue(str)
		} else {
			settings.LoggingLevel = types.StringNull()
		}
	} else {
		settings.LoggingLevel = types.StringNull()
	}

	// Logging to files
	if data.AdvancedSettings.AgentLoggingToFiles != nil {
		if b, ok := data.AdvancedSettings.AgentLoggingToFiles.(bool); ok {
			settings.LoggingToFiles = types.BoolValue(b)
		} else {
			settings.LoggingToFiles = types.BoolNull()
		}
	} else {
		settings.LoggingToFiles = types.BoolNull()
	}

	// Logging files interval
	if data.AdvancedSettings.AgentLoggingFilesInterval != nil {
		if str, ok := data.AdvancedSettings.AgentLoggingFilesInterval.(string); ok {
			settings.LoggingFilesInterval = customtypes.NewDurationValue(str)
		} else {
			settings.LoggingFilesInterval = customtypes.NewDurationNull()
		}
	} else {
		settings.LoggingFilesInterval = customtypes.NewDurationNull()
	}

	// Logging files keepfiles
	if data.AdvancedSettings.AgentLoggingFilesKeepfiles != nil {
		if f, ok := data.AdvancedSettings.AgentLoggingFilesKeepfiles.(float64); ok {
			settings.LoggingFilesKeepfiles = types.Int32Value(int32(f))
		} else {
			settings.LoggingFilesKeepfiles = types.Int32Null()
		}
	} else {
		settings.LoggingFilesKeepfiles = types.Int32Null()
	}

	// Logging files rotateeverybytes
	if data.AdvancedSettings.AgentLoggingFilesRotateeverybytes != nil {
		if f, ok := data.AdvancedSettings.AgentLoggingFilesRotateeverybytes.(float64); ok {
			settings.LoggingFilesRotateeverybytes = types.Int64Value(int64(f))
		} else {
			settings.LoggingFilesRotateeverybytes = types.Int64Null()
		}
	} else {
		settings.LoggingFilesRotateeverybytes = types.Int64Null()
	}

	// Logging metrics period
	if data.AdvancedSettings.AgentLoggingMetricsPeriod != nil {
		if str, ok := data.AdvancedSettings.AgentLoggingMetricsPeriod.(string); ok {
			settings.LoggingMetricsPeriod = customtypes.NewDurationValue(str)
		} else {
			settings.LoggingMetricsPeriod = customtypes.NewDurationNull()
		}
	} else {
		settings.LoggingMetricsPeriod = customtypes.NewDurationNull()
	}

	// Go max procs
	if data.AdvancedSettings.AgentLimitsGoMaxProcs != nil {
		if f, ok := data.AdvancedSettings.AgentLimitsGoMaxProcs.(float64); ok {
			settings.GoMaxProcs = types.Int32Value(int32(f))
		} else {
			settings.GoMaxProcs = types.Int32Null()
		}
	} else {
		settings.GoMaxProcs = types.Int32Null()
	}

	// Download timeout
	if data.AdvancedSettings.AgentDownloadTimeout != nil {
		if str, ok := data.AdvancedSettings.AgentDownloadTimeout.(string); ok {
			settings.DownloadTimeout = customtypes.NewDurationValue(str)
		} else {
			settings.DownloadTimeout = customtypes.NewDurationNull()
		}
	} else {
		settings.DownloadTimeout = customtypes.NewDurationNull()
	}

	// Download target directory
	if data.AdvancedSettings.AgentDownloadTargetDirectory != nil {
		if str, ok := data.AdvancedSettings.AgentDownloadTargetDirectory.(string); ok {
			settings.DownloadTargetDirectory = types.StringValue(str)
		} else {
			settings.DownloadTargetDirectory = types.StringNull()
		}
	} else {
		settings.DownloadTargetDirectory = types.StringNull()
	}

	// Monitoring runtime experimental
	if data.AdvancedSettings.AgentMonitoringRuntimeExperimental != nil {
		if b, ok := data.AdvancedSettings.AgentMonitoringRuntimeExperimental.(bool); ok {
			settings.MonitoringRuntimeExperimental = types.BoolValue(b)
		} else {
			settings.MonitoringRuntimeExperimental = types.BoolNull()
		}
	} else {
		settings.MonitoringRuntimeExperimental = types.BoolNull()
	}

	obj, diags := types.ObjectValueFrom(ctx, advancedSettingsAttrTypes(), settings)
	if diags.HasError() {
		return diags
	}
	model.AdvancedSettings = obj
	return nil
}

// advancedSettingsAPIResult is the return type for convertAdvancedSettingsToAPI
type advancedSettingsAPIResult = struct {
	AgentDownloadTargetDirectory       interface{} `json:"agent_download_target_directory,omitempty"`
	AgentDownloadTimeout               interface{} `json:"agent_download_timeout,omitempty"`
	AgentLimitsGoMaxProcs              interface{} `json:"agent_limits_go_max_procs,omitempty"`
	AgentLoggingFilesInterval          interface{} `json:"agent_logging_files_interval,omitempty"`
	AgentLoggingFilesKeepfiles         interface{} `json:"agent_logging_files_keepfiles,omitempty"`
	AgentLoggingFilesRotateeverybytes  interface{} `json:"agent_logging_files_rotateeverybytes,omitempty"`
	AgentLoggingLevel                  interface{} `json:"agent_logging_level,omitempty"`
	AgentLoggingMetricsPeriod          interface{} `json:"agent_logging_metrics_period,omitempty"`
	AgentLoggingToFiles                interface{} `json:"agent_logging_to_files,omitempty"`
	AgentMonitoringRuntimeExperimental interface{} `json:"agent_monitoring_runtime_experimental,omitempty"`
}

// convertAdvancedSettingsToAPI converts the advanced settings config to API format
func (model *agentPolicyModel) convertAdvancedSettingsToAPI(ctx context.Context) *advancedSettingsAPIResult {
	if !utils.IsKnown(model.AdvancedSettings) {
		return nil
	}

	var settings advancedSettingsModel
	model.AdvancedSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})

	// Check if any values are set
	hasValues := utils.IsKnown(settings.LoggingLevel) ||
		utils.IsKnown(settings.LoggingToFiles) ||
		utils.IsKnown(settings.LoggingFilesInterval) ||
		utils.IsKnown(settings.LoggingFilesKeepfiles) ||
		utils.IsKnown(settings.LoggingFilesRotateeverybytes) ||
		utils.IsKnown(settings.LoggingMetricsPeriod) ||
		utils.IsKnown(settings.GoMaxProcs) ||
		utils.IsKnown(settings.DownloadTimeout) ||
		utils.IsKnown(settings.DownloadTargetDirectory) ||
		utils.IsKnown(settings.MonitoringRuntimeExperimental)

	if !hasValues {
		return nil
	}

	result := &advancedSettingsAPIResult{}

	if utils.IsKnown(settings.LoggingLevel) {
		result.AgentLoggingLevel = settings.LoggingLevel.ValueString()
	}
	if utils.IsKnown(settings.LoggingToFiles) {
		result.AgentLoggingToFiles = settings.LoggingToFiles.ValueBool()
	}
	if utils.IsKnown(settings.LoggingFilesInterval) {
		result.AgentLoggingFilesInterval = settings.LoggingFilesInterval.ValueString()
	}
	if utils.IsKnown(settings.LoggingFilesKeepfiles) {
		result.AgentLoggingFilesKeepfiles = settings.LoggingFilesKeepfiles.ValueInt32()
	}
	if utils.IsKnown(settings.LoggingFilesRotateeverybytes) {
		result.AgentLoggingFilesRotateeverybytes = settings.LoggingFilesRotateeverybytes.ValueInt64()
	}
	if utils.IsKnown(settings.LoggingMetricsPeriod) {
		result.AgentLoggingMetricsPeriod = settings.LoggingMetricsPeriod.ValueString()
	}
	if utils.IsKnown(settings.GoMaxProcs) {
		result.AgentLimitsGoMaxProcs = settings.GoMaxProcs.ValueInt32()
	}
	if utils.IsKnown(settings.DownloadTimeout) {
		result.AgentDownloadTimeout = settings.DownloadTimeout.ValueString()
	}
	if utils.IsKnown(settings.DownloadTargetDirectory) {
		result.AgentDownloadTargetDirectory = settings.DownloadTargetDirectory.ValueString()
	}
	if utils.IsKnown(settings.MonitoringRuntimeExperimental) {
		result.AgentMonitoringRuntimeExperimental = settings.MonitoringRuntimeExperimental.ValueBool()
	}

	return result
}
