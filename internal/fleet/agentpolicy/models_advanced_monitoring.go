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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentpolicyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Default values for advanced monitoring options
const (
	defaultHTTPMonitoringEnabled       = false
	defaultHTTPMonitoringHost          = "localhost"
	defaultHTTPMonitoringPort          = 6791
	defaultHTTPMonitoringBufferEnabled = false
	defaultHTTPMonitoringPprofEnabled  = false
	defaultDiagnosticsInterval         = "1m"
	defaultDiagnosticsBurst            = 1
	defaultDiagnosticsInitDuration     = "1s"
	defaultDiagnosticsBackoffDuration  = "1m"
	defaultDiagnosticsMaxRetries       = 10
)

// Attribute type helpers for advanced monitoring options - pulled from schema to avoid duplication
func advancedMonitoringOptionsAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["advanced_monitoring_options"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func httpMonitoringEndpointAttrTypes() map[string]attr.Type {
	amoAttr := getSchema().Attributes["advanced_monitoring_options"].(schema.SingleNestedAttribute)
	return amoAttr.Attributes["http_monitoring_endpoint"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func diagnosticsAttrTypes() map[string]attr.Type {
	amoAttr := getSchema().Attributes["advanced_monitoring_options"].(schema.SingleNestedAttribute)
	return amoAttr.Attributes["diagnostics"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func rateLimitsAttrTypes() map[string]attr.Type {
	amoAttr := getSchema().Attributes["advanced_monitoring_options"].(schema.SingleNestedAttribute)
	diagAttr := amoAttr.Attributes["diagnostics"].(schema.SingleNestedAttribute)
	return diagAttr.Attributes["rate_limits"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func fileUploaderAttrTypes() map[string]attr.Type {
	amoAttr := getSchema().Attributes["advanced_monitoring_options"].(schema.SingleNestedAttribute)
	diagAttr := amoAttr.Attributes["diagnostics"].(schema.SingleNestedAttribute)
	return diagAttr.Attributes["file_uploader"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// httpMonitoringEndpointDefaults returns the baseline http_monitoring_endpoint
// values the resource reports when the Fleet API confirms the section exists
// but omits a specific field, matching the schema's static defaults.
func httpMonitoringEndpointDefaults() agentpolicyshape.HTTPMonitoringEndpointModel {
	return agentpolicyshape.HTTPMonitoringEndpointModel{
		Enabled:       types.BoolValue(defaultHTTPMonitoringEnabled),
		Host:          types.StringValue(defaultHTTPMonitoringHost),
		Port:          types.Int32Value(defaultHTTPMonitoringPort),
		BufferEnabled: types.BoolValue(defaultHTTPMonitoringBufferEnabled),
		PprofEnabled:  types.BoolValue(defaultHTTPMonitoringPprofEnabled),
	}
}

// rateLimitsDefaults returns the baseline diagnostics.rate_limits values the
// resource reports when the Fleet API confirms diagnostics rate limiting
// exists but omits a specific field, matching the schema's static defaults.
func rateLimitsDefaults() agentpolicyshape.RateLimitsModel {
	return agentpolicyshape.RateLimitsModel{
		Interval: customtypes.NewDurationValue(defaultDiagnosticsInterval),
		Burst:    types.Int32Value(defaultDiagnosticsBurst),
	}
}

// fileUploaderDefaults returns the baseline diagnostics.file_uploader values
// the resource reports when the Fleet API confirms the file uploader exists
// but omits a specific field, matching the schema's static defaults.
func fileUploaderDefaults() agentpolicyshape.FileUploaderModel {
	return agentpolicyshape.FileUploaderModel{
		InitDuration:    customtypes.NewDurationValue(defaultDiagnosticsInitDuration),
		BackoffDuration: customtypes.NewDurationValue(defaultDiagnosticsBackoffDuration),
		MaxRetries:      types.Int32Value(defaultDiagnosticsMaxRetries),
	}
}

// populateAdvancedMonitoringFromAPI populates the advanced monitoring options from API response
func (model *agentPolicyModel) populateAdvancedMonitoringFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	httpEndpoint, hasHTTP := agentpolicyshape.PopulateHTTPMonitoringEndpoint(data, httpMonitoringEndpointDefaults())
	hasDiagnostics := agentpolicyshape.HasDiagnostics(data)

	if !hasHTTP && !hasDiagnostics {
		// No advanced monitoring options in API response
		model.AdvancedMonitoringOptions = types.ObjectNull(advancedMonitoringOptionsAttrTypes())
		return nil
	}

	var httpEndpointObj types.Object
	var diagnosticsObj types.Object

	// Populate HTTP monitoring endpoint
	if hasHTTP {
		obj, diags := types.ObjectValueFrom(ctx, httpMonitoringEndpointAttrTypes(), httpEndpoint)
		if diags.HasError() {
			return diags
		}
		httpEndpointObj = obj
	} else {
		httpEndpointObj = types.ObjectNull(httpMonitoringEndpointAttrTypes())
	}

	// Populate diagnostics
	if hasDiagnostics {
		var rateLimitsObj types.Object
		var fileUploaderObj types.Object

		if rateLimits, ok := agentpolicyshape.PopulateRateLimits(data, rateLimitsDefaults()); ok {
			obj, diags := types.ObjectValueFrom(ctx, rateLimitsAttrTypes(), rateLimits)
			if diags.HasError() {
				return diags
			}
			rateLimitsObj = obj
		} else {
			rateLimitsObj = types.ObjectNull(rateLimitsAttrTypes())
		}

		if fileUploader, ok := agentpolicyshape.PopulateFileUploader(data, fileUploaderDefaults()); ok {
			obj, diags := types.ObjectValueFrom(ctx, fileUploaderAttrTypes(), fileUploader)
			if diags.HasError() {
				return diags
			}
			fileUploaderObj = obj
		} else {
			fileUploaderObj = types.ObjectNull(fileUploaderAttrTypes())
		}

		diagModel := agentpolicyshape.DiagnosticsModel{
			RateLimits:   rateLimitsObj,
			FileUploader: fileUploaderObj,
		}
		obj, diags := types.ObjectValueFrom(ctx, diagnosticsAttrTypes(), diagModel)
		if diags.HasError() {
			return diags
		}
		diagnosticsObj = obj
	} else {
		diagnosticsObj = types.ObjectNull(diagnosticsAttrTypes())
	}

	amo := agentpolicyshape.AdvancedMonitoringOptionsModel{
		HTTPMonitoringEndpoint: httpEndpointObj,
		Diagnostics:            diagnosticsObj,
	}

	obj, diags := types.ObjectValueFrom(ctx, advancedMonitoringOptionsAttrTypes(), amo)
	if diags.HasError() {
		return diags
	}
	model.AdvancedMonitoringOptions = obj
	return nil
}

// httpMonitoringEndpointAPIResult is the return type for convertHTTPMonitoringEndpointToAPI
// This type alias matches the inline struct expected by kbapi.PostFleetAgentPoliciesJSONRequestBody.MonitoringHttp
type httpMonitoringEndpointAPIResult = struct {
	Buffer *struct {
		Enabled *bool `json:"enabled,omitempty"`
	} `json:"buffer,omitempty"`
	Enabled *bool    `json:"enabled,omitempty"`
	Host    *string  `json:"host,omitempty"`
	Port    *float32 `json:"port,omitempty"`
}

// convertHTTPMonitoringEndpointToAPI converts the HTTP monitoring endpoint config to API format
func (model *agentPolicyModel) convertHTTPMonitoringEndpointToAPI(ctx context.Context) (*httpMonitoringEndpointAPIResult, *bool) {
	if !typeutils.IsKnown(model.AdvancedMonitoringOptions) {
		return nil, nil
	}

	var amo agentpolicyshape.AdvancedMonitoringOptionsModel
	model.AdvancedMonitoringOptions.As(ctx, &amo, basetypes.ObjectAsOptions{})

	if !typeutils.IsKnown(amo.HTTPMonitoringEndpoint) {
		return nil, nil
	}

	var http agentpolicyshape.HTTPMonitoringEndpointModel
	amo.HTTPMonitoringEndpoint.As(ctx, &http, basetypes.ObjectAsOptions{})

	enabled := http.Enabled.ValueBool()
	host := http.Host.ValueString()
	port := float32(http.Port.ValueInt32())
	bufferEnabled := http.BufferEnabled.ValueBool()
	pprofEnabled := http.PprofEnabled.ValueBool()

	result := &httpMonitoringEndpointAPIResult{
		Enabled: &enabled,
		Host:    &host,
		Port:    &port,
		Buffer: &struct {
			Enabled *bool `json:"enabled,omitempty"`
		}{
			Enabled: &bufferEnabled,
		},
	}

	return result, &pprofEnabled
}

// diagnosticsAPIResult is the return type for convertDiagnosticsToAPI
// This type alias matches the inline struct expected by kbapi.PostFleetAgentPoliciesJSONRequestBody.MonitoringDiagnostics
type diagnosticsAPIResult = struct {
	Limit *struct {
		Burst    *float32 `json:"burst,omitempty"`
		Interval *string  `json:"interval,omitempty"`
	} `json:"limit,omitempty"`
	Uploader *struct {
		InitDur    *string  `json:"init_dur,omitempty"`
		MaxDur     *string  `json:"max_dur,omitempty"`
		MaxRetries *float32 `json:"max_retries,omitempty"`
	} `json:"uploader,omitempty"`
}

// convertDiagnosticsToAPI converts the diagnostics config to API format
func (model *agentPolicyModel) convertDiagnosticsToAPI(ctx context.Context) *diagnosticsAPIResult {
	if !typeutils.IsKnown(model.AdvancedMonitoringOptions) {
		return nil
	}

	var amo agentpolicyshape.AdvancedMonitoringOptionsModel
	model.AdvancedMonitoringOptions.As(ctx, &amo, basetypes.ObjectAsOptions{})

	if !typeutils.IsKnown(amo.Diagnostics) {
		return nil
	}

	var diag agentpolicyshape.DiagnosticsModel
	amo.Diagnostics.As(ctx, &diag, basetypes.ObjectAsOptions{})

	result := &diagnosticsAPIResult{}

	if typeutils.IsKnown(diag.RateLimits) {
		var rateLimits agentpolicyshape.RateLimitsModel
		diag.RateLimits.As(ctx, &rateLimits, basetypes.ObjectAsOptions{})
		interval := rateLimits.Interval.ValueString()
		burst := float32(rateLimits.Burst.ValueInt32())
		result.Limit = &struct {
			Burst    *float32 `json:"burst,omitempty"`
			Interval *string  `json:"interval,omitempty"`
		}{
			Interval: &interval,
			Burst:    &burst,
		}
	}

	if typeutils.IsKnown(diag.FileUploader) {
		var fileUploader agentpolicyshape.FileUploaderModel
		diag.FileUploader.As(ctx, &fileUploader, basetypes.ObjectAsOptions{})
		initDur := fileUploader.InitDuration.ValueString()
		maxDur := fileUploader.BackoffDuration.ValueString()
		maxRetries := float32(fileUploader.MaxRetries.ValueInt32())
		result.Uploader = &struct {
			InitDur    *string  `json:"init_dur,omitempty"`
			MaxDur     *string  `json:"max_dur,omitempty"`
			MaxRetries *float32 `json:"max_retries,omitempty"`
		}{
			InitDur:    &initDur,
			MaxDur:     &maxDur,
			MaxRetries: &maxRetries,
		}
	}

	return result
}
