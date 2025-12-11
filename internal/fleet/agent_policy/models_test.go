package agent_policy

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMergeAgentFeature(t *testing.T) {
	tests := []struct {
		name       string
		existing   []apiAgentFeature
		newFeature *apiAgentFeature
		want       *[]apiAgentFeature
	}{
		{
			name:       "nil new feature with empty existing returns nil",
			existing:   nil,
			newFeature: nil,
			want:       nil,
		},
		{
			name:       "nil new feature with empty slice returns nil",
			existing:   []apiAgentFeature{},
			newFeature: nil,
			want:       nil,
		},
		{
			name: "nil new feature preserves existing features",
			existing: []apiAgentFeature{
				{Name: "feature1", Enabled: true},
				{Name: "feature2", Enabled: false},
			},
			newFeature: nil,
			want: &[]apiAgentFeature{
				{Name: "feature1", Enabled: true},
				{Name: "feature2", Enabled: false},
			},
		},
		{
			name:       "new feature added to empty existing",
			existing:   nil,
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: true},
			},
		},
		{
			name: "new feature added when not present",
			existing: []apiAgentFeature{
				{Name: "other", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "other", Enabled: true},
				{Name: "fqdn", Enabled: true},
			},
		},
		{
			name: "existing feature replaced",
			existing: []apiAgentFeature{
				{Name: "fqdn", Enabled: false},
				{Name: "other", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: true},
				{Name: "other", Enabled: true},
			},
		},
		{
			name: "feature disabled replaces enabled",
			existing: []apiAgentFeature{
				{Name: "fqdn", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: false},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeAgentFeature(tt.existing, tt.newFeature)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

func TestConvertHostNameFormatToAgentFeature(t *testing.T) {
	tests := []struct {
		name           string
		hostNameFormat types.String
		want           *apiAgentFeature
	}{
		{
			name:           "null host_name_format returns nil",
			hostNameFormat: types.StringNull(),
			want:           nil,
		},
		{
			name:           "unknown host_name_format returns nil",
			hostNameFormat: types.StringUnknown(),
			want:           nil,
		},
		{
			name:           "fqdn returns enabled feature",
			hostNameFormat: types.StringValue(HostNameFormatFQDN),
			want:           &apiAgentFeature{Name: agentFeatureFQDN, Enabled: true},
		},
		{
			name:           "hostname returns disabled feature",
			hostNameFormat: types.StringValue(HostNameFormatHostname),
			want:           &apiAgentFeature{Name: agentFeatureFQDN, Enabled: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &agentPolicyModel{
				HostNameFormat: tt.hostNameFormat,
			}

			got := model.convertHostNameFormatToAgentFeature()

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Enabled, got.Enabled)
		})
	}
}

func TestConvertAdvancedSettingsToAPI(t *testing.T) {
	ctx := context.Background()

	createAdvancedSettingsObject := func(settings advancedSettingsModel) types.Object {
		obj, _ := types.ObjectValueFrom(ctx, advancedSettingsAttrTypes(), settings)
		return obj
	}

	tests := []struct {
		name             string
		advancedSettings types.Object
		wantNil          bool
		checkResult      func(t *testing.T, result *advancedSettingsAPIResult)
	}{
		{
			name:             "null advanced_settings returns nil",
			advancedSettings: types.ObjectNull(advancedSettingsAttrTypes()),
			wantNil:          true,
		},
		{
			name: "all null values returns nil",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringNull(),
				LoggingToFiles:                types.BoolNull(),
				LoggingFilesInterval:          customtypes.NewDurationNull(),
				LoggingFilesKeepfiles:         types.Int32Null(),
				LoggingFilesRotateeverybytes:  types.Int64Null(),
				LoggingMetricsPeriod:          customtypes.NewDurationNull(),
				GoMaxProcs:                    types.Int32Null(),
				DownloadTimeout:               customtypes.NewDurationNull(),
				DownloadTargetDirectory:       types.StringNull(),
				MonitoringRuntimeExperimental: types.BoolNull(),
			}),
			wantNil: true,
		},
		{
			name: "logging_level set returns value",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringValue("debug"),
				LoggingToFiles:                types.BoolNull(),
				LoggingFilesInterval:          customtypes.NewDurationNull(),
				LoggingFilesKeepfiles:         types.Int32Null(),
				LoggingFilesRotateeverybytes:  types.Int64Null(),
				LoggingMetricsPeriod:          customtypes.NewDurationNull(),
				GoMaxProcs:                    types.Int32Null(),
				DownloadTimeout:               customtypes.NewDurationNull(),
				DownloadTargetDirectory:       types.StringNull(),
				MonitoringRuntimeExperimental: types.BoolNull(),
			}),
			wantNil: false,
			checkResult: func(t *testing.T, result *advancedSettingsAPIResult) {
				assert.Equal(t, "debug", result.AgentLoggingLevel)
				assert.Nil(t, result.AgentLoggingToFiles)
			},
		},
		{
			name: "go_max_procs set returns value",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringNull(),
				LoggingToFiles:                types.BoolNull(),
				LoggingFilesInterval:          customtypes.NewDurationNull(),
				LoggingFilesKeepfiles:         types.Int32Null(),
				LoggingFilesRotateeverybytes:  types.Int64Null(),
				LoggingMetricsPeriod:          customtypes.NewDurationNull(),
				GoMaxProcs:                    types.Int32Value(4),
				DownloadTimeout:               customtypes.NewDurationNull(),
				DownloadTargetDirectory:       types.StringNull(),
				MonitoringRuntimeExperimental: types.BoolNull(),
			}),
			wantNil: false,
			checkResult: func(t *testing.T, result *advancedSettingsAPIResult) {
				assert.Equal(t, int32(4), result.AgentLimitsGoMaxProcs)
			},
		},
		{
			name: "multiple values set returns all values",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringValue("info"),
				LoggingToFiles:                types.BoolValue(true),
				LoggingFilesInterval:          customtypes.NewDurationValue("30s"),
				LoggingFilesKeepfiles:         types.Int32Value(7),
				LoggingFilesRotateeverybytes:  types.Int64Value(10485760),
				LoggingMetricsPeriod:          customtypes.NewDurationValue("1m"),
				GoMaxProcs:                    types.Int32Value(2),
				DownloadTimeout:               customtypes.NewDurationValue("2h"),
				DownloadTargetDirectory:       types.StringValue("/tmp/elastic"),
				MonitoringRuntimeExperimental: types.BoolValue(false),
			}),
			wantNil: false,
			checkResult: func(t *testing.T, result *advancedSettingsAPIResult) {
				assert.Equal(t, "info", result.AgentLoggingLevel)
				assert.Equal(t, true, result.AgentLoggingToFiles)
				assert.Equal(t, "30s", result.AgentLoggingFilesInterval)
				assert.Equal(t, int32(7), result.AgentLoggingFilesKeepfiles)
				assert.Equal(t, int64(10485760), result.AgentLoggingFilesRotateeverybytes)
				assert.Equal(t, "1m", result.AgentLoggingMetricsPeriod)
				assert.Equal(t, int32(2), result.AgentLimitsGoMaxProcs)
				assert.Equal(t, "2h", result.AgentDownloadTimeout)
				assert.Equal(t, "/tmp/elastic", result.AgentDownloadTargetDirectory)
				assert.Equal(t, false, result.AgentMonitoringRuntimeExperimental)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &agentPolicyModel{
				AdvancedSettings: tt.advancedSettings,
			}

			got := model.convertAdvancedSettingsToAPI(ctx)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}
