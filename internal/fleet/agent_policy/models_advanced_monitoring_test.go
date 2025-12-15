package agent_policy

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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
			name: "default values are sent (allows reset to defaults)",
			amo: createAmoObject(createHttpEndpointObject(httpMonitoringEndpointModel{
				Enabled:       types.BoolValue(false),
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
			name: "default rate limits values are sent (allows reset to defaults)",
			amo: createAmoObject(createDiagnosticsObject(
				createRateLimitsObject(rateLimitsModel{
					Interval: customtypes.NewDurationValue("1m"),
					Burst:    types.Int32Value(1),
				}),
				types.ObjectNull(fileUploaderAttrTypes()),
			)),
			wantDiag:       true,
			wantRateLimits: true,
		},
		{
			name: "default uploader values are sent (allows reset to defaults)",
			amo: createAmoObject(createDiagnosticsObject(
				types.ObjectNull(rateLimitsAttrTypes()),
				createFileUploaderObject(fileUploaderModel{
					InitDuration:    customtypes.NewDurationValue("1s"),
					BackoffDuration: customtypes.NewDurationValue("1m"),
					MaxRetries:      types.Int32Value(10),
				}),
			)),
			wantDiag:     true,
			wantUploader: true,
		},
		{
			name: "custom rate limits interval returns values",
			amo: createAmoObject(createDiagnosticsObject(
				createRateLimitsObject(rateLimitsModel{
					Interval: customtypes.NewDurationValue("2m"),
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
					Interval: customtypes.NewDurationValue("1m"),
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
					InitDuration:    customtypes.NewDurationValue("1s"),
					BackoffDuration: customtypes.NewDurationValue("1m"),
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
