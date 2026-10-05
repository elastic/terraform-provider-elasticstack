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

package alias

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *tfModel) modifyPlanReadIndexMembership(
	ctx context.Context,
	state tfModel,
	resolveIndexExpression resolveIndexExpressionFunc,
) diag.Diagnostics {
	if model.ReadIndices.IsNull() || model.ReadIndices.IsUnknown() {
		return nil
	}

	var readIndices []readIndexModel
	diags := model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	if diags.HasError() {
		return diags
	}

	stateReadIndices, stateDiags := readIndexModels(ctx, state.ReadIndices)
	if stateDiags.HasError() {
		return stateDiags
	}

	forceUnknown := resolveIndexExpression == nil || !hasSameReadIndexConfigurations(readIndices, stateReadIndices, state.ReadIndices)
	updated := false
	for i := range readIndices {
		if forceUnknown || readIndices[i].Name.IsUnknown() {
			readIndices[i].ConcreteIndices = types.SetUnknown(types.StringType)
			updated = true
			continue
		}

		stateReadIndex, found := stateReadIndexForPlan(readIndices[i], stateReadIndices)
		if found {
			readIndices[i].ConcreteIndices = stateReadIndex.ConcreteIndices
			updated = true
		}

		targets, resolveDiags := resolveIndexExpression(ctx, readIndices[i].Name.ValueString())
		diags.Append(resolveDiags...)
		if diags.HasError() {
			return diags
		}

		if !found || !hasConcreteMembership(ctx, stateReadIndex.ConcreteIndices, targets.Names) {
			readIndices[i].ConcreteIndices = types.SetUnknown(types.StringType)
			updated = true
		}
	}
	if !updated {
		return nil
	}

	readIndicesSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: getReadIndexAttrTypes(ctx),
	}, readIndices)
	if diags.HasError() {
		return diags
	}
	model.ReadIndices = readIndicesSet

	return nil
}

func readIndexModels(ctx context.Context, value types.Set) ([]readIndexModel, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var readIndices []readIndexModel
	diags := value.ElementsAs(ctx, &readIndices, false)
	if diags.HasError() {
		return nil, diags
	}

	return readIndices, nil
}

func hasSameReadIndexConfigurations(planReadIndices, stateReadIndices []readIndexModel, stateReadIndicesValue types.Set) bool {
	if stateReadIndicesValue.IsNull() || stateReadIndicesValue.IsUnknown() {
		return len(planReadIndices) == 0
	}

	if len(planReadIndices) != len(stateReadIndices) {
		return false
	}
	for _, planReadIndex := range planReadIndices {
		if _, found := stateReadIndexForPlan(planReadIndex, stateReadIndices); !found {
			return false
		}
	}

	return true
}

func stateReadIndexForPlan(planReadIndex readIndexModel, stateReadIndices []readIndexModel) (readIndexModel, bool) {
	for _, stateReadIndex := range stateReadIndices {
		if sameReadIndexConfiguration(planReadIndex, stateReadIndex) {
			return stateReadIndex, true
		}
	}

	return readIndexModel{}, false
}

func sameReadIndexConfiguration(a, b readIndexModel) bool {
	return a.Name.Equal(b.Name) &&
		a.Filter.Equal(b.Filter) &&
		a.IndexRouting.Equal(b.IndexRouting) &&
		a.IsHidden.Equal(b.IsHidden) &&
		a.Routing.Equal(b.Routing) &&
		a.SearchRouting.Equal(b.SearchRouting)
}

func hasConcreteMembership(ctx context.Context, concreteIndices types.Set, targets []string) bool {
	if concreteIndices.IsNull() || concreteIndices.IsUnknown() {
		return false
	}

	var current []string
	diags := concreteIndices.ElementsAs(ctx, &current, false)
	if diags.HasError() || len(current) != len(targets) {
		return false
	}

	currentSet := make(map[string]struct{}, len(current))
	for _, index := range current {
		currentSet[index] = struct{}{}
	}
	for _, target := range targets {
		if _, found := currentSet[target]; !found {
			return false
		}
	}

	return true
}
