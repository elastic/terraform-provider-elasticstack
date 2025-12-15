package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
	MonitoringRuntimeExperimental types.String         `tfsdk:"monitoring_runtime_experimental"`
}

// advancedSettingsAttrTypes returns attribute types for advanced_settings pulled from the schema
func advancedSettingsAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["advanced_settings"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// advancedSettingsAPIResult is the return type for convertAdvancedSettingsToAPI
type advancedSettingsAPIResult = struct {
	AgentDownloadTargetDirectory       interface{} `json:"agent_download_target_directory,omitempty"`
	AgentDownloadTimeout               interface{} `json:"agent_download_timeout,omitempty"`
	AgentInternal                      interface{} `json:"agent_internal,omitempty"`
	AgentLimitsGoMaxProcs              interface{} `json:"agent_limits_go_max_procs,omitempty"`
	AgentLoggingFilesInterval          interface{} `json:"agent_logging_files_interval,omitempty"`
	AgentLoggingFilesKeepfiles         interface{} `json:"agent_logging_files_keepfiles,omitempty"`
	AgentLoggingFilesRotateeverybytes  interface{} `json:"agent_logging_files_rotateeverybytes,omitempty"`
	AgentLoggingLevel                  interface{} `json:"agent_logging_level,omitempty"`
	AgentLoggingMetricsPeriod          interface{} `json:"agent_logging_metrics_period,omitempty"`
	AgentLoggingToFiles                interface{} `json:"agent_logging_to_files,omitempty"`
	AgentMonitoringRuntimeExperimental interface{} `json:"agent_monitoring_runtime_experimental,omitempty"`
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
		if str, ok := data.AdvancedSettings.AgentMonitoringRuntimeExperimental.(string); ok {
			settings.MonitoringRuntimeExperimental = types.StringValue(str)
		} else {
			settings.MonitoringRuntimeExperimental = types.StringNull()
		}
	} else {
		settings.MonitoringRuntimeExperimental = types.StringNull()
	}

	obj, diags := types.ObjectValueFrom(ctx, advancedSettingsAttrTypes(), settings)
	if diags.HasError() {
		return diags
	}
	model.AdvancedSettings = obj
	return nil
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
		result.AgentMonitoringRuntimeExperimental = settings.MonitoringRuntimeExperimental.ValueString()
	}

	return result
}

