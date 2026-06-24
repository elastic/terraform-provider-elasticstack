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
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const osqueryFindSavedQueriesPageSize = 100

// OsquerySavedQueryCreateEntity is the unwrapped `data` object from POST /api/osquery/saved_queries.
type OsquerySavedQueryCreateEntity struct {
	CreatedAt           *time.Time
	CreatedBy           *string
	CreatedByProfileUID *string
	Description         *kbapi.SecurityOsqueryAPISavedQueryDescription
	EcsMapping          *kbapi.SecurityOsqueryAPIECSMapping
	ID                  kbapi.SecurityOsqueryAPISavedQueryId
	Interval            *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Interval
	Platform            *kbapi.SecurityOsqueryAPIPlatform
	Prebuilt            *bool
	Query               *kbapi.SecurityOsqueryAPIQuery
	Removed             *kbapi.SecurityOsqueryAPIRemoved
	SavedObjectID       string
	Snapshot            *kbapi.SecurityOsqueryAPISnapshot
	Timeout             *int
	UpdatedAt           *time.Time
	UpdatedBy           *string
	UpdatedByProfileUID *string
	Version             *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Version
}

// OsquerySavedQueryGetEntity is the unwrapped `data` object from GET /api/osquery/saved_queries/{id}.
type OsquerySavedQueryGetEntity struct {
	CreatedAt           *time.Time
	CreatedBy           *string
	CreatedByProfileUID *string
	Description         *kbapi.SecurityOsqueryAPISavedQueryDescription
	EcsMapping          *kbapi.SecurityOsqueryAPIECSMapping
	ID                  kbapi.SecurityOsqueryAPISavedQueryId
	Interval            *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Interval
	Platform            *kbapi.SecurityOsqueryAPIPlatform
	Prebuilt            *bool
	Query               *kbapi.SecurityOsqueryAPIQuery
	Removed             *kbapi.SecurityOsqueryAPIRemoved
	SavedObjectID       string
	Snapshot            *kbapi.SecurityOsqueryAPISnapshot
	Timeout             *int
	UpdatedAt           *time.Time
	UpdatedBy           *string
	UpdatedByProfileUID *string
	Version             *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Version
}

// OsquerySavedQueryUpdateEntity is the unwrapped `data` object from PUT /api/osquery/saved_queries/{id}.
type OsquerySavedQueryUpdateEntity struct {
	CreatedAt           *time.Time
	CreatedBy           *string
	CreatedByProfileUID *string
	Description         *kbapi.SecurityOsqueryAPISavedQueryDescription
	EcsMapping          *kbapi.SecurityOsqueryAPIECSMapping
	ID                  kbapi.SecurityOsqueryAPISavedQueryId
	Interval            *kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse_Data_Interval
	Platform            *kbapi.SecurityOsqueryAPIPlatform
	Prebuilt            *bool
	Query               *kbapi.SecurityOsqueryAPIQuery
	Removed             *kbapi.SecurityOsqueryAPIRemoved
	SavedObjectID       string
	Snapshot            *kbapi.SecurityOsqueryAPISnapshot
	Timeout             *int
	UpdatedAt           *time.Time
	UpdatedBy           *string
	UpdatedByProfileUID *string
	Version             *string
}

func osquerySavedQueryCreateEntityFrom(resp *kbapi.SecurityOsqueryAPICreateSavedQueryResponse) *OsquerySavedQueryCreateEntity {
	if resp == nil {
		return nil
	}
	d := resp.Data
	return &OsquerySavedQueryCreateEntity{
		CreatedAt:           d.CreatedAt,
		CreatedBy:           d.CreatedBy,
		CreatedByProfileUID: d.CreatedByProfileUid,
		Description:         d.Description,
		EcsMapping:          d.EcsMapping,
		ID:                  d.Id,
		Interval:            d.Interval,
		Platform:            d.Platform,
		Prebuilt:            d.Prebuilt,
		Query:               d.Query,
		Removed:             d.Removed,
		SavedObjectID:       d.SavedObjectId,
		Snapshot:            d.Snapshot,
		Timeout:             d.Timeout,
		UpdatedAt:           d.UpdatedAt,
		UpdatedBy:           d.UpdatedBy,
		UpdatedByProfileUID: d.UpdatedByProfileUid,
		Version:             d.Version,
	}
}

