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

package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// toAPIModel converts the PF model to the API request struct.
// It applies version gating via isSettingAllowed checks.
func toAPIModel(ctx context.Context, client *clients.ElasticsearchScopedClient, model tfModel) (*models.Transform, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	var transform models.Transform

	transform.Name = model.Name.ValueString()

	if typeutils.IsKnown(model.Description) {
		transform.Description = model.Description.ValueString()
	}

	// Source
	if model.Source != nil {
		src := model.Source
		transform.Source = &models.TransformSource{}

		indices := make([]string, 0, len(src.Indices))
		for _, idx := range src.Indices {
			indices = append(indices, idx.ValueString())
		}
		transform.Source.Indices = indices

		if typeutils.IsKnown(src.Query) && src.Query.ValueString() != "" {
			var query any
			if err := json.Unmarshal([]byte(src.Query.ValueString()), &query); err != nil {
				diags.AddError("Error parsing source.query", err.Error())
				return nil, diags
			}
			transform.Source.Query = query
		}

		if typeutils.IsKnown(src.RuntimeMappings) && src.RuntimeMappings.ValueString() != "" {
			allowed, allowDiags := isSettingAllowed(ctx, "source.runtime_mappings", client)
			diags.Append(allowDiags...)
			if allowDiags.HasError() {
				return nil, diags
			}
			if allowed {
				var rm any
				if err := json.Unmarshal([]byte(src.RuntimeMappings.ValueString()), &rm); err != nil {
					diags.AddError("Error parsing source.runtime_mappings", err.Error())
					return nil, diags
				}
				transform.Source.RuntimeMappings = rm
			}
		}
	}

	// Destination
	if model.Destination != nil {
		dst := model.Destination
		transform.Destination = &models.TransformDestination{
			Index: dst.Index.ValueString(),
		}

		if len(dst.Aliases) > 0 {
			allowed, allowDiags := isSettingAllowed(ctx, "destination.aliases", client)
			diags.Append(allowDiags...)
			if allowDiags.HasError() {
				return nil, diags
			}
			if allowed {
				transform.Destination.Aliases = make([]models.TransformAlias, len(dst.Aliases))
				for i, a := range dst.Aliases {
					transform.Destination.Aliases[i] = models.TransformAlias{
						Alias:          a.Alias.ValueString(),
						MoveOnCreation: a.MoveOnCreation.ValueBool(),
					}
				}
			}
		}

		if typeutils.IsKnown(dst.Pipeline) && dst.Pipeline.ValueString() != "" {
			transform.Destination.Pipeline = dst.Pipeline.ValueString()
		}
	}

	// Pivot
	if typeutils.IsKnown(model.Pivot) && model.Pivot.ValueString() != "" {
		var pivot any
		if err := json.Unmarshal([]byte(model.Pivot.ValueString()), &pivot); err != nil {
			diags.AddError("Error parsing pivot", err.Error())
			return nil, diags
		}
		transform.Pivot = pivot
	}

	// Latest
	if typeutils.IsKnown(model.Latest) && model.Latest.ValueString() != "" {
		var latest any
		if err := json.Unmarshal([]byte(model.Latest.ValueString()), &latest); err != nil {
			diags.AddError("Error parsing latest", err.Error())
			return nil, diags
		}
		transform.Latest = latest
	}

	// Frequency
	if typeutils.IsKnown(model.Frequency) && model.Frequency.ValueString() != "" {
		transform.Frequency = model.Frequency.ValueString()
	}

	// Metadata
	if meta := typeutils.NormalizedTypeToMap[any](model.Metadata, path.Root("metadata"), &diags); meta != nil {
		transform.Meta = meta
	}

	// RetentionPolicy
	if model.RetentionPolicy != nil && model.RetentionPolicy.Time != nil {
		t := model.RetentionPolicy.Time
		transform.RetentionPolicy = &models.TransformRetentionPolicy{
			Time: models.TransformRetentionPolicyTime{
				Field:  t.Field.ValueString(),
				MaxAge: t.MaxAge.ValueString(),
			},
		}
	}

	// Sync
	if model.Sync != nil && model.Sync.Time != nil {
		t := model.Sync.Time
		transform.Sync = &models.TransformSync{
			Time: models.TransformSyncTime{
				Field: t.Field.ValueString(),
				Delay: t.Delay.ValueString(),
			},
		}
	}

	// Each entry pairs the API setting name (used for version gating) with the
	// configured value. set==false means the user did not configure the field,
	// so we skip both the version check and the assignment (avoids spurious
	// "not allowed" warnings for unset fields).
	settings := models.TransformSettings{}
	setSettings := false
	applies := []struct {
		name  string
		set   bool
		write func()
	}{
		{name: "align_checkpoints", set: typeutils.IsKnown(model.AlignCheckpoints), write: func() { v := model.AlignCheckpoints.ValueBool(); settings.AlignCheckpoints = &v }},
		{name: "dates_as_epoch_millis", set: typeutils.IsKnown(model.DatesAsEpochMillis), write: func() { v := model.DatesAsEpochMillis.ValueBool(); settings.DatesAsEpochMillis = &v }},
		{name: settingDeduceMappings, set: typeutils.IsKnown(model.DeduceMappings), write: func() { v := model.DeduceMappings.ValueBool(); settings.DeduceMappings = &v }},
		{name: "docs_per_second", set: typeutils.IsKnown(model.DocsPerSecond), write: func() { v := model.DocsPerSecond.ValueFloat64(); settings.DocsPerSecond = &v }},
		{name: "max_page_search_size", set: typeutils.IsKnown(model.MaxPageSearchSize), write: func() { settings.MaxPageSearchSize = typeutils.OptionalInt(model.MaxPageSearchSize) }},
		{name: settingNumFailureRetries, set: typeutils.IsKnown(model.NumFailureRetries), write: func() { settings.NumFailureRetries = typeutils.OptionalInt(model.NumFailureRetries) }},
		{name: settingUnattended, set: typeutils.IsKnown(model.Unattended), write: func() { v := model.Unattended.ValueBool(); settings.Unattended = &v }},
	}

	for _, s := range applies {
		if !s.set {
			continue
		}
		allowed, allowDiags := isSettingAllowed(ctx, s.name, client)
		diags.Append(allowDiags...)
		if allowDiags.HasError() {
			return nil, diags
		}
		if !allowed {
			continue
		}
		s.write()
		setSettings = true
	}

	if setSettings {
		transform.Settings = &settings
	}

	return &transform, diags
}

