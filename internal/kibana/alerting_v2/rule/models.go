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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// minVersion is the first Elastic Stack version with the Alerting v2 rules API.
var minVersion = version.Must(version.NewVersion("9.6.0-SNAPSHOT"))

// defaultTimeField is the server default applied when time_field is omitted.
const defaultTimeField = "@timestamp"

type Model struct {
	entitycore.ResourceTimeoutsField
	entitycore.KibanaConnectionField
	ID     types.String `tfsdk:"id"`
	RuleID types.String `tfsdk:"rule_id"`
	entitycore.SpaceIDField
	Kind            types.String          `tfsdk:"kind"`
	Enabled         types.Bool            `tfsdk:"enabled"`
	TimeField       types.String          `tfsdk:"time_field"`
	Metadata        *metadataModel        `tfsdk:"metadata"`
	Schedule        *scheduleModel        `tfsdk:"schedule"`
	Query           *queryModel           `tfsdk:"query"`
	Recovery        *recoveryModel        `tfsdk:"recovery"`
	NoData          *noDataModel          `tfsdk:"no_data"`
	StateTransition *stateTransitionModel `tfsdk:"state_transition"`
	Grouping        *groupingModel        `tfsdk:"grouping"`
	Artifacts       []artifactModel       `tfsdk:"artifacts"`
	Version         types.Int64           `tfsdk:"version"`
	CreatedAt       types.String          `tfsdk:"created_at"`
	UpdatedAt       types.String          `tfsdk:"updated_at"`
}

type metadataModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	RoutingTags types.List   `tfsdk:"routing_tags"`
}

type scheduleModel struct {
	Every    types.String `tfsdk:"every"`
	Lookback types.String `tfsdk:"lookback"`
}

type queryModel struct {
	Base   types.String `tfsdk:"base"`
	Breach *breachModel `tfsdk:"breach"`
}

type breachModel struct {
	Segment types.String `tfsdk:"segment"`
}

type stateTransitionModel struct {
	Pending    *stateTransitionPhaseModel `tfsdk:"pending"`
	Recovering *stateTransitionPhaseModel `tfsdk:"recovering"`
}

type stateTransitionPhaseModel struct {
	Count     types.Int64  `tfsdk:"count"`
	Timeframe types.String `tfsdk:"timeframe"`
	Operator  types.String `tfsdk:"operator"`
}

type groupingModel struct {
	Fields types.List `tfsdk:"fields"`
}

type artifactModel struct {
	ID   types.String         `tfsdk:"id"`
	Type types.String         `tfsdk:"type"`
	Data jsontypes.Normalized `tfsdk:"data"`
}

type recoveryModel struct {
	Strategy types.String `tfsdk:"strategy"`
	Segment  types.String `tfsdk:"segment"`
	Query    types.String `tfsdk:"query"`
}

type noDataModel struct {
	Strategy types.String `tfsdk:"strategy"`
	Query    types.String `tfsdk:"query"`
}

// GetVersionRequirements enforces the 9.6.0 minimum on every lifecycle operation.
func (m Model) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return entitycore.SingleVersionRequirement(
		*minVersion,
		"Alerting v2 rules require Elastic Stack v"+minVersion.Core().String()+" or later.",
	), nil
}

func (m Model) toAPIPutBody(ctx context.Context) (kbapi.KibanaHTTPAPIsAlertingPutRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.KibanaHTTPAPIsAlertingPutRule{
		Kind:      kbapi.KibanaHTTPAPIsAlertingPutRuleKind(m.Kind.ValueString()),
		Enabled:   m.Enabled.ValueBoolPointer(),
		TimeField: m.TimeField.ValueStringPointer(),
		Metadata:  m.Metadata.toAPI(ctx, &diags),
		Schedule: kbapi.KibanaHTTPAPIsAlertingRuleSchedule{
			Every:    m.Schedule.Every.ValueString(),
			Lookback: m.Schedule.Lookback.ValueStringPointer(),
		},
		Query:           m.Query.toAPI(),
		StateTransition: m.StateTransition.toAPI(),
		Grouping:        m.Grouping.toAPI(ctx, &diags),
		Artifacts:       artifactsToAPI(m.Artifacts, &diags),
	}

	recovery, d := m.Recovery.toAPI()
	diags.Append(d...)
	body.Recovery = recovery

	noData, d := m.NoData.toAPI()
	diags.Append(d...)
	body.NoData = noData

	return body, diags
}

func (m *Model) populateFromAPI(ctx context.Context, resp *kbapi.KibanaHTTPAPIsAlertingRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Kind = types.StringValue(string(resp.Kind))
	m.Enabled = types.BoolValue(resp.Enabled)
	m.TimeField = types.StringValue(defaultTimeField)
	if resp.TimeField != nil {
		m.TimeField = types.StringValue(*resp.TimeField)
	}
	m.Metadata = metadataFromAPI(ctx, resp.Metadata, &diags)
	m.Schedule = &scheduleModel{
		Every:    types.StringValue(resp.Schedule.Every),
		Lookback: types.StringPointerValue(resp.Schedule.Lookback),
	}
	m.Query = queryFromAPI(resp.Query)

	recovery, d := recoveryFromAPI(resp.Recovery)
	diags.Append(d...)
	m.Recovery = recovery

	noData, d := noDataFromAPI(resp.NoData)
	diags.Append(d...)
	m.NoData = noData

	m.StateTransition = stateTransitionFromAPI(resp.StateTransition)
	m.Grouping = groupingFromAPI(ctx, resp.Grouping, &diags)
	m.Artifacts = artifactsFromAPI(resp.Artifacts, &diags)

	m.Version = types.Int64Value(int64(resp.Version))
	m.CreatedAt = types.StringValue(resp.CreatedAt.UTC().Format(kbschema.KibanaTimestampLayout))
	m.UpdatedAt = types.StringValue(resp.UpdatedAt.UTC().Format(kbschema.KibanaTimestampLayout))

	return diags
}

func (m Model) GetID() types.String         { return m.ID }
func (m Model) GetResourceID() types.String { return m.RuleID }
