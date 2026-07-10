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

package apmservicemap_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/apmservicemap"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfig_nilConfig(t *testing.T) {
	pm := models.PanelModel{}
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap
	diags := apmservicemap.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{}, panel.Config)
}

func TestBuildConfig_allOptionalFields(t *testing.T) {
	pm := models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			Title:                    types.StringValue("APM Map"),
			Description:              types.StringValue("Service dependencies"),
			HideTitle:                types.BoolValue(true),
			HideBorder:               types.BoolValue(false),
			Environment:              types.StringValue("production"),
			ServiceName:              types.StringValue("checkout"),
			ServiceGroupID:           types.StringValue("group-abc"),
			Kuery:                    types.StringValue("service.name : checkout"),
			MapOrientation:           types.StringValue("horizontal"),
			SyncWithDashboardFilters: types.BoolValue(true),
			AlertStatusFilter: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("active"),
				types.StringValue("delayed"),
			}),
			AnomalySeverityFilter: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("major"),
				types.StringValue("critical"),
			}),
			ConnectionFilter: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("connected"),
			}),
			SloStatusFilter: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("healthy"),
				types.StringValue("noData"),
			}),
			TimeRange: &models.TimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
				Mode: types.StringValue("relative"),
			},
		},
	}

	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap
	diags := apmservicemap.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := panel.Config
	require.NotNil(t, cfg.Title)
	assert.Equal(t, "APM Map", *cfg.Title)
	require.NotNil(t, cfg.Environment)
	assert.Equal(t, "production", *cfg.Environment)
	require.NotNil(t, cfg.ServiceName)
	assert.Equal(t, "checkout", *cfg.ServiceName)
	require.NotNil(t, cfg.ServiceGroupId)
	assert.Equal(t, "group-abc", *cfg.ServiceGroupId)
	require.NotNil(t, cfg.MapOrientation)
	assert.Equal(t, kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableMapOrientationHorizontal, *cfg.MapOrientation)
	require.NotNil(t, cfg.AlertStatusFilter)
	assert.ElementsMatch(t, []string{"active", "delayed"}, enumStrings(*cfg.AlertStatusFilter))
	require.NotNil(t, cfg.AnomalySeverityFilter)
	assert.ElementsMatch(t, []string{"major", "critical"}, enumStrings(*cfg.AnomalySeverityFilter))
	require.NotNil(t, cfg.ConnectionFilter)
	assert.Equal(t, []string{"connected"}, enumStrings(*cfg.ConnectionFilter))
	require.NotNil(t, cfg.SloStatusFilter)
	assert.ElementsMatch(t, []string{"healthy", "noData"}, enumStrings(*cfg.SloStatusFilter))
	require.NotNil(t, cfg.TimeRange)
	assert.Equal(t, "now-7d", cfg.TimeRange.From)
	assert.Equal(t, "now", cfg.TimeRange.To)
}

func TestBuildConfig_filterSets_multipleValues(t *testing.T) {
	pm := models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			AlertStatusFilter: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("recovered"),
				types.StringValue("untracked"),
			}),
			ConnectionFilter: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("connected"),
				types.StringValue("orphaned"),
			}),
		},
	}

	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap
	diags := apmservicemap.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, panel.Config.AlertStatusFilter)
	assert.ElementsMatch(t, []string{"recovered", "untracked"}, enumStrings(*panel.Config.AlertStatusFilter))
	require.NotNil(t, panel.Config.ConnectionFilter)
	assert.ElementsMatch(t, []string{"connected", "orphaned"}, enumStrings(*panel.Config.ConnectionFilter))
}

func TestPopulateFromAPI_import_emptyConfig_blockIsNull(t *testing.T) {
	pm := &models.PanelModel{}
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{}
	diags := apmservicemap.PopulateFromAPI(pm, nil, panel)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Nil(t, pm.ApmServiceMapConfig)
}

func TestPopulateFromAPI_import_withFields(t *testing.T) {
	pm := &models.PanelModel{}
	env := "staging"
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
		Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
			Environment: &env,
		},
	}
	diags := apmservicemap.PopulateFromAPI(pm, nil, panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, pm.ApmServiceMapConfig)
	assert.Equal(t, "staging", pm.ApmServiceMapConfig.Environment.ValueString())
}