// fromAPIModel populates the PF model from a Get Transform response and Get Transform Stats.
func fromAPIModel(ctx context.Context, transform *models.Transform, stats *types.TransformStats, state tfModel) (tfModel, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	model := state

	// Name — always set from the API response to ensure it is populated on import reads.
	model.Name = basetypes.NewStringValue(transform.Name)

	// Description
	if transform.Description != "" {
		model.Description = basetypes.NewStringValue(transform.Description)
	} else {
		model.Description = basetypes.NewStringNull()
	}

	// Source
	if transform.Source != nil {
		src := tfModelSource{}

		indices := make([]basetypes.StringValue, 0, len(transform.Source.Indices))
		for _, idx := range transform.Source.Indices {
			indices = append(indices, basetypes.NewStringValue(idx))
		}
		src.Indices = indices

		src.Query = typeutils.MarshalToNormalized(transform.Source.Query, path.Root("source").AtName("query"), &diags)
		if diags.HasError() {
			return model, diags
		}

		src.RuntimeMappings = typeutils.MarshalToNormalized(transform.Source.RuntimeMappings, path.Root("source").AtName("runtime_mappings"), &diags)
		if diags.HasError() {
			return model, diags
		}

		model.Source = &src
	} else {
		model.Source = nil
	}

	// Destination
	if transform.Destination != nil {
		dst := tfModelDestination{
			Index: basetypes.NewStringValue(transform.Destination.Index),
		}

		if transform.Destination.Pipeline != "" {
			dst.Pipeline = basetypes.NewStringValue(transform.Destination.Pipeline)
		} else {
			dst.Pipeline = basetypes.NewStringNull()
		}

		if len(transform.Destination.Aliases) > 0 {
			aliases := make([]tfModelAlias, len(transform.Destination.Aliases))
			for i, a := range transform.Destination.Aliases {
				aliases[i] = tfModelAlias{
					Alias:          basetypes.NewStringValue(a.Alias),
					MoveOnCreation: basetypes.NewBoolValue(a.MoveOnCreation),
				}
			}
			dst.Aliases = aliases
		} else {
			dst.Aliases = nil
		}

		model.Destination = &dst
	} else {
		model.Destination = nil
	}

	// Pivot
	model.Pivot = typeutils.MarshalToNormalized(transform.Pivot, path.Root("pivot"), &diags)
	if diags.HasError() {
		return model, diags
	}

	// Latest
	model.Latest = typeutils.MarshalToNormalized(transform.Latest, path.Root("latest"), &diags)
	if diags.HasError() {
		return model, diags
	}

	// Frequency
	model.Frequency = basetypes.NewStringValue(transform.Frequency)

	// Sync
	if transform.Sync != nil {
		syncTime := tfModelSyncTime{
			Field: basetypes.NewStringValue(transform.Sync.Time.Field),
			Delay: basetypes.NewStringValue(transform.Sync.Time.Delay),
		}
		model.Sync = &tfModelSync{Time: &syncTime}
	} else {
		model.Sync = nil
	}

	// RetentionPolicy
	if transform.RetentionPolicy != nil {
		retTime := tfModelRetentionTime{
			Field:  basetypes.NewStringValue(transform.RetentionPolicy.Time.Field),
			MaxAge: basetypes.NewStringValue(transform.RetentionPolicy.Time.MaxAge),
		}
		model.RetentionPolicy = &tfModelRetention{Time: &retTime}
	} else {
		model.RetentionPolicy = nil
	}

	// Settings
	if transform.Settings != nil {
		if transform.Settings.AlignCheckpoints != nil {
			model.AlignCheckpoints = basetypes.NewBoolValue(*transform.Settings.AlignCheckpoints)
		}
		if transform.Settings.DatesAsEpochMillis != nil {
			model.DatesAsEpochMillis = basetypes.NewBoolValue(*transform.Settings.DatesAsEpochMillis)
		}
		if transform.Settings.DeduceMappings != nil {
			model.DeduceMappings = basetypes.NewBoolValue(*transform.Settings.DeduceMappings)
		}
		if transform.Settings.DocsPerSecond != nil {
			model.DocsPerSecond = basetypes.NewFloat64Value(*transform.Settings.DocsPerSecond)
		}
		if transform.Settings.MaxPageSearchSize != nil {
			model.MaxPageSearchSize = basetypes.NewInt64Value(int64(*transform.Settings.MaxPageSearchSize))
		}
		if transform.Settings.NumFailureRetries != nil {
			model.NumFailureRetries = basetypes.NewInt64Value(int64(*transform.Settings.NumFailureRetries))
		}
		if transform.Settings.Unattended != nil {
			model.Unattended = basetypes.NewBoolValue(*transform.Settings.Unattended)
		}
	}

	// Metadata
	model.Metadata = typeutils.MarshalToNormalized(transform.Meta, path.Root("metadata"), &diags)
	if diags.HasError() {
		return model, diags
	}

	// Enabled: derived from transform stats
	if stats != nil {
		isStarted := stats.State == "started" || stats.State == "indexing"
		model.Enabled = basetypes.NewBoolValue(isStarted)
	} else {
		tflog.Warn(ctx, fmt.Sprintf("Transform stats not available for %s; leaving enabled state unchanged", transform.Name))
	}

	return model, diags
}
