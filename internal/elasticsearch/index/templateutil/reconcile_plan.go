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

package templateutil

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Template attribute key constants shared across packages that embed a template block.
const (
	AttrSettings = "settings"
	AttrAlias    = "alias"
)

// ReconcileTemplateWithPriorStateForSemanticDrift aligns planned template.settings and
// template.alias attributes with prior state when Terraform would show a spurious diff:
// strict inequality but semantic equality (e.g. index settings canonical form in state
// vs practitioner JSON in configuration).
//
// Returns the updated Template object and true when any attribute was reconciled.
// Returns the original planTemplate and false when no reconciliation is needed.
// configTemplate may be null/unknown; it is used only for alias projection.
func ReconcileTemplateWithPriorStateForSemanticDrift(
	ctx context.Context,
	planTemplate, stateTemplate, configTemplate types.Object,
	attrTypes map[string]attr.Type,
) (types.Object, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if planTemplate.IsNull() || planTemplate.IsUnknown() || stateTemplate.IsNull() || stateTemplate.IsUnknown() {
		return planTemplate, false, diags
	}

	planAttrs := planTemplate.Attributes()
	stateAttrs := stateTemplate.Attributes()
	changed := false

	if ps, ok := planAttrs[AttrSettings]; ok && !ps.IsNull() && !ps.IsUnknown() {
		if ss, ok := stateAttrs[AttrSettings]; ok && !ss.IsNull() && !ss.IsUnknown() {
			pSet, okP := ps.(customtypes.IndexSettingsValue)
			sSet, okS := ss.(customtypes.IndexSettingsValue)
			if okP && okS {
				reconciled, settingsChanged, d := ReconcileSettingsIfSemanticallyEqual(ctx, pSet, sSet)
				diags.Append(d...)
				if diags.HasError() {
					return planTemplate, false, diags
				}
				if settingsChanged {
					planAttrs[AttrSettings] = reconciled
					changed = true
				}
			}
		}
	}

	if pa, ok := planAttrs[AttrAlias]; ok && !pa.IsNull() && !pa.IsUnknown() {
		if sa, ok := stateAttrs[AttrAlias]; ok && !sa.IsNull() && !sa.IsUnknown() {
			newAlias, aliasChanged, d := aliasutil.MergePlanAliasSetWithPriorState(ctx, pa, sa)
			diags.Append(d...)
			if diags.HasError() {
				return planTemplate, false, diags
			}
			if !aliasChanged && !configTemplate.IsNull() && !configTemplate.IsUnknown() {
				cfgAttrs := configTemplate.Attributes()
				if ca, ok := cfgAttrs[AttrAlias]; ok && !ca.IsNull() && !ca.IsUnknown() {
					// Use config encodings to match state (handles plan unknowns), but project
					// the result back onto the plan's element set so plan-only aliases are
					// preserved. mergePlanAliasSetWithPriorState alone would build the result
					// from its first argument and drop any aliases present in plan but not config.
					newAlias, aliasChanged, d = aliasutil.ProjectConfigAliasMatchesOntoPlan(ctx, pa, ca, sa)
					diags.Append(d...)
					if diags.HasError() {
						return planTemplate, false, diags
					}
				}
			}
			if aliasChanged {
				canonAlias, d := aliasutil.CanonicalizeAliasSetElements(ctx, newAlias)
				diags.Append(d...)
				if diags.HasError() {
					return planTemplate, false, diags
				}
				planAttrs[AttrAlias] = canonAlias
				changed = true
			}
		}
	}

	if !changed {
		return planTemplate, false, diags
	}

	newTpl, d := types.ObjectValue(attrTypes, planAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return planTemplate, false, diags
	}
	return newTpl, true, diags
}
