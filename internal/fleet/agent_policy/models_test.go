package agent_policy

import (
	"context"
	"testing"

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

func TestConvertHttpMonitoringEndpointToAPI(t *testing.T) {
	ctx := context.Background()

	// Helper to create types.Object from httpMonitoringEndpointModel
	createHttpEndpointObject := func(m httpMonitoringEndpointModel) types.Object {
		obj, _ := types.ObjectValueFrom(ctx, httpMonitoringEndpointAttrTypes(), m)
		return obj
	}

	// Helper to create types.Object from advancedMonitoringOptionsModel
	createAmoObject := func(httpEndpoint types.Object) types.Object {
		amo := advancedMonitoringOptionsModel{
			HttpMonitoringEndpoint: httpEndpoint,
			Diagnostics:            types.ObjectNull(diagnosticsAttrTypes()),
		}
		obj, _ := types.ObjectValueFrom(ctx, advancedMonitoringOptionsAttrTypes(), amo)
		return obj
	}

	tests := []struct {
		name           string
		amo            types.Object
		wantHttp       bool
		wantPprof      bool
		wantPprofValue bool
	}{
		{
			name:     "null advanced monitoring options returns nil",
			amo:      types.ObjectNull(advancedMonitoringOptionsAttrTypes()),
			wantHttp: false,
		},
		{
			name:     "null http monitoring endpoint returns nil",
			amo:      createAmoObject(types.ObjectNull(httpMonitoringEndpointAttrTypes())),
			wantHttp: false,
		},
		{
			name: "default values returns nil (omit from payload)",
			amo: createAmoObject(createHttpEndpointObject(httpMonitoringEndpointModel{
				Enabled:       types.BoolValue(false),
				Host:          types.StringValue("localhost"),
				Port:          types.Int32Value(6791),
				BufferEnabled: types.BoolValue(false),
				PprofEnabled:  types.BoolValue(false),
			})),
			wantHttp: false,
		},
		{
			name: "enabled http endpoint returns values",
			amo: createAmoObject(createHttpEndpointObject(httpMonitoringEndpointModel{
				Enabled:       types.BoolValue(true),
				Host:          types.StringValue("localhost"),
				Port:          types.Int32Value(6791),
				BufferEnabled: types.BoolValue(false),
				PprofEnabled:  types.BoolValue(false),
			})),
			wantHttp:       true,
			wantPprof:      true,
			wantPprofValue: false,
		},
		{
			name: "custom port returns values",
			amo: createAmoObject(createHttpEndpointObject(httpMonitoringEndpointModel{
				Enabled:       types.BoolValue(false),
				Host:          types.StringValue("localhost"),
				Port:          types.Int32Value(8080),
				BufferEnabled: types.BoolValue(false),
				PprofEnabled:  types.BoolValue(false),
			})),
			wantHttp:       true,
			wantPprof:      true,
			wantPprofValue: false,
		},
		{
			name: "pprof enabled returns values",
			amo: createAmoObject(createHttpEndpointObject(httpMonitoringEndpointModel{
				Enabled:       types.BoolValue(true),
				Host:          types.StringValue("localhost"),
				Port:          types.Int32Value(6791),
				BufferEnabled: types.BoolValue(false),
				PprofEnabled:  types.BoolValue(true),
			})),
			wantHttp:       true,
			wantPprof:      true,
			wantPprofValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &agentPolicyModel{
				AdvancedMonitoringOptions: tt.amo,
			}

			gotHttp, gotPprof := model.convertHttpMonitoringEndpointToAPI(ctx)

			if !tt.wantHttp {
				assert.Nil(t, gotHttp)
				assert.Nil(t, gotPprof)
				return
			}

			assert.NotNil(t, gotHttp)
			if tt.wantPprof {
				assert.NotNil(t, gotPprof)
				assert.Equal(t, tt.wantPprofValue, *gotPprof)
			}
		})
	}
}

