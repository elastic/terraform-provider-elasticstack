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

// OsquerySavedQueryCreateEntity is the unwrapped `data` object from POST /api/osquery/saved_queries.
type OsquerySavedQueryCreateEntity struct {
	CreatedAt           *time.Time
	CreatedBy           *string
	CreatedByProfileUid *string
	Description         *kbapi.SecurityOsqueryAPISavedQueryDescription
	EcsMapping          *kbapi.SecurityOsqueryAPIECSMapping
	Id                  kbapi.SecurityOsqueryAPISavedQueryId
	Interval            *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Interval
	Platform            *kbapi.SecurityOsqueryAPIPlatform
	Prebuilt            *bool
	Query               *kbapi.SecurityOsqueryAPIQuery
	Removed             *kbapi.SecurityOsqueryAPIRemoved
	SavedObjectId       string
	Snapshot            *kbapi.SecurityOsqueryAPISnapshot
	Timeout             *int
	UpdatedAt           *time.Time
	UpdatedBy           *string
	UpdatedByProfileUid *string
	Version             *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Version
}

// OsquerySavedQueryGetEntity is the unwrapped `data` object from GET /api/osquery/saved_queries/{id}.
type OsquerySavedQueryGetEntity struct {
	CreatedAt           *time.Time
	CreatedBy           *string
	CreatedByProfileUid *string
	Description         *kbapi.SecurityOsqueryAPISavedQueryDescription
	EcsMapping          *kbapi.SecurityOsqueryAPIECSMapping
	Id                  kbapi.SecurityOsqueryAPISavedQueryId
	Interval            *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Interval
	Platform            *kbapi.SecurityOsqueryAPIPlatform
	Prebuilt            *bool
	Query               *kbapi.SecurityOsqueryAPIQuery
	Removed             *kbapi.SecurityOsqueryAPIRemoved
	SavedObjectId       string
	Snapshot            *kbapi.SecurityOsqueryAPISnapshot
	Timeout             *int
	UpdatedAt           *time.Time
	UpdatedBy           *string
	UpdatedByProfileUid *string
	Version             *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Version
}

// OsquerySavedQueryUpdateEntity is the unwrapped `data` object from PUT /api/osquery/saved_queries/{id}.
type OsquerySavedQueryUpdateEntity struct {
	CreatedAt           *time.Time
	CreatedBy           *string
	CreatedByProfileUid *string
	Description         *kbapi.SecurityOsqueryAPISavedQueryDescription
	EcsMapping          *kbapi.SecurityOsqueryAPIECSMapping
	Id                  kbapi.SecurityOsqueryAPISavedQueryId
	Interval            *kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse_Data_Interval
	Platform            *kbapi.SecurityOsqueryAPIPlatform
	Prebuilt            *bool
	Query               *kbapi.SecurityOsqueryAPIQuery
	Removed             *kbapi.SecurityOsqueryAPIRemoved
	SavedObjectId       string
	Snapshot            *kbapi.SecurityOsqueryAPISnapshot
	Timeout             *int
	UpdatedAt           *time.Time
	UpdatedBy           *string
	UpdatedByProfileUid *string
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
		CreatedByProfileUid: d.CreatedByProfileUid,
		Description:         d.Description,
		EcsMapping:          d.EcsMapping,
		Id:                  d.Id,
		Interval:            d.Interval,
		Platform:            d.Platform,
		Prebuilt:            d.Prebuilt,
		Query:               d.Query,
		Removed:             d.Removed,
		SavedObjectId:       d.SavedObjectId,
		Snapshot:            d.Snapshot,
		Timeout:             d.Timeout,
		UpdatedAt:           d.UpdatedAt,
		UpdatedBy:           d.UpdatedBy,
		UpdatedByProfileUid: d.UpdatedByProfileUid,
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
		CreatedByProfileUid: d.CreatedByProfileUid,
		Description:         d.Description,
		EcsMapping:          d.EcsMapping,
		Id:                  d.Id,
		Interval:            d.Interval,
		Platform:            d.Platform,
		Prebuilt:            d.Prebuilt,
		Query:               d.Query,
		Removed:             d.Removed,
		SavedObjectId:       d.SavedObjectId,
		Snapshot:            d.Snapshot,
		Timeout:             d.Timeout,
		UpdatedAt:           d.UpdatedAt,
		UpdatedBy:           d.UpdatedBy,
		UpdatedByProfileUid: d.UpdatedByProfileUid,
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
		CreatedByProfileUid: d.CreatedByProfileUid,
		Description:         d.Description,
		EcsMapping:          d.EcsMapping,
		Id:                  d.Id,
		Interval:            d.Interval,
		Platform:            d.Platform,
		Prebuilt:            d.Prebuilt,
		Query:               d.Query,
		Removed:             d.Removed,
		SavedObjectId:       d.SavedObjectId,
		Snapshot:            d.Snapshot,
		Timeout:             d.Timeout,
		UpdatedAt:           d.UpdatedAt,
		UpdatedBy:           d.UpdatedBy,
		UpdatedByProfileUid: d.UpdatedByProfileUid,
		Version:             d.Version,
	}
}

// CreateOsquerySavedQuery creates a new Osquery saved query via POST /api/osquery/saved_queries.
func CreateOsquerySavedQuery(ctx context.Context, client *Client, spaceID string, body kbapi.OsqueryCreateSavedQueryJSONRequestBody) (*OsquerySavedQueryCreateEntity, diag.Diagnostics) {
	resp, err := client.API.OsqueryCreateSavedQueryWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *OsquerySavedQueryCreateEntity {
			return osquerySavedQueryCreateEntityFrom(resp.JSON200)
		})
}

// GetOsquerySavedQuery reads an Osquery saved query by ID via GET /api/osquery/saved_queries/{id}.
// Returns (nil, nil) when the saved query is not found (HTTP 404).
func GetOsquerySavedQuery(ctx context.Context, client *Client, spaceID string, savedQueryID kbapi.SecurityOsqueryAPISavedQueryId) (*OsquerySavedQueryGetEntity, diag.Diagnostics) {
	resp, err := client.API.OsqueryGetSavedQueryDetailsWithResponse(ctx, savedQueryID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return HandleGetTypedResponse(resp.StatusCode(), resp.Body,
		func() *OsquerySavedQueryGetEntity {
			return osquerySavedQueryGetEntityFrom(resp.JSON200)
		})
}

// UpdateOsquerySavedQuery updates an Osquery saved query via PUT /api/osquery/saved_queries/{id}.
func UpdateOsquerySavedQuery(ctx context.Context, client *Client, spaceID string, savedQueryID kbapi.SecurityOsqueryAPISavedQueryId, body kbapi.OsqueryUpdateSavedQueryJSONRequestBody) (*OsquerySavedQueryUpdateEntity, diag.Diagnostics) {
	resp, err := client.API.OsqueryUpdateSavedQueryWithResponse(ctx, savedQueryID, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *OsquerySavedQueryUpdateEntity {
			return osquerySavedQueryUpdateEntityFrom(resp.JSON200)
		})
}

// DeleteOsquerySavedQuery deletes an Osquery saved query via DELETE /api/osquery/saved_queries/{id}.
// HTTP 404 is treated as success (idempotent delete).
func DeleteOsquerySavedQuery(ctx context.Context, client *Client, spaceID string, savedQueryID kbapi.SecurityOsqueryAPISavedQueryId) diag.Diagnostics {
	resp, err := client.API.OsqueryDeleteSavedQueryWithResponse(ctx, savedQueryID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
