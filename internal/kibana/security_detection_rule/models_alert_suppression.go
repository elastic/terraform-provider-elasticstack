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
	"fmt"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Helper function to convert alert suppression from TF data to API type
func (d Data) alertSuppressionToAPI(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIAlertSuppression {
	if !typeutils.IsKnown(d.AlertSuppression) {
		return nil
	}

	var model AlertSuppressionModel
	objDiags := d.AlertSuppression.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(objDiags...)
	if diags.HasError() {
		return nil
	}

	suppression := &kbapi.SecurityDetectionsAPIAlertSuppression{}

	// Handle group_by (required)
	if typeutils.IsKnown(model.GroupBy) {
		groupByList := typeutils.ListTypeToSliceString(ctx, model.GroupBy, path.Root("alert_suppression").AtName("group_by"), diags)
		if len(groupByList) > 0 {
			suppression.GroupBy = groupByList
		}
	}

	// Handle duration (optional)
	if typeutils.IsKnown(model.Duration) {
		duration, durationDiags := parseDurationToAPI(model.Duration)
		diags.Append(durationDiags...)
		if !durationDiags.HasError() {
			suppression.Duration = &duration
		}
	}

	// Handle missing_fields_strategy (optional)
	if typeutils.IsKnown(model.MissingFieldsStrategy) {
		strategy := kbapi.SecurityDetectionsAPIAlertSuppressionMissingFieldsStrategy(model.MissingFieldsStrategy.ValueString())
		suppression.MissingFieldsStrategy = &strategy
	}

	return suppression
}

// Helper function to convert alert suppression from TF data to threshold-specific API type
func (d Data) alertSuppressionToThresholdAPI(ctx context.Context, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIThresholdAlertSuppression {
	if !typeutils.IsKnown(d.AlertSuppression) {
		return nil
	}

	var model AlertSuppressionModel
	objDiags := d.AlertSuppression.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(objDiags...)
	if diags.HasError() {
		return nil
	}

	suppression := &kbapi.SecurityDetectionsAPIThresholdAlertSuppression{}

	// Handle duration (required for threshold alert suppression)
	if !typeutils.IsKnown(model.Duration) {
		diags.AddError(
			"Duration required for threshold alert suppression",
			"Threshold alert suppression requires a duration to be specified",
		)
		return nil
	}

	duration, durationDiags := parseDurationToAPI(model.Duration)
	diags.Append(durationDiags...)
	if !durationDiags.HasError() {
		suppression.Duration = duration
	}

	// Note: Threshold alert suppression only supports duration field.
	// GroupBy and MissingFieldsStrategy are not supported for threshold rules.

	return suppression
}

func (d *Data) updateAlertSuppressionFromAPI(ctx context.Context, apiSuppression *kbapi.SecurityDetectionsAPIAlertSuppression) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiSuppression == nil {
		d.AlertSuppression = types.ObjectNull(getAlertSuppressionType())
		return diags
	}

	model := AlertSuppressionModel{}

	// Convert group_by (required field according to API)
	if len(apiSuppression.GroupBy) > 0 {
		groupByList := make([]attr.Value, len(apiSuppression.GroupBy))
		for i, field := range apiSuppression.GroupBy {
			groupByList[i] = types.StringValue(field)
		}
		model.GroupBy = types.ListValueMust(types.StringType, groupByList)
	} else {
		model.GroupBy = types.ListNull(types.StringType)
	}

	// Convert duration (optional)
	if apiSuppression.Duration != nil {
		model.Duration = parseDurationFromAPI(*apiSuppression.Duration)
	} else {
		model.Duration = customtypes.NewDurationNull()
	}

	// Convert missing_fields_strategy (optional)
	if apiSuppression.MissingFieldsStrategy != nil {
		model.MissingFieldsStrategy = types.StringValue(string(*apiSuppression.MissingFieldsStrategy))
	} else {
		model.MissingFieldsStrategy = types.StringNull()
	}

	alertSuppressionObj, objDiags := types.ObjectValueFrom(ctx, getAlertSuppressionType(), model)
	diags.Append(objDiags...)

	d.AlertSuppression = alertSuppressionObj

	return diags
}

func (d *Data) updateThresholdAlertSuppressionFromAPI(ctx context.Context, apiSuppression *kbapi.SecurityDetectionsAPIThresholdAlertSuppression) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiSuppression == nil {
		d.AlertSuppression = types.ObjectNull(getAlertSuppressionType())
		return diags
	}

	model := AlertSuppressionModel{

		// Threshold alert suppression only has duration field, so we set group_by and missing_fields_strategy to null
		GroupBy:               types.ListNull(types.StringType),
		MissingFieldsStrategy: types.StringNull(),

		// Convert duration (always present in threshold alert suppression)
		Duration: parseDurationFromAPI(apiSuppression.Duration)}

	alertSuppressionObj, objDiags := types.ObjectValueFrom(ctx, getAlertSuppressionType(), model)
	diags.Append(objDiags...)

	d.AlertSuppression = alertSuppressionObj

	return diags
}

// parseDurationToAPI converts a customtypes.Duration to the API structure
func parseDurationToAPI(duration customtypes.Duration) (kbapi.SecurityDetectionsAPIAlertSuppressionDuration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(duration) {
		diags.AddError("Duration Parse error", "duration string value is unknown")
		return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{}, diags
	}

	// Get the raw duration string (e.g. "5m", "1h", "30s")
	durationStr := duration.ValueString()

	value, unitStr, err := validators.ParseUnitDuration(durationStr, "smhd")
	if err != nil {
		diags.AddError(
			"Invalid duration format",
			fmt.Sprintf("Duration '%s' is not in valid format. Expected format: number followed by unit (s, m, h, d)", durationStr),
		)
		return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{}, diags
	}

	// Map the unit from the string to the API unit type
	var unit kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnit
	switch unitStr {
	case "s":
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitS
	case "m":
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitM
	case "h":
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitH
	case "d":
		// Convert days to hours since API doesn't support days unit
		value *= 24
		unit = kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitH
	default:
		diags.AddError(
			"Unsupported duration unit",
			fmt.Sprintf("Unit '%s' is not supported. Supported units: s, m, h, d", unitStr),
		)
		return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{}, diags
	}

	return kbapi.SecurityDetectionsAPIAlertSuppressionDuration{
		Value: value,
		Unit:  unit,
	}, diags
}

// parseDurationFromAPI converts an API duration to customtypes.Duration
func parseDurationFromAPI(apiDuration kbapi.SecurityDetectionsAPIAlertSuppressionDuration) customtypes.Duration {
	// Convert the API's Value + Unit format back to a duration string
	durationStr := strconv.Itoa(apiDuration.Value) + string(apiDuration.Unit)
	return customtypes.NewDurationValue(durationStr)
}