func TestPopulateFromAPI_environmentServerDefault_readPath(t *testing.T) {
	t.Run("prior null API ENVIRONMENT_ALL stays null", func(t *testing.T) {
		pm := &models.PanelModel{
			ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
				Environment: types.StringNull(),
			},
		}
		prior := &models.PanelModel{
			ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
				Environment: types.StringNull(),
			},
		}

		env := "ENVIRONMENT_ALL"
		panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
			Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
				Environment: &env,
			},
		}
		diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
		require.False(t, diags.HasError(), "%v", diags)

		require.NotNil(t, pm.ApmServiceMapConfig)
		assert.True(t, pm.ApmServiceMapConfig.Environment.IsNull())
	})

	t.Run("prior production API ENVIRONMENT_ALL preserved", func(t *testing.T) {
		pm := &models.PanelModel{}
		prior := &models.PanelModel{
			ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
				Environment: types.StringValue("production"),
			},
		}

		env := "ENVIRONMENT_ALL"
		panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
			Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
				Environment: &env,
			},
		}
		diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
		require.False(t, diags.HasError(), "%v", diags)

		require.NotNil(t, pm.ApmServiceMapConfig)
		assert.Equal(t, "ENVIRONMENT_ALL", pm.ApmServiceMapConfig.Environment.ValueString())
	})

	t.Run("prior ENVIRONMENT_ALL explicit API same preserved", func(t *testing.T) {
		pm := &models.PanelModel{}
		prior := &models.PanelModel{
			ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
				Environment: types.StringValue("ENVIRONMENT_ALL"),
			},
		}

		env := "ENVIRONMENT_ALL"
		panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
			Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
				Environment: &env,
			},
		}
		diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
		require.False(t, diags.HasError(), "%v", diags)

		require.NotNil(t, pm.ApmServiceMapConfig)
		assert.Equal(t, "ENVIRONMENT_ALL", pm.ApmServiceMapConfig.Environment.ValueString())
	})
}

func TestPopulateFromAPI_environmentServerDefault_importPath(t *testing.T) {
	t.Run("API ENVIRONMENT_ALL imported as null", func(t *testing.T) {
		pm := &models.PanelModel{}
		env := "ENVIRONMENT_ALL"
		panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
			Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
				Environment: &env,
			},
		}
		diags := apmservicemap.PopulateFromAPI(pm, nil, panel)
		require.False(t, diags.HasError(), "%v", diags)

		require.NotNil(t, pm.ApmServiceMapConfig)
		assert.True(t, pm.ApmServiceMapConfig.Environment.IsNull())
	})

	t.Run("API nil environment imported as null", func(t *testing.T) {
		pm := &models.PanelModel{}
		serviceName := "checkout"
		panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
			Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
				ServiceName: &serviceName,
			},
		}
		diags := apmservicemap.PopulateFromAPI(pm, nil, panel)
		require.False(t, diags.HasError(), "%v", diags)

		require.NotNil(t, pm.ApmServiceMapConfig)
		assert.True(t, pm.ApmServiceMapConfig.Environment.IsNull())
	})

	t.Run("API production imported verbatim", func(t *testing.T) {
		pm := &models.PanelModel{}
		env := "production"
		panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
			Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
				Environment: &env,
			},
		}
		diags := apmservicemap.PopulateFromAPI(pm, nil, panel)
		require.False(t, diags.HasError(), "%v", diags)

		require.NotNil(t, pm.ApmServiceMapConfig)
		assert.Equal(t, "production", pm.ApmServiceMapConfig.Environment.ValueString())
	})
}

func TestPopulateFromAPI_nullPreservation_scalars(t *testing.T) {
	pm := &models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			Environment: types.StringNull(),
		},
	}
	prior := &models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			Environment: types.StringNull(),
		},
	}

	env := "production"
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
		Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
			Environment: &env,
		},
	}
	diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, pm.ApmServiceMapConfig)
	assert.True(t, pm.ApmServiceMapConfig.Environment.IsNull())
}

