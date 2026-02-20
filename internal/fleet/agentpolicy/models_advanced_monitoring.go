package agentpolicy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Advanced Monitoring Options models
type advancedMonitoringOptionsModel struct {
	HTTPMonitoringEndpoint types.Object `tfsdk:"http_monitoring_endpoint"`
	Diagnostics            types.Object `tfsdk:"diagnostics"`
}

type httpMonitoringEndpointModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	Host          types.String `tfsdk:"host"`
	Port          types.Int32  `tfsdk:"port"`
	BufferEnabled types.Bool   `tfsdk:"buffer_enabled"`
	PprofEnabled  types.Bool   `tfsdk:"pprof_enabled"`
}

type diagnosticsModel struct {
	RateLimits   types.Object `tfsdk:"rate_limits"`
	FileUploader types.Object `tfsdk:"file_uploader"`
}

type rateLimitsModel struct {
	Interval customtypes.Duration `tfsdk:"interval"`
	Burst    types.Int32          `tfsdk:"burst"`
}

type fileUploaderModel struct {
	InitDuration    customtypes.Duration `tfsdk:"init_duration"`
	BackoffDuration customtypes.Duration `tfsdk:"backoff_duration"`
	MaxRetries      types.Int32          `tfsdk:"max_retries"`
}

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

// populateAdvancedMonitoringFromAPI populates the advanced monitoring options from API response
func (model *agentPolicyModel) populateAdvancedMonitoringFromAPI(ctx context.Context, data *kbapi.AgentPolicy) diag.Diagnostics {
	// Check if any advanced monitoring data exists in the API response
	hasHTTPMonitoring := data.MonitoringHttp != nil
	hasPprofEnabled := data.MonitoringPprofEnabled != nil
	hasDiagnostics := data.MonitoringDiagnostics != nil

	if !hasHTTPMonitoring && !hasPprofEnabled && !hasDiagnostics {
		// No advanced monitoring options in API response
		model.AdvancedMonitoringOptions = types.ObjectNull(advancedMonitoringOptionsAttrTypes())
		return nil
	}

	var httpEndpointObj types.Object
	var diagnosticsObj types.Object

	// Populate HTTP monitoring endpoint
	if hasHTTPMonitoring || hasPprofEnabled {
		httpEndpoint := httpMonitoringEndpointModel{
			Enabled:       types.BoolValue(defaultHTTPMonitoringEnabled),
			Host:          types.StringValue(defaultHTTPMonitoringHost),
			Port:          types.Int32Value(defaultHTTPMonitoringPort),
			BufferEnabled: types.BoolValue(defaultHTTPMonitoringBufferEnabled),
			PprofEnabled:  types.BoolValue(defaultHTTPMonitoringPprofEnabled),
		}

		if data.MonitoringHttp != nil {
			if data.MonitoringHttp.Enabled != nil {
				httpEndpoint.Enabled = types.BoolValue(*data.MonitoringHttp.Enabled)
			}
			if data.MonitoringHttp.Host != nil {
				httpEndpoint.Host = types.StringValue(*data.MonitoringHttp.Host)
			}
			if data.MonitoringHttp.Port != nil {
				httpEndpoint.Port = types.Int32Value(int32(*data.MonitoringHttp.Port))
			}
			if data.MonitoringHttp.Buffer != nil && data.MonitoringHttp.Buffer.Enabled != nil {
				httpEndpoint.BufferEnabled = types.BoolValue(*data.MonitoringHttp.Buffer.Enabled)
			}
		}

		if data.MonitoringPprofEnabled != nil {
			httpEndpoint.PprofEnabled = types.BoolValue(*data.MonitoringPprofEnabled)
		}

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

		if data.MonitoringDiagnostics.Limit != nil {
			rateLimits := rateLimitsModel{
				Interval: customtypes.NewDurationValue(defaultDiagnosticsInterval),
				Burst:    types.Int32Value(defaultDiagnosticsBurst),
			}
			if data.MonitoringDiagnostics.Limit.Interval != nil {
				rateLimits.Interval = customtypes.NewDurationValue(*data.MonitoringDiagnostics.Limit.Interval)
			}
			if data.MonitoringDiagnostics.Limit.Burst != nil {
				rateLimits.Burst = types.Int32Value(int32(*data.MonitoringDiagnostics.Limit.Burst))
			}
			obj, diags := types.ObjectValueFrom(ctx, rateLimitsAttrTypes(), rateLimits)
			if diags.HasError() {
				return diags
			}
			rateLimitsObj = obj
		} else {
			rateLimitsObj = types.ObjectNull(rateLimitsAttrTypes())
		}

		if data.MonitoringDiagnostics.Uploader != nil {
			fileUploader := fileUploaderModel{
				InitDuration:    customtypes.NewDurationValue(defaultDiagnosticsInitDuration),
				BackoffDuration: customtypes.NewDurationValue(defaultDiagnosticsBackoffDuration),
				MaxRetries:      types.Int32Value(defaultDiagnosticsMaxRetries),
			}
			if data.MonitoringDiagnostics.Uploader.InitDur != nil {
				fileUploader.InitDuration = customtypes.NewDurationValue(*data.MonitoringDiagnostics.Uploader.InitDur)
			}
			if data.MonitoringDiagnostics.Uploader.MaxDur != nil {
				fileUploader.BackoffDuration = customtypes.NewDurationValue(*data.MonitoringDiagnostics.Uploader.MaxDur)
			}
			if data.MonitoringDiagnostics.Uploader.MaxRetries != nil {
				fileUploader.MaxRetries = types.Int32Value(int32(*data.MonitoringDiagnostics.Uploader.MaxRetries))
			}
			obj, diags := types.ObjectValueFrom(ctx, fileUploaderAttrTypes(), fileUploader)
			if diags.HasError() {
				return diags
			}
			fileUploaderObj = obj
		} else {
			fileUploaderObj = types.ObjectNull(fileUploaderAttrTypes())
		}

		diagModel := diagnosticsModel{
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

	amo := advancedMonitoringOptionsModel{
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

	var amo advancedMonitoringOptionsModel
	model.AdvancedMonitoringOptions.As(ctx, &amo, basetypes.ObjectAsOptions{})

	if !typeutils.IsKnown(amo.HTTPMonitoringEndpoint) {
		return nil, nil
	}

	var http httpMonitoringEndpointModel
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

	var amo advancedMonitoringOptionsModel
	model.AdvancedMonitoringOptions.As(ctx, &amo, basetypes.ObjectAsOptions{})

	if !typeutils.IsKnown(amo.Diagnostics) {
		return nil
	}

	var diag diagnosticsModel
	amo.Diagnostics.As(ctx, &diag, basetypes.ObjectAsOptions{})

	result := &diagnosticsAPIResult{}

	if typeutils.IsKnown(diag.RateLimits) {
		var rateLimits rateLimitsModel
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
		var fileUploader fileUploaderModel
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
