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

package kibanaoapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetSlo retrieves a single SLO by space and ID. Returns (nil, nil) when
// the SLO is not found (HTTP 404), consistent with the resource layer's
// "not found" contract.
func GetSlo(ctx context.Context, client *Client, spaceID string, sloID string) (*kbapi.SLOsSloWithSummaryResponse, diag.Diagnostics) {
	resp, err := client.API.GetSloOpWithResponse(
		ctx,
		spaceID,
		sloID,
		&kbapi.GetSloOpParams{},
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to get SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return diagutil.UnwrapJSON200(resp.JSON200, "SLO")
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// EnableSlo calls the Kibana API to enable an existing SLO.
func EnableSlo(ctx context.Context, client *Client, spaceID, sloID string) diag.Diagnostics {
	resp, err := client.API.EnableSloOpWithResponse(ctx, spaceID, sloID)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to enable SLO", err.Error())}
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNoContent)
}

// DisableSlo calls the Kibana API to disable an existing SLO.
func DisableSlo(ctx context.Context, client *Client, spaceID, sloID string) diag.Diagnostics {
	resp, err := client.API.DisableSloOpWithResponse(ctx, spaceID, sloID)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to disable SLO", err.Error())}
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNoContent)
}

// CreateSlo creates a new SLO in the given space and returns the created SLO's ID.
func CreateSlo(ctx context.Context, client *Client, spaceID string, req kbapi.SLOsCreateSloRequest) (*kbapi.SLOsCreateSloResponse, diag.Diagnostics) {
	resp, err := client.API.CreateSloOpWithResponse(
		ctx,
		spaceID,
		req,
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to create SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return diagutil.UnwrapJSON200(resp.JSON200, "SLO")
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// UpdateSlo updates an existing SLO by space and ID.
func UpdateSlo(ctx context.Context, client *Client, spaceID string, sloID string, req kbapi.SLOsUpdateSloRequest) diag.Diagnostics {
	resp, err := client.API.UpdateSloOpWithResponse(
		ctx,
		spaceID,
		sloID,
		req,
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to update SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"SLO not found during update",
			"The SLO with ID "+sloID+" was not found in space "+spaceID+".",
		)}
	default:
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// DeleteSlo deletes an SLO by space and ID. A 404 response is treated as
// success (idempotent delete).
func DeleteSlo(ctx context.Context, client *Client, spaceID string, sloID string) diag.Diagnostics {
	resp, err := client.API.DeleteSloOpWithResponse(
		ctx,
		spaceID,
		sloID,
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to delete SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// FindSlos performs a paginated search for SLOs in the given space. The
// optional params allow filtering by KQL query, pagination, and sorting.
func FindSlos(ctx context.Context, client *Client, spaceID string, params *kbapi.FindSlosOpParams) (*kbapi.SLOsFindSloResponse, diag.Diagnostics) {
	resp, err := client.API.FindSlosOpWithResponse(
		ctx,
		spaceID,
		params,
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to find SLOs", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return diagutil.UnwrapJSON200(resp.JSON200, "SLOs")
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// SloResponseToModel converts a kbapi SLO response into the internal models.Slo type.
func SloResponseToModel(spaceID string, res *kbapi.SLOsSloWithSummaryResponse) *models.Slo {
	if res == nil {
		return nil
	}

	return &models.Slo{
		SloID:           res.Id,
		SpaceID:         spaceID,
		Name:            res.Name,
		Description:     res.Description,
		BudgetingMethod: res.BudgetingMethod,
		Indicator:       res.Indicator,
		TimeWindow:      res.TimeWindow,
		Objective:       res.Objective,
		Settings:        &res.Settings,
		GroupBy:         TransformGroupByFromResponse(res.GroupBy),
		Tags:            res.Tags,
		Enabled:         res.Enabled,
		Artifacts:       res.Artifacts,
	}
}

// sloIndicatorTarget is implemented by both SLOsCreateSloRequest_Indicator and
// SLOsUpdateSloRequest_Indicator, allowing a single switch to serve both.
type sloIndicatorTarget interface {
	FromSLOsIndicatorPropertiesApmAvailability(v kbapi.SLOsIndicatorPropertiesApmAvailability) error
	FromSLOsIndicatorPropertiesApmLatency(v kbapi.SLOsIndicatorPropertiesApmLatency) error
	FromSLOsIndicatorPropertiesCustomKql(v kbapi.SLOsIndicatorPropertiesCustomKql) error
	FromSLOsIndicatorPropertiesCustomMetric(v kbapi.SLOsIndicatorPropertiesCustomMetric) error
	FromSLOsIndicatorPropertiesHistogram(v kbapi.SLOsIndicatorPropertiesHistogram) error
	FromSLOsIndicatorPropertiesTimesliceMetric(v kbapi.SLOsIndicatorPropertiesTimesliceMetric) error
}

// applyResponseIndicator resolves the discriminator from s and calls the
// matching From* method on target. Adding a new indicator type requires
// updating only this function.
func applyResponseIndicator(s kbapi.SLOsSloWithSummaryResponse_Indicator, target sloIndicatorTarget) error {
	v, err := s.ValueByDiscriminator()
	if err != nil {
		return fmt.Errorf("unknown indicator type: %w", err)
	}
	switch ind := v.(type) {
	case kbapi.SLOsIndicatorPropertiesApmAvailability:
		return target.FromSLOsIndicatorPropertiesApmAvailability(ind)
	case kbapi.SLOsIndicatorPropertiesApmLatency:
		return target.FromSLOsIndicatorPropertiesApmLatency(ind)
	case kbapi.SLOsIndicatorPropertiesCustomKql:
		return target.FromSLOsIndicatorPropertiesCustomKql(ind)
	case kbapi.SLOsIndicatorPropertiesCustomMetric:
		return target.FromSLOsIndicatorPropertiesCustomMetric(ind)
	case kbapi.SLOsIndicatorPropertiesHistogram:
		return target.FromSLOsIndicatorPropertiesHistogram(ind)
	case kbapi.SLOsIndicatorPropertiesTimesliceMetric:
		return target.FromSLOsIndicatorPropertiesTimesliceMetric(ind)
	default:
		return fmt.Errorf("unhandled indicator type: %T", v)
	}
}

// ResponseIndicatorToCreateIndicator converts the response indicator union type to the
// create request indicator union type.
func ResponseIndicatorToCreateIndicator(s kbapi.SLOsSloWithSummaryResponse_Indicator) (kbapi.SLOsCreateSloRequest_Indicator, error) {
	var ret kbapi.SLOsCreateSloRequest_Indicator
	return ret, applyResponseIndicator(s, &ret)
}

// ResponseIndicatorToUpdateIndicator converts the response indicator union type to the
// update request indicator union type.
func ResponseIndicatorToUpdateIndicator(s kbapi.SLOsSloWithSummaryResponse_Indicator) (kbapi.SLOsUpdateSloRequest_Indicator, error) {
	var ret kbapi.SLOsUpdateSloRequest_Indicator
	return ret, applyResponseIndicator(s, &ret)
}

// TransformGroupBy converts a slice of group-by field names to the kbapi union type.
func TransformGroupBy(groupBy []string, supportsGroupByList bool) *kbapi.SLOsGroupBy {
	if groupBy == nil {
		return nil
	}

	var gb kbapi.SLOsGroupBy
	if supportsGroupByList {
		if err := gb.FromSLOsGroupBy1(groupBy); err != nil {
			return nil
		}
		return &gb
	}

	if len(groupBy) == 0 {
		return nil
	}

	if err := gb.FromSLOsGroupBy0(groupBy[0]); err != nil {
		return nil
	}
	return &gb
}

// TransformGroupByFromResponse converts the kbapi GroupBy union back to a string slice.
func TransformGroupByFromResponse(groupBy kbapi.SLOsGroupBy) []string {
	// Try string first
	if s, err := groupBy.AsSLOsGroupBy0(); err == nil {
		return []string{s}
	}

	// Try array
	if arr, err := groupBy.AsSLOsGroupBy1(); err == nil {
		return arr
	}

	return nil
}
