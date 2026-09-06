// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package agentpolicy

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/globaldatatags"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func agentFeaturesFromPolicy(p *kbapi.KibanaHTTPAPIsAgentPolicyResponse) []apiAgentFeature {
	if p == nil || p.AgentFeatures == nil {
		return nil
	}
	out := make([]apiAgentFeature, len(*p.AgentFeatures))
	for i, f := range *p.AgentFeatures {
		out[i] = apiAgentFeature{Enabled: f.Enabled, Name: f.Name}
	}
	return out
}

type agentPolicyModel struct {
	ID                        types.String         `tfsdk:"id"`
	KibanaConnection          types.List           `tfsdk:"kibana_connection"`
	PolicyID                  types.String         `tfsdk:"policy_id"`
	Name                      types.String         `tfsdk:"name"`
	Namespace                 types.String         `tfsdk:"namespace"`
	Description               types.String         `tfsdk:"description"`
	DataOutputID              types.String         `tfsdk:"data_output_id"`
	MonitoringOutputID        types.String         `tfsdk:"monitoring_output_id"`
	FleetServerHostID         types.String         `tfsdk:"fleet_server_host_id"`
	DownloadSourceID          types.String         `tfsdk:"download_source_id"`
	MonitorLogs               types.Bool           `tfsdk:"monitor_logs"`
	MonitorMetrics            types.Bool           `tfsdk:"monitor_metrics"`
	IsProtected               types.Bool           `tfsdk:"is_protected"`
	SysMonitoring             types.Bool           `tfsdk:"sys_monitoring"`
	SkipDestroy               types.Bool           `tfsdk:"skip_destroy"`
	HostNameFormat            types.String         `tfsdk:"host_name_format"`
	SupportsAgentless         types.Bool           `tfsdk:"supports_agentless"`
	InactivityTimeout         customtypes.Duration `tfsdk:"inactivity_timeout"`
	UnenrollmentTimeout       customtypes.Duration `tfsdk:"unenrollment_timeout"`
	GlobalDataTags            types.Map            `tfsdk:"global_data_tags"` // > globaldatatags.Item
	SpaceIDs                  types.Set            `tfsdk:"space_ids"`
	RequiredVersions          types.Map            `tfsdk:"required_versions"`
	AdvancedMonitoringOptions types.Object         `tfsdk:"advanced_monitoring_options"`
	AdvancedSettings          types.Object         `tfsdk:"advanced_settings"`
}

func (model *agentPolicyModel) populateFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	if data == nil {
		return nil
	}

	model.ID = types.StringValue(data.Id)
	model.PolicyID = types.StringValue(data.Id)
	// The Fleet update API treats omitted optional *string fields as
	// "preserve existing value". When the user removes one of these fields
	// from configuration, the provider sends nil (omitted via omitempty) and
	// the API response continues to report the previous server-side value.
	// Writing that value into state would conflict with the planned null and
	// trigger "Provider produced inconsistent result after apply". Preserve
	// the configured null in state instead; the Fleet-side value is
	// intentionally retained.
	preserveNullStr := func(current types.String, apiVal *string) types.String {
		if current.IsNull() && apiVal != nil && *apiVal != "" {
			return types.StringNull()
		}
		return types.StringPointerValue(apiVal)
	}
	model.DataOutputID = preserveNullStr(model.DataOutputID, data.DataOutputId)
	// The Fleet API normalizes an omitted description to an empty string in
	// its response body. When the user's plan omits description (null), that
	// empty string would be written to state and trigger "Provider produced
	// inconsistent result after apply: was null, but now cty.StringVal("")".
	// Treat an API empty-string as equivalent to null when the configured
	// value is already null. Kibana treats null and "" as equivalent for
	// this field. See https://github.com/elastic/terraform-provider-elasticstack/issues/993.
	apiEmpty := data.Description != nil && *data.Description == ""
	if apiEmpty && model.Description.IsNull() {
		model.Description = types.StringNull()
	} else {
		model.Description = types.StringPointerValue(data.Description)
	}
	model.DownloadSourceID = preserveNullStr(model.DownloadSourceID, data.DownloadSourceId)
	model.FleetServerHostID = preserveNullStr(model.FleetServerHostID, data.FleetServerHostId)

	if data.MonitoringEnabled != nil {
		if slices.Contains(*data.MonitoringEnabled, kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabledLogs) {
			model.MonitorLogs = types.BoolValue(true)
		}
		if slices.Contains(*data.MonitoringEnabled, kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabledMetrics) {
			model.MonitorMetrics = types.BoolValue(true)
		}
	}
	if !typeutils.IsKnown(model.MonitorLogs) {
		model.MonitorLogs = types.BoolValue(false)
	}
	if !typeutils.IsKnown(model.MonitorMetrics) {
		model.MonitorMetrics = types.BoolValue(false)
	}

	model.IsProtected = types.BoolValue(data.IsProtected)

	model.MonitoringOutputID = preserveNullStr(model.MonitoringOutputID, data.MonitoringOutputId)
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
	if tags := typeutils.Deref(data.GlobalDataTags); tags != nil {
		diags := diag.Diagnostics{}

		items := make([]globaldatatags.Tag[kbapi.AgentPolicyGlobalDataTagsItem_Value], len(tags))
		for i, t := range tags {
			items[i] = globaldatatags.Tag[kbapi.AgentPolicyGlobalDataTagsItem_Value]{Name: t.Name, Value: t.Value}
		}

		model.GlobalDataTags = globaldatatags.ToModel(
			ctx,
			items,
			path.Root("global_data_tags"),
			&diags,
			kbapi.AgentPolicyGlobalDataTagsItem_Value.AsAgentPolicyGlobalDataTagsItemValue1,
			kbapi.AgentPolicyGlobalDataTagsItem_Value.AsAgentPolicyGlobalDataTagsItemValue0,
		)
		if diags.HasError() {
			return diags
		}
	}

	spaceIDs, d := typeutils.SetFromAPIStringsPreserveKnownEmpty(ctx, data.SpaceIds, model.SpaceIDs)
	if d.HasError() {
		return d
	}
	model.SpaceIDs = spaceIDs

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

	// Handle advanced monitoring options
	if diags := model.populateAdvancedMonitoringFromAPI(ctx, data); diags.HasError() {
		return diags
	}

	return nil
}