func osquerySavedQueryGetEntityFrom(resp *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse) *OsquerySavedQueryGetEntity {
	if resp == nil {
		return nil
	}
	d := resp.Data
	return &OsquerySavedQueryGetEntity{
		CreatedAt:           d.CreatedAt,
		CreatedBy:           d.CreatedBy,
		CreatedByProfileUID: d.CreatedByProfileUid,
		Description:         d.Description,
		EcsMapping:          d.EcsMapping,
		ID:                  d.Id,
		Interval:            d.Interval,
		Platform:            d.Platform,
		Prebuilt:            d.Prebuilt,
		Query:               d.Query,
		Removed:             d.Removed,
		SavedObjectID:       d.SavedObjectId,
		Snapshot:            d.Snapshot,
		Timeout:             d.Timeout,
		UpdatedAt:           d.UpdatedAt,
		UpdatedBy:           d.UpdatedBy,
		UpdatedByProfileUID: d.UpdatedByProfileUid,
		Version:             d.Version,
	}
}

func osquerySavedQueryUpdateEntityFrom(resp *kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse) *OsquerySavedQueryUpdateEntity {
	if resp == nil {
		return nil
	}
	d := resp.Data
	return &OsquerySavedQueryUpdateEntity{
		CreatedAt:           d.CreatedAt,
		CreatedBy:           d.CreatedBy,
		CreatedByProfileUID: d.CreatedByProfileUid,
		Description:         d.Description,
		EcsMapping:          d.EcsMapping,
		ID:                  d.Id,
		Interval:            d.Interval,
		Platform:            d.Platform,
		Prebuilt:            d.Prebuilt,
		Query:               d.Query,
		Removed:             d.Removed,
		SavedObjectID:       d.SavedObjectId,
		Snapshot:            d.Snapshot,
		Timeout:             d.Timeout,
		UpdatedAt:           d.UpdatedAt,
		UpdatedBy:           d.UpdatedBy,
		UpdatedByProfileUID: d.UpdatedByProfileUid,
		Version:             d.Version,
	}
}

// CreateOsquerySavedQuery creates a new Osquery saved query via POST /api/osquery/saved_queries.
func CreateOsquerySavedQuery(
	ctx context.Context,
	client *Client,
	spaceID string,
	body kbapi.OsqueryCreateSavedQueryJSONRequestBody,
) (*OsquerySavedQueryCreateEntity, diag.Diagnostics) {
	resp, err := client.API.OsqueryCreateSavedQueryWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *OsquerySavedQueryCreateEntity {
			return osquerySavedQueryCreateEntityFrom(resp.JSON200)
		})
}

// FindOsquerySavedObjectID resolves the Kibana saved object ID for a user-facing saved_query_id
// by paginating GET /api/osquery/saved_queries (OsqueryFindSavedQueries). Returns ("", false, nil)
// when no matching query exists in the space.
func FindOsquerySavedObjectID(
	ctx context.Context,
	client *Client,
	spaceID string,
	savedQueryID kbapi.SecurityOsqueryAPISavedQueryId,
) (string, bool, diag.Diagnostics) {
	pageSize := kbapi.SecurityOsqueryAPIPageSizeOrUndefined(osqueryFindSavedQueriesPageSize)
	page := 1

	for {
		pageParam := page
		resp, err := client.API.OsqueryFindSavedQueriesWithResponse(ctx, &kbapi.OsqueryFindSavedQueriesParams{
			Page:     &pageParam,
			PageSize: &pageSize,
		}, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return "", false, diagutil.FrameworkDiagFromError(err)
		}

		if resp.StatusCode() != http.StatusOK {
			return "", false, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
		}
		if resp.JSON200 == nil {
			return "", false, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to parse response",
					"Osquery find saved queries returned success status but response body was nil or not JSON",
				),
			}
		}

		for _, item := range resp.JSON200.Data {
			if item.Id == savedQueryID {
				return item.SavedObjectId, true, nil
			}
		}

		if len(resp.JSON200.Data) == 0 {
			break
		}

		perPage := pageSize
		if resp.JSON200.PerPage > 0 {
			perPage = resp.JSON200.PerPage
		}
		if page*perPage >= resp.JSON200.Total {
			break
		}
		page++
	}

	return "", false, nil
}

