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

package securitydetectionrule

import (
	"context"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const kqlQueryLanguageKuery = "kuery"

// getKQLQueryLanguage maps language string to kbapi.SecurityDetectionsAPIKqlQueryLanguage
func (d Data) getKQLQueryLanguage() *kbapi.SecurityDetectionsAPIKqlQueryLanguage {
	if !typeutils.IsKnown(d.Language) {
		return nil
	}
	var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
	switch d.Language.ValueString() {
	case kqlQueryLanguageKuery:
		language = kqlQueryLanguageKuery
	case "lucene":
		language = "lucene"
	default:
		language = kqlQueryLanguageKuery
	}
	return &language
}

// commonAPIRuleFields holds the common fields extracted from any API rule response.
// Each updateFrom*Rule function populates this struct and calls updateCommonRuleFieldsFromAPI.
// Fields not applicable to a rule type (e.g. DataViewId for ESQL/ML) should be left nil.
type commonAPIRuleFields struct {
	ResourceID  string // rule.Id.String() — used to build the composite ID
	RuleID      string
	Name        string
	Type        string
	Enabled     bool
	From        string
	To          string
	Interval    string
	Description string
	RiskScore   int64
	Severity    string
	MaxSignals  int64
	Version     int64
	Revision    int64
	CreatedAt   time.Time
	CreatedBy   string
	UpdatedAt   time.Time
	UpdatedBy   string

	TimelineID                        *kbapi.SecurityDetectionsAPITimelineTemplateId
	TimelineTitle                     *kbapi.SecurityDetectionsAPITimelineTemplateTitle
	DataViewID                        *kbapi.SecurityDetectionsAPIDataViewId // nil for ESQL/ML → sets DataViewID to null
	Namespace                         *kbapi.SecurityDetectionsAPIAlertsIndexNamespace
	RuleNameOverride                  *kbapi.SecurityDetectionsAPIRuleNameOverride
	TimestampOverride                 *kbapi.SecurityDetectionsAPITimestampOverride
	TimestampOverrideFallbackDisabled *kbapi.SecurityDetectionsAPITimestampOverrideFallbackDisabled
	BuildingBlockType                 *kbapi.SecurityDetectionsAPIBuildingBlockType
	License                           *kbapi.SecurityDetectionsAPIRuleLicense
	Note                              *kbapi.SecurityDetectionsAPIInvestigationGuide

	Index          *[]string // nil for ESQL/ML → sets Index to empty list
	Author         []string
	Tags           []string
	FalsePositives []string
	References     []string
	Setup          kbapi.SecurityDetectionsAPISetupGuide

	Actions             []kbapi.SecurityDetectionsAPIRuleAction
	ExceptionsList      []kbapi.SecurityDetectionsAPIRuleExceptionList
	RiskScoreMapping    kbapi.SecurityDetectionsAPIRiskScoreMapping
	InvestigationFields *kbapi.SecurityDetectionsAPIInvestigationFields
	Threat              kbapi.SecurityDetectionsAPIThreatArray
	SeverityMapping     kbapi.SecurityDetectionsAPISeverityMapping
	RelatedIntegrations kbapi.SecurityDetectionsAPIRelatedIntegrationArray
	RequiredFields      kbapi.SecurityDetectionsAPIRequiredFieldArray
	// AlertSuppression is nil for Threshold rules (which use a different API type handled separately).
	AlertSuppression *kbapi.SecurityDetectionsAPIAlertSuppression
	ResponseActions  *[]kbapi.SecurityDetectionsAPIResponseAction
}

// updateCommonRuleFieldsFromAPI populates the Data fields that are shared across all rule types.
func (d *Data) updateCommonRuleFieldsFromAPI(ctx context.Context, fields commonAPIRuleFields) diag.Diagnostics {
	var diags diag.Diagnostics

	compID := clients.CompositeID{
		ClusterID:  d.SpaceID.ValueString(),
		ResourceID: fields.ResourceID,
	}
	d.ID = types.StringValue(compID.String())
	d.RuleID = types.StringValue(fields.RuleID)
	d.Name = types.StringValue(fields.Name)
	d.Type = types.StringValue(fields.Type)
	d.Enabled = types.BoolValue(fields.Enabled)
	d.From = types.StringValue(fields.From)
	d.To = types.StringValue(fields.To)
	d.Interval = types.StringValue(fields.Interval)
	d.Description = types.StringValue(fields.Description)
	d.RiskScore = types.Int64Value(fields.RiskScore)
	d.Severity = types.StringValue(fields.Severity)
	d.MaxSignals = types.Int64Value(fields.MaxSignals)
	d.Version = types.Int64Value(fields.Version)
	d.CreatedAt = typeutils.TimeToStringValue(fields.CreatedAt)
	d.CreatedBy = types.StringValue(fields.CreatedBy)
	d.UpdatedAt = typeutils.TimeToStringValue(fields.UpdatedAt)
	d.UpdatedBy = types.StringValue(fields.UpdatedBy)
	d.Revision = types.Int64Value(fields.Revision)

	d.TimelineID = typeutils.StringishPointerValue(fields.TimelineID)
	d.TimelineTitle = typeutils.StringishPointerValue(fields.TimelineTitle)
	d.DataViewID = typeutils.StringishPointerValue(fields.DataViewID)
	d.Namespace = typeutils.StringishPointerValue(fields.Namespace)
	d.RuleNameOverride = typeutils.StringishPointerValue(fields.RuleNameOverride)
	d.TimestampOverride = typeutils.StringishPointerValue(fields.TimestampOverride)
	d.TimestampOverrideFallbackDisabled = types.BoolPointerValue(fields.TimestampOverrideFallbackDisabled)
	d.BuildingBlockType = typeutils.StringishPointerValue(fields.BuildingBlockType)
	d.License = typeutils.StringishPointerValue(fields.License)
	d.Note = typeutils.StringishPointerValue(fields.Note)
	d.Setup = typeutils.NonEmptyStringishValue(fields.Setup)

	var indexStrs []string
	if fields.Index != nil {
		indexStrs = *fields.Index
	}
	d.Index = typeutils.StringsToListMust(indexStrs)
	d.Author = typeutils.StringsToListMust(fields.Author)
	d.Tags = typeutils.StringsToListMust(fields.Tags)
	d.FalsePositives = typeutils.StringsToListMust(fields.FalsePositives)
	d.References = typeutils.StringsToListMust(fields.References)

	diags.Append(d.updateActionsFromAPI(ctx, fields.Actions)...)
	diags.Append(d.updateExceptionsListFromAPI(ctx, fields.ExceptionsList)...)
	diags.Append(d.updateRiskScoreMappingFromAPI(ctx, fields.RiskScoreMapping)...)
	diags.Append(d.updateInvestigationFieldsFromAPI(ctx, fields.InvestigationFields)...)
	diags.Append(d.updateThreatFromAPI(ctx, &fields.Threat)...)
	diags.Append(d.updateSeverityMappingFromAPI(ctx, &fields.SeverityMapping)...)
	diags.Append(d.updateRelatedIntegrationsFromAPI(ctx, &fields.RelatedIntegrations)...)
	diags.Append(d.updateRequiredFieldsFromAPI(ctx, &fields.RequiredFields)...)
	diags.Append(d.updateAlertSuppressionFromAPI(ctx, fields.AlertSuppression)...)
	diags.Append(d.updateResponseActionsFromAPI(ctx, fields.ResponseActions)...)

	return diags
}

// updateListFieldFromAPI converts a slice to a types.List field using the provided
// converter. Returns nullList when slice is empty.
func updateListFieldFromAPI[T any](
	ctx context.Context,
	slice []T,
	nullList types.List,
	converter func(context.Context, []T) (types.List, diag.Diagnostics),
) (types.List, diag.Diagnostics) {
	if len(slice) > 0 {
		return converter(ctx, slice)
	}
	return nullList, nil
}