// convertGlobalDataTags converts the global data tags from the Terraform model
// to the API model after version requirements have been enforced.
func (model *agentPolicyModel) convertGlobalDataTags(ctx context.Context, feat agentPolicyFeatures) (*[]kbapi.AgentPolicyGlobalDataTagsItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(model.GlobalDataTags.Elements()) == 0 {
		if feat.SupportsGlobalDataTags {
			emptyList := make([]kbapi.AgentPolicyGlobalDataTagsItem, 0)
			return &emptyList, diags
		}
		return nil, diags
	}

	if !feat.SupportsGlobalDataTags {
		return nil, diags
	}

	tags := globaldatatags.FromModel(ctx, model.GlobalDataTags, path.Root("global_data_tags"), &diags,
		func(s string) (kbapi.AgentPolicyGlobalDataTagsItem_Value, error) {
			var v kbapi.AgentPolicyGlobalDataTagsItem_Value
			return v, v.FromAgentPolicyGlobalDataTagsItemValue0(s)
		},
		func(n float32) (kbapi.AgentPolicyGlobalDataTagsItem_Value, error) {
			var v kbapi.AgentPolicyGlobalDataTagsItem_Value
			return v, v.FromAgentPolicyGlobalDataTagsItemValue1(n)
		},
	)
	if diags.HasError() {
		return nil, diags
	}

	itemsList := make([]kbapi.AgentPolicyGlobalDataTagsItem, len(tags))
	for i, t := range tags {
		itemsList[i] = kbapi.AgentPolicyGlobalDataTagsItem{Name: t.Name, Value: t.Value}
	}
	return &itemsList, diags
}