func TestConvertDiagnosticsToAPI(t *testing.T) {
	ctx := context.Background()

	// Helper to create types.Object from rateLimitsModel
	createRateLimitsObject := func(m rateLimitsModel) types.Object {
		obj, _ := types.ObjectValueFrom(ctx, rateLimitsAttrTypes(), m)
		return obj
	}

	// Helper to create types.Object from fileUploaderModel
	createFileUploaderObject := func(m fileUploaderModel) types.Object {
		obj, _ := types.ObjectValueFrom(ctx, fileUploaderAttrTypes(), m)
		return obj
	}

	// Helper to create types.Object from diagnosticsModel
	createDiagnosticsObject := func(rateLimits, fileUploader types.Object) types.Object {
		diag := diagnosticsModel{
			RateLimits:   rateLimits,
			FileUploader: fileUploader,
		}
		obj, _ := types.ObjectValueFrom(ctx, diagnosticsAttrTypes(), diag)
		return obj
	}

	// Helper to create types.Object from advancedMonitoringOptionsModel
	createAmoObject := func(diagnostics types.Object) types.Object {
		amo := advancedMonitoringOptionsModel{
			HttpMonitoringEndpoint: types.ObjectNull(httpMonitoringEndpointAttrTypes()),
			Diagnostics:            diagnostics,
		}
		obj, _ := types.ObjectValueFrom(ctx, advancedMonitoringOptionsAttrTypes(), amo)
		return obj
	}

	tests := []struct {
		name           string
		amo            types.Object
		wantDiag       bool
		wantRateLimits bool
		wantUploader   bool
	}{
		{
			name:     "null advanced monitoring options returns nil",
			amo:      types.ObjectNull(advancedMonitoringOptionsAttrTypes()),
			wantDiag: false,
		},
		{
			name:     "null diagnostics returns nil",
			amo:      createAmoObject(types.ObjectNull(diagnosticsAttrTypes())),
			wantDiag: false,
		},
		{
			name: "default rate limits values returns nil (omit from payload)",
			amo: createAmoObject(createDiagnosticsObject(
				createRateLimitsObject(rateLimitsModel{
					Interval: types.StringValue("1m"),
					Burst:    types.Int32Value(1),
				}),
				types.ObjectNull(fileUploaderAttrTypes()),
			)),
			wantDiag: false,
		},
		{
			name: "default uploader values returns nil (omit from payload)",
			amo: createAmoObject(createDiagnosticsObject(
				types.ObjectNull(rateLimitsAttrTypes()),
				createFileUploaderObject(fileUploaderModel{
					InitDuration:    types.StringValue("1s"),
					BackoffDuration: types.StringValue("1m"),
					MaxRetries:      types.Int32Value(10),
				}),
			)),
			wantDiag: false,
		},
		{
			name: "custom rate limits interval returns values",
			amo: createAmoObject(createDiagnosticsObject(
				createRateLimitsObject(rateLimitsModel{
					Interval: types.StringValue("2m"),
					Burst:    types.Int32Value(1),
				}),
				types.ObjectNull(fileUploaderAttrTypes()),
			)),
			wantDiag:       true,
			wantRateLimits: true,
		},
		{
			name: "custom rate limits burst returns values",
			amo: createAmoObject(createDiagnosticsObject(
				createRateLimitsObject(rateLimitsModel{
					Interval: types.StringValue("1m"),
					Burst:    types.Int32Value(5),
				}),
				types.ObjectNull(fileUploaderAttrTypes()),
			)),
			wantDiag:       true,
			wantRateLimits: true,
		},
		{
			name: "custom uploader max_retries returns values",
			amo: createAmoObject(createDiagnosticsObject(
				types.ObjectNull(rateLimitsAttrTypes()),
				createFileUploaderObject(fileUploaderModel{
					InitDuration:    types.StringValue("1s"),
					BackoffDuration: types.StringValue("1m"),
					MaxRetries:      types.Int32Value(20),
				}),
			)),
			wantDiag:     true,
			wantUploader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &agentPolicyModel{
				AdvancedMonitoringOptions: tt.amo,
			}

			got := model.convertDiagnosticsToAPI(ctx)

			if !tt.wantDiag {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			if tt.wantRateLimits {
				assert.NotNil(t, got.Limit)
			}
			if tt.wantUploader {
				assert.NotNil(t, got.Uploader)
			}
		})
	}
}
