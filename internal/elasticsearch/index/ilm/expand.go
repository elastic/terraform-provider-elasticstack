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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var ilmActionSettingOptions = map[string]struct {
	skipEmptyCheck bool
	def            any
	minVersion     *version.Version
}{
	attrAllowWriteAfterShrink: {def: false, minVersion: version.Must(version.NewVersion("8.14.0"))},
	attrNumberOfReplicas:      {skipEmptyCheck: true},
	attrPriority:              {skipEmptyCheck: true},
	attrMaxPrimaryShardDocs:   {def: 0, minVersion: MaxPrimaryShardDocsMinSupportedVersion},
	attrMinAge:                {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	attrMinDocs:               {def: 0, minVersion: RolloverMinConditionsMinSupportedVersion},
	attrMinSize:               {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	attrMinPrimaryShardDocs:   {def: 0, minVersion: RolloverMinConditionsMinSupportedVersion},
	attrMinPrimaryShardSize:   {def: "", minVersion: RolloverMinConditionsMinSupportedVersion},
	attrTotalShardsPerNode:    {skipEmptyCheck: true},
}

func expandPhase(p map[string]any, serverVersion *version.Version) (*models.Phase, diag.Diagnostics) {
	var diags diag.Diagnostics
	var phase models.Phase

	if v, ok := p[attrMinAge].(string); ok && v != "" {
		phase.MinAge = v
	}
	delete(p, attrMinAge)

	actions := make(map[string]models.Action)
	for actionName, action := range p {
		a, ok := action.([]any)
		if !ok || len(a) == 0 {
			continue
		}

		switch actionName {
		case ilmActionAllocate:
			actions[actionName], diags = expandAction(a, serverVersion, attrNumberOfReplicas, attrTotalShardsPerNode, attrInclude, attrExclude, attrRequire)
		case ilmPhaseDelete:
			actions[actionName], diags = expandAction(a, serverVersion, attrDeleteSearchableSnapshot)
		case ilmActionForcemerge:
			actions[actionName], diags = expandAction(a, serverVersion, "max_num_segments", "index_codec")
		case ilmActionFreeze:
			if a[0] != nil {
				ac := a[0].(map[string]any)
				if ac[attrEnabled].(bool) {
					actions[actionName], diags = expandAction(a, serverVersion)
				}
			}
		case ilmActionMigrate:
			actions[actionName], diags = expandAction(a, serverVersion, attrEnabled)
		case ilmActionReadonly:
			if a[0] != nil {
				ac := a[0].(map[string]any)
				if ac[attrEnabled].(bool) {
					actions[actionName], diags = expandAction(a, serverVersion)
				}
			}
		case ilmActionRollover:
			actions[actionName], diags = expandAction(
				a,
				serverVersion,
				attrMaxAge,
				"max_docs",
				"max_size",
				attrMaxPrimaryShardDocs,
				attrMaxPrimaryShardSize,
				attrMinAge,
				attrMinDocs,
				attrMinSize,
				attrMinPrimaryShardDocs,
				attrMinPrimaryShardSize,
			)
		case ilmActionSearchableSnapshot:
			actions[actionName], diags = expandAction(a, serverVersion, attrSnapshotRepository, attrForceMergeIndex)
		case ilmActionSetPriority:
			actions[actionName], diags = expandAction(a, serverVersion, attrPriority)
		case ilmActionShrink:
			actions[actionName], diags = expandAction(a, serverVersion, "number_of_shards", attrMaxPrimaryShardSize, attrAllowWriteAfterShrink)
		case ilmActionUnfollow:
			if a[0] != nil {
				ac := a[0].(map[string]any)
				if ac[attrEnabled].(bool) {
					actions[actionName], diags = expandAction(a, serverVersion)
				}
			}
		case ilmActionWaitForSnapshot:
			actions[actionName], diags = expandAction(a, serverVersion, "policy")
		case ilmActionDownsample:
			actions[actionName], diags = expandAction(a, serverVersion, attrFixedInterval, attrWaitTimeout)
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

				if options.skipEmptyCheck || !typeutils.IsEmpty(v) {
					if setting == attrInclude || setting == attrExclude || setting == attrRequire {
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
