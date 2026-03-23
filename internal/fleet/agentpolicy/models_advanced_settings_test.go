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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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
				MonitoringRuntimeExperimental: types.StringNull(),
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
				MonitoringRuntimeExperimental: types.StringNull(),
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
				MonitoringRuntimeExperimental: types.StringNull(),
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
				MonitoringRuntimeExperimental: types.StringValue(""),
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
				assert.Empty(t, result.AgentMonitoringRuntimeExperimental)
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
