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

package connector

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestVersionGate_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	reqs, diags := VersionGate{}.GetVersionRequirements()
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].MinVersion.String() != MinSupportedVersion.String() {
		t.Errorf("MinVersion = %s, want %s", reqs[0].MinVersion.String(), MinSupportedVersion.String())
	}
	if reqs[0].ErrorMessage == "" {
		t.Error("ErrorMessage must not be empty")
	}
}

func TestNestedModelAttrTypes_objectRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("pipeline", func(t *testing.T) {
		t.Parallel()
		roundTripObject(ctx, t, PipelineModelAttrTypes(), PipelineModel{
			Name:                 fwtypes.StringValue("search-default-ingestion"),
			ExtractBinaryContent: fwtypes.BoolValue(true),
			ReduceWhitespace:     fwtypes.BoolValue(true),
			RunMlInference:       fwtypes.BoolValue(false),
		})
	})

	t.Run("schedule_entry", func(t *testing.T) {
		t.Parallel()
		roundTripObject(ctx, t, ScheduleEntryModelAttrTypes(), ScheduleEntryModel{
			Enabled:  fwtypes.BoolValue(true),
			Interval: fwtypes.StringValue("0 0 0 * * ?"),
		})
	})

	t.Run("scheduling", func(t *testing.T) {
		t.Parallel()
		full, diags := fwtypes.ObjectValueFrom(ctx, ScheduleEntryModelAttrTypes(), ScheduleEntryModel{
			Enabled:  fwtypes.BoolValue(true),
			Interval: fwtypes.StringValue("0 0 0 * * ?"),
		})
		if diags.HasError() {
			t.Fatalf("building full schedule entry: %v", diags)
		}
		roundTripObject(ctx, t, SchedulingModelAttrTypes(), SchedulingModel{
			Full:          full,
			Incremental:   fwtypes.ObjectNull(ScheduleEntryModelAttrTypes()),
			AccessControl: fwtypes.ObjectNull(ScheduleEntryModelAttrTypes()),
		})
	})

	t.Run("feature_flag", func(t *testing.T) {
		t.Parallel()
		roundTripObject(ctx, t, FeatureFlagModelAttrTypes(), FeatureFlagModel{
			Enabled: fwtypes.BoolValue(true),
		})
	})

	t.Run("sync_rules", func(t *testing.T) {
		t.Parallel()
		basic, diags := fwtypes.ObjectValueFrom(ctx, FeatureFlagModelAttrTypes(), FeatureFlagModel{
			Enabled: fwtypes.BoolValue(true),
		})
		if diags.HasError() {
			t.Fatalf("building basic sync rules flag: %v", diags)
		}
		roundTripObject(ctx, t, SyncRulesModelAttrTypes(), SyncRulesModel{
			Basic:    basic,
			Advanced: fwtypes.ObjectNull(FeatureFlagModelAttrTypes()),
		})
	})

	t.Run("features", func(t *testing.T) {
		t.Parallel()
		dls, diags := fwtypes.ObjectValueFrom(ctx, FeatureFlagModelAttrTypes(), FeatureFlagModel{
			Enabled: fwtypes.BoolValue(false),
		})
		if diags.HasError() {
			t.Fatalf("building document_level_security flag: %v", diags)
		}
		roundTripObject(ctx, t, FeaturesModelAttrTypes(), FeaturesModel{
			DocumentLevelSecurity:  dls,
			IncrementalSync:        fwtypes.ObjectNull(FeatureFlagModelAttrTypes()),
			NativeConnectorAPIKeys: fwtypes.ObjectNull(FeatureFlagModelAttrTypes()),
			SyncRules:              fwtypes.ObjectNull(SyncRulesModelAttrTypes()),
		})
	})
}

func roundTripObject[T any](ctx context.Context, t *testing.T, attrTypes map[string]attr.Type, model T) {
	t.Helper()

	obj, diags := fwtypes.ObjectValueFrom(ctx, attrTypes, model)
	if diags.HasError() {
		t.Fatalf("ObjectValueFrom: %v", diags)
	}

	var decoded T
	diags = obj.As(ctx, &decoded, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("As: %v", diags)
	}
}