// GetOsquerySavedQueryBySavedObjectID reads an Osquery saved query using Kibana's saved_object_id.
func GetOsquerySavedQueryBySavedObjectID(
	ctx context.Context,
	client *Client,
	spaceID string,
	savedObjectID kbapi.SecurityOsqueryAPISavedQueryId,
) (*OsquerySavedQueryGetEntity, diag.Diagnostics) {
	resp, err := client.API.OsqueryGetSavedQueryDetailsWithResponse(ctx, savedObjectID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return HandleGetTypedResponse(resp.StatusCode(), resp.Body,
		func() *OsquerySavedQueryGetEntity {
			return osquerySavedQueryGetEntityFrom(resp.JSON200)
		})
}

// GetOsquerySavedQuery reads an Osquery saved query by user-facing saved_query_id.
// It resolves saved_object_id via OsqueryFindSavedQueries, then GETs details by saved object ID.
// Returns (nil, nil) when the saved query is not found.
func GetOsquerySavedQuery(
	ctx context.Context,
	client *Client,
	spaceID string,
	savedQueryID kbapi.SecurityOsqueryAPISavedQueryId,
) (*OsquerySavedQueryGetEntity, diag.Diagnostics) {
	savedObjectID, found, diags := FindOsquerySavedObjectID(ctx, client, spaceID, savedQueryID)
	if diags.HasError() {
		return nil, diags
	}
	if !found {
		return nil, diags
	}

	return GetOsquerySavedQueryBySavedObjectID(ctx, client, spaceID, savedObjectID)
}

// UpdateOsquerySavedQueryBySavedObjectID updates an Osquery saved query using Kibana's saved_object_id.
func UpdateOsquerySavedQueryBySavedObjectID(
	ctx context.Context,
	client *Client,
	spaceID string,
	savedObjectID kbapi.SecurityOsqueryAPISavedQueryId,
	body kbapi.OsqueryUpdateSavedQueryJSONRequestBody,
) (*OsquerySavedQueryUpdateEntity, diag.Diagnostics) {
	resp, err := client.API.OsqueryUpdateSavedQueryWithResponse(ctx, savedObjectID, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *OsquerySavedQueryUpdateEntity {
			return osquerySavedQueryUpdateEntityFrom(resp.JSON200)
		})
}

// UpdateOsquerySavedQuery updates an Osquery saved query via PUT /api/osquery/saved_queries/{saved_object_id}.
// The path parameter is the Kibana saved object ID resolved from saved_query_id.
func UpdateOsquerySavedQuery(
	ctx context.Context,
	client *Client,
	spaceID string,
	savedQueryID kbapi.SecurityOsqueryAPISavedQueryId,
	body kbapi.OsqueryUpdateSavedQueryJSONRequestBody,
) (*OsquerySavedQueryUpdateEntity, diag.Diagnostics) {
	savedObjectID, found, diags := FindOsquerySavedObjectID(ctx, client, spaceID, savedQueryID)
	if diags.HasError() {
		return nil, diags
	}
	if !found {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Osquery saved query not found",
				"Cannot update Osquery saved query: no saved query with the configured saved_query_id exists in the target Kibana space.",
			),
		}
	}

	return UpdateOsquerySavedQueryBySavedObjectID(ctx, client, spaceID, savedObjectID, body)
}

// DeleteOsquerySavedQueryBySavedObjectID deletes an Osquery saved query using Kibana's saved_object_id.
func DeleteOsquerySavedQueryBySavedObjectID(
	ctx context.Context,
	client *Client,
	spaceID string,
	savedObjectID kbapi.SecurityOsqueryAPISavedQueryId,
) diag.Diagnostics {
	resp, err := client.API.OsqueryDeleteSavedQueryWithResponse(ctx, savedObjectID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}

// DeleteOsquerySavedQuery deletes an Osquery saved query via DELETE /api/osquery/saved_queries/{saved_object_id}.
// The path parameter is the Kibana saved object ID resolved from saved_query_id.
// HTTP 404 is treated as success (idempotent delete). When the query is already absent from the find
// list, delete is treated as success.
func DeleteOsquerySavedQuery(
	ctx context.Context,
	client *Client,
	spaceID string,
	savedQueryID kbapi.SecurityOsqueryAPISavedQueryId,
) diag.Diagnostics {
	savedObjectID, found, diags := FindOsquerySavedObjectID(ctx, client, spaceID, savedQueryID)
	if diags.HasError() {
		return diags
	}
	if !found {
		return nil
	}

	return DeleteOsquerySavedQueryBySavedObjectID(ctx, client, spaceID, savedObjectID)
}
