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

package ilm

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	RolloverMinConditionsMinSupportedVersion = version.Must(version.NewVersion("8.4.0"))
	MaxPrimaryShardDocsMinSupportedVersion   = version.Must(version.NewVersion("8.2.0"))
)

const (
	phaseDelete    = "delete"
	actionFreeze   = "freeze"
	actionReadonly = "readonly"
	actionUnfollow = "unfollow"
)

var ilmActionSettingOptions = map[string]struct {
	skipEmptyCheck bool
	def            any
	minVersion     *version.Version
}{
	"allow_write_after_shrink": {def: false, minVersion: version.Must(version.NewVersion("8.14.0"))},
	"number_of_replicas":       {skipEmptyCheck: true},
	"priority":                 {skipEmptyCheck: true},
	"max_primary_shard_docs":   {def: int64(0), minVersion: MaxPrimaryShardDocsMinSupportedVersion},
	"min_age":                  {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_docs":                 {def: int64(0), minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_size":                 {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_primary_shard_docs":   {def: int64(0), minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_primary_shard_size":   {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	"total_shards_per_node":    {skipEmptyCheck: true, def: int64(-1), minVersion: version.Must(version.NewVersion("7.16.0"))},
}

type ilmModel struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	Hot                     types.List           `tfsdk:"hot"`
	Warm                    types.List           `tfsdk:"warm"`
	Cold                    types.List           `tfsdk:"cold"`
	Frozen                  types.List           `tfsdk:"frozen"`
	Delete                  types.List           `tfsdk:"delete"`
	ModifiedDate            types.String         `tfsdk:"modified_date"`
}

type phaseModel struct {
	MinAge             types.String `tfsdk:"min_age"`
	Allocate           types.List   `tfsdk:"allocate"`
	Delete             types.List   `tfsdk:"delete"`
	Forcemerge         types.List   `tfsdk:"forcemerge"`
	Freeze             types.List   `tfsdk:"freeze"`
	Migrate            types.List   `tfsdk:"migrate"`
	Readonly           types.List   `tfsdk:"readonly"`
	Rollover           types.List   `tfsdk:"rollover"`
	SearchableSnapshot types.List   `tfsdk:"searchable_snapshot"`
	SetPriority        types.List   `tfsdk:"set_priority"`
	Shrink             types.List   `tfsdk:"shrink"`
	Unfollow           types.List   `tfsdk:"unfollow"`
	WaitForSnapshot    types.List   `tfsdk:"wait_for_snapshot"`
	Downsample         types.List   `tfsdk:"downsample"`
}

type allocateActionModel struct {
	NumberOfReplicas   types.Int64          `tfsdk:"number_of_replicas"`
	TotalShardsPerNode types.Int64          `tfsdk:"total_shards_per_node"`
	Include            jsontypes.Normalized `tfsdk:"include"`
	Exclude            jsontypes.Normalized `tfsdk:"exclude"`
	Require            jsontypes.Normalized `tfsdk:"require"`
}

type deleteActionModel struct {
	DeleteSearchableSnapshot types.Bool `tfsdk:"delete_searchable_snapshot"`
}

type forcemergeActionModel struct {
	MaxNumSegments types.Int64  `tfsdk:"max_num_segments"`
	IndexCodec     types.String `tfsdk:"index_codec"`
}

type enabledActionModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type rolloverActionModel struct {
	MaxAge              types.String `tfsdk:"max_age"`
	MaxDocs             types.Int64  `tfsdk:"max_docs"`
	MaxSize             types.String `tfsdk:"max_size"`
	MaxPrimaryShardDocs types.Int64  `tfsdk:"max_primary_shard_docs"`
	MaxPrimaryShardSize types.String `tfsdk:"max_primary_shard_size"`
	MinAge              types.String `tfsdk:"min_age"`
	MinDocs             types.Int64  `tfsdk:"min_docs"`
	MinSize             types.String `tfsdk:"min_size"`
	MinPrimaryShardDocs types.Int64  `tfsdk:"min_primary_shard_docs"`
	MinPrimaryShardSize types.String `tfsdk:"min_primary_shard_size"`
}

type searchableSnapshotActionModel struct {
	SnapshotRepository types.String `tfsdk:"snapshot_repository"`
	ForceMergeIndex    types.Bool   `tfsdk:"force_merge_index"`
}

type setPriorityActionModel struct {
	Priority types.Int64 `tfsdk:"priority"`
}

type shrinkActionModel struct {
	NumberOfShards      types.Int64  `tfsdk:"number_of_shards"`
	MaxPrimaryShardSize types.String `tfsdk:"max_primary_shard_size"`
	AllowWriteAfter     types.Bool   `tfsdk:"allow_write_after_shrink"`
}

type waitForSnapshotActionModel struct {
	Policy types.String `tfsdk:"policy"`
}

type downsampleActionModel struct {
	FixedInterval types.String `tfsdk:"fixed_interval"`
	WaitTimeout   types.String `tfsdk:"wait_timeout"`
}

func newNullModel() ilmModel {
	return ilmModel{
		Metadata: jsontypes.NewNormalizedNull(),
		Hot:      types.ListNull(phaseElementType()),
		Warm:     types.ListNull(phaseElementType()),
		Cold:     types.ListNull(phaseElementType()),
		Frozen:   types.ListNull(phaseElementType()),
		Delete:   types.ListNull(phaseElementType()),
	}
}

func newPhaseModel() phaseModel {
	return phaseModel{
		MinAge:             types.StringNull(),
		Allocate:           types.ListNull(allocateActionElementType()),
		Delete:             types.ListNull(deleteActionElementType()),
		Forcemerge:         types.ListNull(forcemergeActionElementType()),
		Freeze:             types.ListNull(enabledActionElementType()),
		Migrate:            types.ListNull(enabledActionElementType()),
		Readonly:           types.ListNull(enabledActionElementType()),
		Rollover:           types.ListNull(rolloverActionElementType()),
		SearchableSnapshot: types.ListNull(searchableSnapshotActionElementType()),
		SetPriority:        types.ListNull(setPriorityActionElementType()),
		Shrink:             types.ListNull(shrinkActionElementType()),
		Unfollow:           types.ListNull(enabledActionElementType()),
		WaitForSnapshot:    types.ListNull(waitForSnapshotActionElementType()),
		Downsample:         types.ListNull(downsampleActionElementType()),
	}
}

func (m ilmModel) toPolicy(ctx context.Context, serverVersion *version.Version) (*models.Policy, diag.Diagnostics) {
	policy := &models.Policy{
		Name:   m.Name.ValueString(),
		Phases: make(map[string]models.Phase),
	}

	if !m.Metadata.IsNull() && !m.Metadata.IsUnknown() {
		metadata := make(map[string]any)
		diags := m.Metadata.Unmarshal(&metadata)
		if diags.HasError() {
			return nil, diags
		}
		policy.Metadata = metadata
	}

	for _, phaseName := range []string{"hot", "warm", "cold", "frozen", phaseDelete} {
		phase, exists, diags := expandPhaseFromState(ctx, phaseName, m.phaseByName(phaseName), serverVersion)
		if diags.HasError() {
			return nil, diags
		}
		if exists {
			policy.Phases[phaseName] = phase
		}
	}

	return policy, nil
}

func (m *ilmModel) setPhase(name string, value types.List) {
	switch name {
	case "hot":
		m.Hot = value
	case "warm":
		m.Warm = value
	case "cold":
		m.Cold = value
	case "frozen":
		m.Frozen = value
	case phaseDelete:
		m.Delete = value
	}
}

func (m ilmModel) phaseByName(name string) types.List {
	switch name {
	case "hot":
		return m.Hot
	case "warm":
		return m.Warm
	case "cold":
		return m.Cold
	case "frozen":
		return m.Frozen
	default:
		return m.Delete
	}
}

func expandPhaseFromState(ctx context.Context, phaseName string, phaseList types.List, serverVersion *version.Version) (models.Phase, bool, diag.Diagnostics) {
	if phaseList.IsNull() || phaseList.IsUnknown() {
		return models.Phase{}, false, nil
	}

	var phases []phaseModel
	diags := phaseList.ElementsAs(ctx, &phases, false)
	if diags.HasError() {
		return models.Phase{}, false, diags
	}
	if len(phases) == 0 {
		return models.Phase{}, false, nil
	}

	phase := models.Phase{Actions: make(map[string]models.Action)}
	p := phases[0]
	if !p.MinAge.IsNull() && !p.MinAge.IsUnknown() && p.MinAge.ValueString() != "" {
		phase.MinAge = p.MinAge.ValueString()
	}

	if action, exists, d := expandAllocate(ctx, p.Allocate, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["allocate"] = action
	}
	if action, exists, d := expandDelete(ctx, p.Delete, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["delete"] = action
	}
	if action, exists, d := expandForcemerge(ctx, p.Forcemerge, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["forcemerge"] = action
	}
	if action, exists, d := expandEnabled(ctx, p.Freeze, false, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["freeze"] = action
	}
	if action, exists, d := expandEnabled(ctx, p.Migrate, true, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["migrate"] = action
	}
	if action, exists, d := expandReadonly(ctx, p.Readonly, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions[actionReadonly] = action
	}
	if action, exists, d := expandRollover(ctx, p.Rollover, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["rollover"] = action
	}
	if action, exists, d := expandSearchableSnapshot(ctx, p.SearchableSnapshot, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["searchable_snapshot"] = action
	}
	if action, exists, d := expandSetPriority(ctx, p.SetPriority, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["set_priority"] = action
	}
	if action, exists, d := expandShrink(ctx, p.Shrink, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["shrink"] = action
	}
	if action, exists, d := expandEnabled(ctx, p.Unfollow, false, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions[actionUnfollow] = action
	}
	if action, exists, d := expandWaitForSnapshot(ctx, p.WaitForSnapshot, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["wait_for_snapshot"] = action
	}
	if action, exists, d := expandDownsample(ctx, p.Downsample, serverVersion); d.HasError() {
		return models.Phase{}, false, d
	} else if exists {
		phase.Actions["downsample"] = action
	}

	// Keep previous SDK behavior: unknown actions in a phase are rejected.
	for action := range phase.Actions {
		switch action {
		case "allocate", phaseDelete, "forcemerge", actionFreeze, "migrate", actionReadonly, "rollover", "searchable_snapshot", "set_priority", "shrink", actionUnfollow, "wait_for_snapshot", "downsample":
		default:
			var d diag.Diagnostics
			d.AddError("Unknown action defined.", fmt.Sprintf(`Configured action "%s" is not supported in phase "%s"`, action, phaseName))
			return models.Phase{}, false, d
		}
	}

	return phase, true, nil
}

func expandAllocate(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[allocateActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "number_of_replicas", action.NumberOfReplicas.ValueInt64(), !action.NumberOfReplicas.IsNull() && !action.NumberOfReplicas.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(
		out,
		"total_shards_per_node",
		action.TotalShardsPerNode.ValueInt64(),
		!action.TotalShardsPerNode.IsNull() && !action.TotalShardsPerNode.IsUnknown(),
		serverVersion,
	); d.HasError() {
		return nil, false, d
	}
	if d := applyJSONSetting(out, "include", action.Include, serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applyJSONSetting(out, "exclude", action.Exclude, serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applyJSONSetting(out, "require", action.Require, serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandDelete(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[deleteActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(
		out,
		"delete_searchable_snapshot",
		action.DeleteSearchableSnapshot.ValueBool(),
		!action.DeleteSearchableSnapshot.IsNull() && !action.DeleteSearchableSnapshot.IsUnknown(),
		serverVersion,
	); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandForcemerge(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[forcemergeActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "max_num_segments", action.MaxNumSegments.ValueInt64(), !action.MaxNumSegments.IsNull() && !action.MaxNumSegments.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "index_codec", action.IndexCodec.ValueString(), !action.IndexCodec.IsNull() && !action.IndexCodec.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandEnabled(ctx context.Context, list types.List, includeDisabled bool, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[enabledActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	enabled := !action.Enabled.IsNull() && !action.Enabled.IsUnknown() && action.Enabled.ValueBool()
	if !enabled && !includeDisabled {
		return nil, false, nil
	}
	out := map[string]any{}
	if d := applySetting(out, "enabled", action.Enabled.ValueBool(), !action.Enabled.IsNull() && !action.Enabled.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

// expandReadonly returns {} for the readonly action. The Elasticsearch API does not
// support any options (including "enabled") for readonly—it expects an empty object.
func expandReadonly(ctx context.Context, list types.List, _ *version.Version) (map[string]any, bool, diag.Diagnostics) {
	_, exists, diags := getSingle[enabledActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	return map[string]any{}, true, nil
}

func expandRollover(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[rolloverActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "max_age", action.MaxAge.ValueString(), !action.MaxAge.IsNull() && !action.MaxAge.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "max_docs", action.MaxDocs.ValueInt64(), !action.MaxDocs.IsNull() && !action.MaxDocs.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "max_size", action.MaxSize.ValueString(), !action.MaxSize.IsNull() && !action.MaxSize.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(
		out,
		"max_primary_shard_docs",
		action.MaxPrimaryShardDocs.ValueInt64(),
		!action.MaxPrimaryShardDocs.IsNull() && !action.MaxPrimaryShardDocs.IsUnknown(),
		serverVersion,
	); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(
		out,
		"max_primary_shard_size",
		action.MaxPrimaryShardSize.ValueString(),
		!action.MaxPrimaryShardSize.IsNull() && !action.MaxPrimaryShardSize.IsUnknown(),
		serverVersion,
	); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "min_age", action.MinAge.ValueString(), !action.MinAge.IsNull() && !action.MinAge.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "min_docs", action.MinDocs.ValueInt64(), !action.MinDocs.IsNull() && !action.MinDocs.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "min_size", action.MinSize.ValueString(), !action.MinSize.IsNull() && !action.MinSize.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(
		out,
		"min_primary_shard_docs",
		action.MinPrimaryShardDocs.ValueInt64(),
		!action.MinPrimaryShardDocs.IsNull() && !action.MinPrimaryShardDocs.IsUnknown(),
		serverVersion,
	); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(
		out,
		"min_primary_shard_size",
		action.MinPrimaryShardSize.ValueString(),
		!action.MinPrimaryShardSize.IsNull() && !action.MinPrimaryShardSize.IsUnknown(),
		serverVersion,
	); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandSearchableSnapshot(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[searchableSnapshotActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "snapshot_repository", action.SnapshotRepository.ValueString(), !action.SnapshotRepository.IsNull() && !action.SnapshotRepository.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "force_merge_index", action.ForceMergeIndex.ValueBool(), !action.ForceMergeIndex.IsNull() && !action.ForceMergeIndex.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandSetPriority(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[setPriorityActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "priority", action.Priority.ValueInt64(), !action.Priority.IsNull() && !action.Priority.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandShrink(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[shrinkActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "number_of_shards", action.NumberOfShards.ValueInt64(), !action.NumberOfShards.IsNull() && !action.NumberOfShards.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(
		out,
		"max_primary_shard_size",
		action.MaxPrimaryShardSize.ValueString(),
		!action.MaxPrimaryShardSize.IsNull() && !action.MaxPrimaryShardSize.IsUnknown(),
		serverVersion,
	); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "allow_write_after_shrink", action.AllowWriteAfter.ValueBool(), !action.AllowWriteAfter.IsNull() && !action.AllowWriteAfter.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandWaitForSnapshot(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[waitForSnapshotActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "policy", action.Policy.ValueString(), !action.Policy.IsNull() && !action.Policy.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func expandDownsample(ctx context.Context, list types.List, serverVersion *version.Version) (map[string]any, bool, diag.Diagnostics) {
	action, exists, diags := getSingle[downsampleActionModel](ctx, list)
	if !exists || diags.HasError() {
		return nil, exists, diags
	}
	out := map[string]any{}
	if d := applySetting(out, "fixed_interval", action.FixedInterval.ValueString(), !action.FixedInterval.IsNull() && !action.FixedInterval.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	if d := applySetting(out, "wait_timeout", action.WaitTimeout.ValueString(), !action.WaitTimeout.IsNull() && !action.WaitTimeout.IsUnknown(), serverVersion); d.HasError() {
		return nil, false, d
	}
	return out, true, nil
}

func applyJSONSetting(out map[string]any, key string, value jsontypes.Normalized, serverVersion *version.Version) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return applySetting(out, key, value.ValueString(), true, serverVersion)
}

func applySetting(out map[string]any, key string, value any, exists bool, serverVersion *version.Version) diag.Diagnostics {
	if !exists {
		return nil
	}

	options := ilmActionSettingOptions[key]
	if options.minVersion != nil && options.minVersion.GreaterThan(serverVersion) {
		if !reflect.DeepEqual(value, options.def) {
			var d diag.Diagnostics
			d.AddError(
				"Unsupported ILM setting for Elasticsearch version",
				fmt.Sprintf("[%s] is not supported in the target Elasticsearch server. Remove the setting from your module definition or set it to the default [%v] value", key, options.def),
			)
			return d
		}
		return nil
	}

	if options.skipEmptyCheck || !schemautil.IsEmpty(value) {
		if key == "include" || key == "exclude" || key == "require" {
			res := make(map[string]any)
			if err := json.Unmarshal([]byte(value.(string)), &res); err != nil {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic("Unable to decode allocate JSON setting", err.Error()),
				}
			}
			out[key] = res
		} else {
			out[key] = value
		}
	}
	return nil
}

func flattenPhase(ctx context.Context, _ string, p models.Phase, prior *phaseModel) (types.List, diag.Diagnostics) {
	phase := newPhaseModel()
	if p.MinAge != "" {
		phase.MinAge = types.StringValue(p.MinAge)
	}

	if prior != nil {
		for _, actionName := range []string{actionReadonly, actionFreeze, actionUnfollow} {
			if actionConfigured(ctx, prior, actionName) {
				list, diags := listWithSingle(ctx, enabledActionElementType(), enabledActionModel{Enabled: types.BoolValue(false)})
				if diags.HasError() {
					return types.ListNull(phaseElementType()), diags
				}
				setPhaseAction(&phase, actionName, list)
			}
		}
	}

	for actionName, action := range p.Actions {
		switch actionName {
		case actionReadonly, actionFreeze, actionUnfollow:
			list, diags := listWithSingle(ctx, enabledActionElementType(), enabledActionModel{Enabled: types.BoolValue(true)})
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			setPhaseAction(&phase, actionName, list)
		case "allocate":
			a := allocateActionModel{
				NumberOfReplicas:   types.Int64Null(),
				TotalShardsPerNode: types.Int64Value(-1),
				Include:            jsontypes.NewNormalizedNull(),
				Exclude:            jsontypes.NewNormalizedNull(),
				Require:            jsontypes.NewNormalizedNull(),
			}
			if v, ok := action["number_of_replicas"]; ok {
				a.NumberOfReplicas = types.Int64Value(int64FromAny(v))
			}
			if v, ok := action["total_shards_per_node"]; ok {
				a.TotalShardsPerNode = types.Int64Value(int64FromAny(v))
			}
			for _, field := range []string{"include", "exclude", "require"} {
				if v, ok := action[field]; ok {
					b, err := json.Marshal(v)
					if err != nil {
						return types.ListNull(phaseElementType()), diag.Diagnostics{
							diag.NewErrorDiagnostic("Unable to marshal allocate action JSON", err.Error()),
						}
					}
					switch field {
					case "include":
						a.Include = jsontypes.NewNormalizedValue(string(b))
					case "exclude":
						a.Exclude = jsontypes.NewNormalizedValue(string(b))
					case "require":
						a.Require = jsontypes.NewNormalizedValue(string(b))
					}
				}
			}
			list, diags := listWithSingle(ctx, allocateActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.Allocate = list
		case phaseDelete:
			a := deleteActionModel{DeleteSearchableSnapshot: types.BoolNull()}
			if v, ok := action["delete_searchable_snapshot"]; ok {
				a.DeleteSearchableSnapshot = types.BoolValue(boolFromAny(v))
			}
			list, diags := listWithSingle(ctx, deleteActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.Delete = list
		case "forcemerge":
			a := forcemergeActionModel{MaxNumSegments: types.Int64Null(), IndexCodec: types.StringNull()}
			if v, ok := action["max_num_segments"]; ok {
				a.MaxNumSegments = types.Int64Value(int64FromAny(v))
			}
			if v, ok := action["index_codec"]; ok {
				a.IndexCodec = types.StringValue(stringFromAny(v))
			}
			list, diags := listWithSingle(ctx, forcemergeActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.Forcemerge = list
		case "migrate":
			a := enabledActionModel{Enabled: types.BoolNull()}
			if v, ok := action["enabled"]; ok {
				a.Enabled = types.BoolValue(boolFromAny(v))
			}
			list, diags := listWithSingle(ctx, enabledActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.Migrate = list
		case "rollover":
			a := rolloverActionModel{
				MaxAge:              stringValueFromMap(action, "max_age"),
				MaxDocs:             int64ValueFromMap(action, "max_docs"),
				MaxSize:             stringValueFromMap(action, "max_size"),
				MaxPrimaryShardDocs: int64ValueFromMap(action, "max_primary_shard_docs"),
				MaxPrimaryShardSize: stringValueFromMap(action, "max_primary_shard_size"),
				MinAge:              stringValueFromMap(action, "min_age"),
				MinDocs:             int64ValueFromMap(action, "min_docs"),
				MinSize:             stringValueFromMap(action, "min_size"),
				MinPrimaryShardDocs: int64ValueFromMap(action, "min_primary_shard_docs"),
				MinPrimaryShardSize: stringValueFromMap(action, "min_primary_shard_size"),
			}
			list, diags := listWithSingle(ctx, rolloverActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.Rollover = list
		case "searchable_snapshot":
			a := searchableSnapshotActionModel{
				SnapshotRepository: stringValueFromMap(action, "snapshot_repository"),
				ForceMergeIndex:    boolValueFromMap(action, "force_merge_index"),
			}
			list, diags := listWithSingle(ctx, searchableSnapshotActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.SearchableSnapshot = list
		case "set_priority":
			a := setPriorityActionModel{Priority: int64ValueFromMap(action, "priority")}
			list, diags := listWithSingle(ctx, setPriorityActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.SetPriority = list
		case "shrink":
			a := shrinkActionModel{
				NumberOfShards:      int64ValueFromMap(action, "number_of_shards"),
				MaxPrimaryShardSize: stringValueFromMap(action, "max_primary_shard_size"),
				AllowWriteAfter:     boolValueFromMap(action, "allow_write_after_shrink"),
			}
			list, diags := listWithSingle(ctx, shrinkActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.Shrink = list
		case "wait_for_snapshot":
			a := waitForSnapshotActionModel{Policy: stringValueFromMap(action, "policy")}
			list, diags := listWithSingle(ctx, waitForSnapshotActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.WaitForSnapshot = list
		case "downsample":
			a := downsampleActionModel{
				FixedInterval: stringValueFromMap(action, "fixed_interval"),
				WaitTimeout:   stringValueFromMap(action, "wait_timeout"),
			}
			list, diags := listWithSingle(ctx, downsampleActionElementType(), a)
			if diags.HasError() {
				return types.ListNull(phaseElementType()), diags
			}
			phase.Downsample = list
		}
	}

	phaseList, diags := listWithSingle(ctx, phaseElementType(), phase)
	if diags.HasError() {
		return types.ListNull(phaseElementType()), diags
	}
	return phaseList, nil
}

func setPhaseAction(p *phaseModel, action string, list types.List) {
	switch action {
	case actionReadonly:
		p.Readonly = list
	case actionFreeze:
		p.Freeze = list
	case actionUnfollow:
		p.Unfollow = list
	}
}

func actionConfigured(_ context.Context, prior *phaseModel, action string) bool {
	var list types.List
	switch action {
	case actionReadonly:
		list = prior.Readonly
	case actionFreeze:
		list = prior.Freeze
	case actionUnfollow:
		list = prior.Unfollow
	default:
		return false
	}
	if list.IsNull() || list.IsUnknown() {
		return false
	}
	return len(list.Elements()) > 0
}

func getSingle[T any](ctx context.Context, list types.List) (*T, bool, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, false, nil
	}
	var values []T
	diags := list.ElementsAs(ctx, &values, false)
	if diags.HasError() {
		return nil, false, diags
	}
	if len(values) == 0 {
		return nil, false, nil
	}
	return &values[0], true, nil
}

func listWithSingle[T any](ctx context.Context, elementType attr.Type, value T) (types.List, diag.Diagnostics) {
	return types.ListValueFrom(ctx, elementType, []T{value})
}

func phaseElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"min_age":             types.StringType,
			"allocate":            types.ListType{ElemType: allocateActionElementType()},
			"delete":              types.ListType{ElemType: deleteActionElementType()},
			"forcemerge":          types.ListType{ElemType: forcemergeActionElementType()},
			"freeze":              types.ListType{ElemType: enabledActionElementType()},
			"migrate":             types.ListType{ElemType: enabledActionElementType()},
			"readonly":            types.ListType{ElemType: enabledActionElementType()},
			"rollover":            types.ListType{ElemType: rolloverActionElementType()},
			"searchable_snapshot": types.ListType{ElemType: searchableSnapshotActionElementType()},
			"set_priority":        types.ListType{ElemType: setPriorityActionElementType()},
			"shrink":              types.ListType{ElemType: shrinkActionElementType()},
			"unfollow":            types.ListType{ElemType: enabledActionElementType()},
			"wait_for_snapshot":   types.ListType{ElemType: waitForSnapshotActionElementType()},
			"downsample":          types.ListType{ElemType: downsampleActionElementType()},
		},
	}
}

func allocateActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"number_of_replicas":    types.Int64Type,
			"total_shards_per_node": types.Int64Type,
			"include":               jsontypes.NormalizedType{},
			"exclude":               jsontypes.NormalizedType{},
			"require":               jsontypes.NormalizedType{},
		},
	}
}

func deleteActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"delete_searchable_snapshot": types.BoolType,
		},
	}
}

func forcemergeActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"max_num_segments": types.Int64Type,
			"index_codec":      types.StringType,
		},
	}
}

func enabledActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
		},
	}
}

func rolloverActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"max_age":                types.StringType,
			"max_docs":               types.Int64Type,
			"max_size":               types.StringType,
			"max_primary_shard_docs": types.Int64Type,
			"max_primary_shard_size": types.StringType,
			"min_age":                types.StringType,
			"min_docs":               types.Int64Type,
			"min_size":               types.StringType,
			"min_primary_shard_docs": types.Int64Type,
			"min_primary_shard_size": types.StringType,
		},
	}
}

func searchableSnapshotActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"snapshot_repository": types.StringType,
			"force_merge_index":   types.BoolType,
		},
	}
}

func setPriorityActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"priority": types.Int64Type,
		},
	}
}

func shrinkActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"number_of_shards":         types.Int64Type,
			"max_primary_shard_size":   types.StringType,
			"allow_write_after_shrink": types.BoolType,
		},
	}
}

func waitForSnapshotActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"policy": types.StringType,
		},
	}
}

func downsampleActionElementType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"fixed_interval": types.StringType,
			"wait_timeout":   types.StringType,
		},
	}
}

func int64FromAny(v any) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}

func boolFromAny(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func stringFromAny(v any) string {
	s, _ := v.(string)
	return s
}

func int64ValueFromMap(m map[string]any, key string) types.Int64 {
	if v, ok := m[key]; ok {
		return types.Int64Value(int64FromAny(v))
	}
	return types.Int64Null()
}

func boolValueFromMap(m map[string]any, key string) types.Bool {
	if v, ok := m[key]; ok {
		return types.BoolValue(boolFromAny(v))
	}
	return types.BoolNull()
}

func stringValueFromMap(m map[string]any, key string) types.String {
	if v, ok := m[key]; ok {
		return types.StringValue(stringFromAny(v))
	}
	return types.StringNull()
}
