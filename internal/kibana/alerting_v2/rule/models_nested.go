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

package rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	tagsPath           = path.Root("metadata").AtName("tags")
	routingTagsPath    = path.Root("metadata").AtName("routing_tags")
	groupingFieldsPath = path.Root("grouping").AtName("fields")
)

func (m *metadataModel) toAPI(ctx context.Context, diags *diag.Diagnostics) kbapi.KibanaHTTPAPIsAlertingRuleMetadata {
	return kbapi.KibanaHTTPAPIsAlertingRuleMetadata{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueStringPointer(),
		Tags:        typeutils.CollectionToSliceStringPtr(ctx, m.Tags, tagsPath, diags),
		RoutingTags: typeutils.CollectionToSliceStringPtr(ctx, m.RoutingTags, routingTagsPath, diags),
	}
}

func (q *queryModel) toAPI() kbapi.KibanaHTTPAPIsAlertingRuleQuery {
	query := kbapi.KibanaHTTPAPIsAlertingRuleQuery{Base: q.Base.ValueString()}
	if q.Breach != nil {
		query.Breach = &kbapi.KibanaHTTPAPIsAlertingRuleBreach{Segment: q.Breach.Segment.ValueString()}
	}
	return query
}

func (s *stateTransitionModel) toAPI() *kbapi.KibanaHTTPAPIsAlertingRuleStateTransition {
	if s == nil {
		return nil
	}
	transition := &kbapi.KibanaHTTPAPIsAlertingRuleStateTransition{}
	if s.Pending != nil {
		transition.Pending = &kbapi.KibanaHTTPAPIsAlertingRuleStateTransitionPending{
			Count:     typeutils.OptionalInt(s.Pending.Count),
			Timeframe: s.Pending.Timeframe.ValueStringPointer(),
			Operator:  enumPointer[kbapi.KibanaHTTPAPIsAlertingRuleStateTransitionPendingOperator](s.Pending.Operator),
		}
	}
	if s.Recovering != nil {
		transition.Recovering = &kbapi.KibanaHTTPAPIsAlertingRuleStateTransitionRecovering{
			Count:     typeutils.OptionalInt(s.Recovering.Count),
			Timeframe: s.Recovering.Timeframe.ValueStringPointer(),
			Operator:  enumPointer[kbapi.KibanaHTTPAPIsAlertingRuleStateTransitionRecoveringOperator](s.Recovering.Operator),
		}
	}
	return transition
}

func (g *groupingModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *kbapi.KibanaHTTPAPIsAlertingRuleGrouping {
	if g == nil {
		return nil
	}
	return &kbapi.KibanaHTTPAPIsAlertingRuleGrouping{
		Fields: typeutils.ListTypeToSliceString(ctx, g.Fields, groupingFieldsPath, diags),
	}
}

func enumPointer[T ~string](value types.String) *T {
	if !typeutils.IsKnown(value) {
		return nil
	}
	converted := T(value.ValueString())
	return &converted
}

func metadataFromAPI(ctx context.Context, metadata kbapi.KibanaHTTPAPIsAlertingRuleMetadata, diags *diag.Diagnostics) *metadataModel {
	if metadata.Builder != nil {
		diags.AddWarning(
			"Alerting v2 rule was authored by a Kibana UI builder",
			"The rule carries metadata.builder.type \""+metadata.Builder.Type+"\", which this resource does not manage. "+
				"Terraform-managed rules are ES|QL rules, so the next write that changes the rule converts it to an ES|QL rule and clears the builder.",
		)
	}
	return &metadataModel{
		Name:        types.StringValue(metadata.Name),
		Description: types.StringPointerValue(metadata.Description),
		Tags:        optionalStringListFromAPI(ctx, metadata.Tags, tagsPath, diags),
		RoutingTags: optionalStringListFromAPI(ctx, metadata.RoutingTags, routingTagsPath, diags),
	}
}

func queryFromAPI(query kbapi.KibanaHTTPAPIsAlertingRuleQuery) *queryModel {
	model := &queryModel{Base: types.StringValue(query.Base)}
	if query.Breach != nil {
		model.Breach = &breachModel{Segment: types.StringValue(query.Breach.Segment)}
	}
	return model
}

func stateTransitionFromAPI(transition *kbapi.KibanaHTTPAPIsAlertingRuleStateTransition) *stateTransitionModel {
	if transition == nil {
		return nil
	}
	model := &stateTransitionModel{}
	if transition.Pending != nil {
		model.Pending = stateTransitionPhaseFromAPI(transition.Pending.Count, transition.Pending.Timeframe, transition.Pending.Operator)
	}
	if transition.Recovering != nil {
		model.Recovering = stateTransitionPhaseFromAPI(transition.Recovering.Count, transition.Recovering.Timeframe, transition.Recovering.Operator)
	}
	return model
}

func stateTransitionPhaseFromAPI[T ~string](count *int, timeframe *string, operator *T) *stateTransitionPhaseModel {
	return &stateTransitionPhaseModel{
		Count:     typeutils.IntPointerToInt64Value(count),
		Timeframe: types.StringPointerValue(timeframe),
		Operator:  typeutils.StringishPointerValue(operator),
	}
}

func groupingFromAPI(ctx context.Context, grouping *kbapi.KibanaHTTPAPIsAlertingRuleGrouping, diags *diag.Diagnostics) *groupingModel {
	if grouping == nil {
		return nil
	}
	return &groupingModel{Fields: typeutils.SliceToListTypeString(ctx, grouping.Fields, groupingFieldsPath, diags)}
}

func optionalStringListFromAPI(ctx context.Context, values *[]string, p path.Path, diags *diag.Diagnostics) types.List {
	if values == nil {
		return types.ListNull(types.StringType)
	}
	return typeutils.SliceToListTypeString(ctx, *values, p, diags)
}
