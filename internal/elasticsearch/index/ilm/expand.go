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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var ilmActionSettingOptions = map[string]struct {
	skipEmptyCheck bool
	def            any
	minVersion     *version.Version
}{
	"allow_write_after_shrink": {def: false, minVersion: version.Must(version.NewVersion("8.14.0"))},
	"number_of_replicas":       {skipEmptyCheck: true},
	"priority":                 {skipEmptyCheck: true},
	"max_primary_shard_docs":   {def: 0, minVersion: MaxPrimaryShardDocsMinSupportedVersion},
	"min_age":                  {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_docs":                 {def: 0, minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_size":                 {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_primary_shard_docs":   {def: 0, minVersion: RolloverMinConditionsMinSupportedVersion},
	"min_primary_shard_size":   {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	"total_shards_per_node":    {skipEmptyCheck: true, def: -1, minVersion: version.Must(version.NewVersion("7.16.0"))},
}

func expandPhase(p map[string]any, serverVersion *version.Version) (*models.Phase, diag.Diagnostics) {
	var diags diag.Diagnostics
	var phase models.Phase

	if v, ok := p["min_age"].(string); ok && v != "" {
		phase.MinAge = v
	}
	delete(p, "min_age")

	actions := make(map[string]models.Action)
	for actionName, action := range p {
		a, ok := action.([]any)
		if !ok || len(a) == 0 {
			continue
		}

		switch actionName {
		case "allocate":
			actions[actionName], diags = expandAction(a, serverVersion, "number_of_replicas", "total_shards_per_node", "include", "exclude", "require")
		case ilmPhaseDelete:
			actions[actionName], diags = expandAction(a, serverVersion, "delete_searchable_snapshot")
		case "forcemerge":
			actions[actionName], diags = expandAction(a, serverVersion, "max_num_segments", "index_codec")
		case "freeze":
			if a[0] != nil {
				ac := a[0].(map[string]any)
				if ac["enabled"].(bool) {
					actions[actionName], diags = expandAction(a, serverVersion)
				}
			}
		case "migrate":
			actions[actionName], diags = expandAction(a, serverVersion, "enabled")
		case "readonly":
			if a[0] != nil {
				ac := a[0].(map[string]any)
				if ac["enabled"].(bool) {
					actions[actionName], diags = expandAction(a, serverVersion)
				}
			}
		case "rollover":
			actions[actionName], diags = expandAction(
				a,
				serverVersion,
				"max_age",
				"max_docs",
				"max_size",
				"max_primary_shard_docs",
				"max_primary_shard_size",
				"min_age",
				"min_docs",
				"min_size",
				"min_primary_shard_docs",
				"min_primary_shard_size",
			)
		case "searchable_snapshot":
			actions[actionName], diags = expandAction(a, serverVersion, "snapshot_repository", "force_merge_index")
		case "set_priority":
			actions[actionName], diags = expandAction(a, serverVersion, "priority")
		case "shrink":
			actions[actionName], diags = expandAction(a, serverVersion, "number_of_shards", "max_primary_shard_size", "allow_write_after_shrink")
		case "unfollow":
			if a[0] != nil {
				ac := a[0].(map[string]any)
				if ac["enabled"].(bool) {
					actions[actionName], diags = expandAction(a, serverVersion)
				}
			}
		case "wait_for_snapshot":
			actions[actionName], diags = expandAction(a, serverVersion, "policy")
		case "downsample":
			actions[actionName], diags = expandAction(a, serverVersion, "fixed_interval", "wait_timeout")
		default:
			diags.AddError("Unknown action defined.", fmt.Sprintf(`Configured action "%s" is not supported`, actionName))
			return nil, diags
		}
		if diags.HasError() {
			return nil, diags
		}
	}

	phase.Actions = actions
	return &phase, diags
}

func expandAction(a []any, serverVersion *version.Version, settings ...string) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	def := make(map[string]any)

	if action := a[0]; action != nil {
		for _, setting := range settings {
			if v, ok := action.(map[string]any)[setting]; ok && v != nil {
				options := ilmActionSettingOptions[setting]

				if options.minVersion != nil && serverVersion != nil && options.minVersion.GreaterThan(serverVersion) {
					if v != options.def {
						var unsupported diag.Diagnostics
						unsupported.AddError(
							"Unsupported ILM setting",
							fmt.Sprintf("[%s] is not supported in the target Elasticsearch server. Remove the setting from your module definition or set it to the default [%v] value", setting, options.def),
						)
						return nil, unsupported
					}
					continue
				}

				if options.skipEmptyCheck || !schemautil.IsEmpty(v) {
					if setting == "include" || setting == "exclude" || setting == "require" {
						res := make(map[string]any)
						if err := json.Unmarshal([]byte(v.(string)), &res); err != nil {
							diags.AddError("Invalid JSON", err.Error())
							return nil, diags
						}
						def[setting] = res
					} else {
						def[setting] = v
					}
				}
			}
		}
	}
	return def, diags
}

func expandIlmPolicy(name string, metadata string, phases map[string]map[string]any, serverVersion *version.Version) (*models.Policy, diag.Diagnostics) {
	var diags diag.Diagnostics
	var policy models.Policy

	policy.Name = name

	if strings.TrimSpace(metadata) != "" {
		meta := make(map[string]any)
		if err := json.NewDecoder(strings.NewReader(metadata)).Decode(&meta); err != nil {
			diags.AddError("Invalid metadata JSON", err.Error())
			return nil, diags
		}
		policy.Metadata = meta
	}

	outPhases := make(map[string]models.Phase)
	for ph, raw := range phases {
		if raw == nil {
			continue
		}
		phase, d := expandPhase(raw, serverVersion)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		outPhases[ph] = *phase
	}

	policy.Phases = outPhases
	return &policy, diags
}