// convertRequiredVersions converts the required versions from terraform model to API model
func (model *agentPolicyModel) convertRequiredVersions(feat agentPolicyFeatures) (*[]struct {
	Percentage float32 `json:"percentage"`
	Version    string  `json:"version"`
}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(model.RequiredVersions) {
		return nil, diags
	}

	// Omit the field when the connected API does not support it. A configured
	// value has already been rejected by GetVersionRequirements.
	if !feat.SupportsRequiredVersions {
		return nil, diags
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

		if !typeutils.IsKnown(percentageInt32) {
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

func (model *agentPolicyModel) toAPICreateModel(ctx context.Context, feat agentPolicyFeatures) (kbapi.PostFleetAgentPoliciesJSONRequestBody, diag.Diagnostics) {
	monitoring := make([]kbapi.KibanaHTTPAPIsNewAgentPolicyMonitoringEnabled, 0, 2)

	if model.MonitorLogs.ValueBool() {
		monitoring = append(monitoring, kbapi.KibanaHTTPAPIsNewAgentPolicyMonitoringEnabledLogs)
	}
	if model.MonitorMetrics.ValueBool() {
		monitoring = append(monitoring, kbapi.KibanaHTTPAPIsNewAgentPolicyMonitoringEnabledMetrics)
	}

	body := kbapi.PostFleetAgentPoliciesJSONRequestBody{
		DataOutputId:       model.DataOutputID.ValueStringPointer(),
		Description:        model.Description.ValueStringPointer(),
		DownloadSourceId:   model.DownloadSourceID.ValueStringPointer(),
		FleetServerHostId:  model.FleetServerHostID.ValueStringPointer(),
		Id:                 typeutils.OptionalString(model.PolicyID),
		MonitoringEnabled:  &monitoring,
		MonitoringOutputId: model.MonitoringOutputID.ValueStringPointer(),
		Name:               model.Name.ValueString(),
		Namespace:          model.Namespace.ValueString(),
	}

	gated, diags := model.computeFeatureGatedFields(ctx, feat)
	if diags.HasError() {
		return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
	}
	body.IsProtected = gated.isProtected
	body.SupportsAgentless = gated.supportsAgentless
	body.InactivityTimeout = gated.inactivityTimeout
	body.UnenrollTimeout = gated.unenrollTimeout
	body.SpaceIds = gated.spaceIDs

	tags, diags := model.convertGlobalDataTags(ctx, feat)
	if diags.HasError() {
		return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
	}
	body.GlobalDataTags = tags

	// Handle required_versions
	requiredVersions, d := model.convertRequiredVersions(feat)
	if d.HasError() {
		return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, d
	}
	body.RequiredVersions = requiredVersions

	// Handle host_name_format via AgentFeatures
	if agentFeature := model.convertHostNameFormatToAgentFeature(); agentFeature != nil {
		if feat.SupportsAgentFeatures {
			body.AgentFeatures = &[]apiAgentFeature{*agentFeature}
		}
	}

	// Handle advanced_settings
	if typeutils.IsKnown(model.AdvancedSettings) && feat.SupportsAdvancedSettings {
		advancedSettings, diags := model.convertAdvancedSettingsToAPI(ctx, feat)
		if diags.HasError() {
			return kbapi.PostFleetAgentPoliciesJSONRequestBody{}, diags
		}
		body.AdvancedSettings = advancedSettings
	}

	// Handle advanced monitoring options
	if typeutils.IsKnown(model.AdvancedMonitoringOptions) && feat.SupportsAdvancedMonitoring {
		monitoringHTTP, pprofEnabled := model.convertHTTPMonitoringEndpointToAPI(ctx)
		body.MonitoringHttp = monitoringHTTP
		body.MonitoringPprofEnabled = pprofEnabled
		body.MonitoringDiagnostics = model.convertDiagnosticsToAPI(ctx)
	}

	return body, nil
}

func (model *agentPolicyModel) toAPIUpdateModel(
	ctx context.Context,
	feat agentPolicyFeatures,
	existingFeatures []apiAgentFeature,
) (kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody, diag.Diagnostics) {
	monitoring := make([]kbapi.KibanaHTTPAPIsUpdateAgentPolicyRequestBodyMonitoringEnabled, 0, 2)
	if model.MonitorLogs.ValueBool() {
		monitoring = append(monitoring, kbapi.KibanaHTTPAPIsUpdateAgentPolicyRequestBodyMonitoringEnabledLogs)
	}
	if model.MonitorMetrics.ValueBool() {
		monitoring = append(monitoring, kbapi.KibanaHTTPAPIsUpdateAgentPolicyRequestBodyMonitoringEnabledMetrics)
	}

	body := kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{
		DataOutputId:       model.DataOutputID.ValueStringPointer(),
		Description:        model.Description.ValueStringPointer(),
		DownloadSourceId:   model.DownloadSourceID.ValueStringPointer(),
		FleetServerHostId:  model.FleetServerHostID.ValueStringPointer(),
		MonitoringEnabled:  &monitoring,
		MonitoringOutputId: model.MonitoringOutputID.ValueStringPointer(),
		Name:               model.Name.ValueString(),
		Namespace:          model.Namespace.ValueString(),
	}

	gated, diags := model.computeFeatureGatedFields(ctx, feat)
	if diags.HasError() {
		return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
	}
	body.IsProtected = gated.isProtected
	body.SupportsAgentless = gated.supportsAgentless
	body.InactivityTimeout = gated.inactivityTimeout
	body.UnenrollTimeout = gated.unenrollTimeout
	body.SpaceIds = gated.spaceIDs

	tags, diags := model.convertGlobalDataTags(ctx, feat)
	if diags.HasError() {
		return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
	}
	if tags != nil {
		updateTags := make([]struct {
			Name  string                                                                `json:"name"`
			Value kbapi.KibanaHTTPAPIsUpdateAgentPolicyRequestBody_GlobalDataTags_Value `json:"value"`
		}, len(*tags))
		for i, tag := range *tags {
			updateTags[i].Name = tag.Name
			if value, err := tag.Value.AsAgentPolicyGlobalDataTagsItemValue0(); err == nil {
				if err := updateTags[i].Value.FromKibanaHTTPAPIsUpdateAgentPolicyRequestBodyGlobalDataTagsValue0(value); err != nil {
					diags.AddError("global_data_tags validation_error_converting_values", err.Error())
				}
			} else if value, err := tag.Value.AsAgentPolicyGlobalDataTagsItemValue1(); err == nil {
				if err := updateTags[i].Value.FromKibanaHTTPAPIsUpdateAgentPolicyRequestBodyGlobalDataTagsValue1(value); err != nil {
					diags.AddError("global_data_tags validation_error_converting_values", err.Error())
				}
			} else {
				diags.AddError("global_data_tags validation_error_converting_values", "unsupported global data tag value")
			}
		}
		body.GlobalDataTags = &updateTags
	}
	if diags.HasError() {
		return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
	}

	// Handle required_versions
	requiredVersions, d := model.convertRequiredVersions(feat)
	if d.HasError() {
		return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, d
	}
	body.RequiredVersions = requiredVersions

	// Handle host_name_format via AgentFeatures, preserving other existing features
	if agentFeature := model.convertHostNameFormatToAgentFeature(); agentFeature != nil {
		if feat.SupportsAgentFeatures {
			body.AgentFeatures = mergeAgentFeature(existingFeatures, agentFeature)
		}
	} else if feat.SupportsAgentFeatures && len(existingFeatures) > 0 {
		// Preserve existing features even when host_name_format is not set
		body.AgentFeatures = &existingFeatures
	}

	// Handle advanced_settings
	if typeutils.IsKnown(model.AdvancedSettings) && feat.SupportsAdvancedSettings {
		advancedSettings, diags := model.convertAdvancedSettingsToAPI(ctx, feat)
		if diags.HasError() {
			return kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody{}, diags
		}
		body.AdvancedSettings = advancedSettings
	}

	// Handle advanced monitoring options
	if typeutils.IsKnown(model.AdvancedMonitoringOptions) && feat.SupportsAdvancedMonitoring {
		monitoringHTTP, pprofEnabled := model.convertHTTPMonitoringEndpointToAPI(ctx)
		body.MonitoringHttp = monitoringHTTP
		body.MonitoringPprofEnabled = pprofEnabled
		body.MonitoringDiagnostics = model.convertDiagnosticsToAPI(ctx)
	}

	return body, nil
}

// featureGatedFields holds values for attributes gated behind minimum-version
// feature flags. Nil means the attribute was not set or is unavailable.
type featureGatedFields struct {
	isProtected       *bool
	supportsAgentless *bool
	inactivityTimeout *float32
	unenrollTimeout   *float32
	spaceIDs          *[]string
}

// computeFeatureGatedFields shapes version-gated attributes after
// GetVersionRequirements has rejected unsupported configured values.
func (model *agentPolicyModel) computeFeatureGatedFields(ctx context.Context, feat agentPolicyFeatures) (featureGatedFields, diag.Diagnostics) {
	var f featureGatedFields

	if typeutils.IsKnown(model.IsProtected) {
		if feat.SupportsTamperProtection {
			v := model.IsProtected.ValueBool()
			f.isProtected = &v
		}
	}

	if typeutils.IsKnown(model.SupportsAgentless) {
		if !feat.SupportsSupportsAgentless {
			return f, nil
		}
		f.supportsAgentless = model.SupportsAgentless.ValueBoolPointer()
	}

	if typeutils.IsKnown(model.InactivityTimeout) {
		if !feat.SupportsInactivityTimeout {
			return f, nil
		}
		duration, diags := model.InactivityTimeout.Parse()
		if diags.HasError() {
			return f, diags
		}
		seconds := float32(duration.Seconds())
		f.inactivityTimeout = &seconds
	}

	if typeutils.IsKnown(model.UnenrollmentTimeout) {
		if !feat.SupportsUnenrollmentTimeout {
			return f, nil
		}
		duration, diags := model.UnenrollmentTimeout.Parse()
		if diags.HasError() {
			return f, diags
		}
		seconds := float32(duration.Seconds())
		f.unenrollTimeout = &seconds
	}

	if typeutils.IsKnown(model.SpaceIDs) {
		if !feat.SupportsSpaceIDs {
			return f, nil
		}
		var spaceIDs []string
		diags := model.SpaceIDs.ElementsAs(ctx, &spaceIDs, false)
		if diags.HasError() {
			return f, diags
		}
		f.spaceIDs = &spaceIDs
	}

	return f, nil
}

// convertHostNameFormatToAgentFeature converts the host_name_format field to a single AgentFeature.
// - When host_name_format is "fqdn": returns {"name": "fqdn", "enabled": true}
// - When host_name_format is "hostname": returns {"name": "fqdn", "enabled": false} to explicitly disable
// - When not set: returns nil (no change to existing features)
func (model *agentPolicyModel) convertHostNameFormatToAgentFeature() *apiAgentFeature {
	// If host_name_format is not set or unknown, don't modify AgentFeatures
	if !typeutils.IsKnown(model.HostNameFormat) {
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