func TestPopulateFromAPI_filterSet_reordering(t *testing.T) {
	expectedFilter := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("active"),
		types.StringValue("delayed"),
	})
	// pm starts empty, matching the real dashboardMapPanelFromAPI flow: it is never
	// pre-populated with a value to merge against, only `prior` (the plan) carries intent.
	pm := &models.PanelModel{}
	prior := &models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			AlertStatusFilter: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("active"),
				types.StringValue("delayed"),
			}),
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
		Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
			AlertStatusFilter: &[]kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilter{
				kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilterDelayed,
				kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilterActive,
			},
		},
	}
	diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, pm.ApmServiceMapConfig)
	assert.True(t, pm.ApmServiceMapConfig.AlertStatusFilter.Equal(expectedFilter))
}

func TestPopulateFromAPI_filterSet_nullPreservation(t *testing.T) {
	// pm starts empty, matching the real dashboardMapPanelFromAPI flow.
	pm := &models.PanelModel{}
	prior := &models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			AlertStatusFilter: types.SetNull(types.StringType),
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
		Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
			AlertStatusFilter: &[]kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilter{
				kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilterActive,
			},
		},
	}
	diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, pm.ApmServiceMapConfig)
	assert.True(t, pm.ApmServiceMapConfig.AlertStatusFilter.IsNull(),
		"an unconfigured filter must stay null even if the API reports a value (e.g. out-of-band drift)")
}

// TestPopulateFromAPI_filterSet_emptySetPreservedAcrossRead is a regression test for a bug where a
// practitioner-configured empty set (e.g. `alert_status_filter = []`) drifted to null on every
// subsequent read. BuildConfig omits both null and empty sets from the API payload identically, so
// the API response alone cannot tell them apart; PopulateFromAPI must fall back to the prior
// (plan/state) set, not to a value derived only from the API, to preserve the distinction.
func TestPopulateFromAPI_filterSet_emptySetPreservedAcrossRead(t *testing.T) {
	// pm starts empty, matching the real dashboardMapPanelFromAPI flow.
	pm := &models.PanelModel{}
	prior := &models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			Environment:       types.StringValue("production"),
			AlertStatusFilter: types.SetValueMust(types.StringType, []attr.Value{}),
		},
	}

	env := "production"
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
		Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
			Environment: &env,
			// AlertStatusFilter omitted entirely, exactly as BuildConfig would send it for an
			// empty set (see stringSetToEnumSlice).
		},
	}
	diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, pm.ApmServiceMapConfig)
	assert.False(t, pm.ApmServiceMapConfig.AlertStatusFilter.IsNull(),
		"a known-empty filter set must not drift to null across a read")
	assert.False(t, pm.ApmServiceMapConfig.AlertStatusFilter.IsUnknown())
	elems := pm.ApmServiceMapConfig.AlertStatusFilter.Elements()
	assert.Empty(t, elems)
}

func TestPopulateFromAPI_timeRange_nullPreservation(t *testing.T) {
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	pm := &models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			Environment: types.StringNull(),
			TimeRange:   nil,
		},
	}
	prior := &models.PanelModel{
		ApmServiceMapConfig: &models.ApmServiceMapConfigModel{
			Environment: types.StringNull(),
			TimeRange:   nil,
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap{
		Config: kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable{
			TimeRange: &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
				From: "now-30d",
				To:   "now",
				Mode: &mode,
			},
		},
	}
	diags := apmservicemap.PopulateFromAPI(pm, prior, panel)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, pm.ApmServiceMapConfig)
	assert.Nil(t, pm.ApmServiceMapConfig.TimeRange)
}

func Test_mapOrientationValidator_rejectsInvalidValue(t *testing.T) {
	ctx := context.Background()
	v := stringvalidator.OneOf("horizontal", "vertical")
	var resp validator.StringResponse
	v.ValidateString(ctx, validator.StringRequest{
		ConfigValue: types.StringValue("diagonal"),
		Path:        path.Root("map_orientation"),
	}, &resp)
	require.True(t, resp.Diagnostics.HasError())
}

func Test_alertStatusFilterValidator_rejectsInvalidValue(t *testing.T) {
	ctx := context.Background()
	v := setvalidator.ValueStringsAre(
		stringvalidator.OneOf("active", "delayed", "recovered", "untracked"),
	)
	var resp validator.SetResponse
	v.ValidateSet(ctx, validator.SetRequest{
		ConfigValue: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("invalid_value"),
		}),
		Path: path.Root("alert_status_filter"),
	}, &resp)
	require.True(t, resp.Diagnostics.HasError())
}

func enumStrings[T ~string](vals []T) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = string(v)
	}
	return out
}
