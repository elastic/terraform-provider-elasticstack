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
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// OsqueryPackShards is the normalized shard map for Terraform state (percent 1-100 per policy ID).
type OsqueryPackShards map[string]float64

// OsqueryPackDetail is the unwrapped pack payload from GET/Create/Update detail responses.
// Shards are normalized to map[string]float64 regardless of API wire format.
// GetOsqueryPack detail is authoritative for read-only and other GET-only fields.
// CreateOsqueryPack and UpdateOsqueryPack return non-authoritative detail; resource CRUD
// must follow with read-after-write GetOsqueryPack for full detail and the prebuilt guard.
type OsqueryPackDetail struct {
	CreatedAt           *time.Time
	CreatedBy           *string
	CreatedByProfileUID *string
	Description         *kbapi.SecurityOsqueryAPIPackDescription
	Enabled             *kbapi.SecurityOsqueryAPIEnabled
	Name                kbapi.SecurityOsqueryAPIPackName
	Namespaces          *[]string
	PolicyIDs           *kbapi.SecurityOsqueryAPIPolicyIds
	Queries             *kbapi.SecurityOsqueryAPIObjectQueries
	ReadOnly            *bool
	SavedObjectID       string
	Shards              OsqueryPackShards
	Type                *string
	UpdatedAt           *time.Time
	UpdatedBy           *string
	UpdatedByProfileUID *string
	Version             *int
}

// CreateOsqueryPack creates a new Osquery pack via the OpenAPI client.
// The returned detail is not authoritative final state (Create omits read_only and may
// return array-form shards). Callers must read-after-write via GetOsqueryPack.
func CreateOsqueryPack(ctx context.Context, client *Client, spaceID string, body kbapi.OsqueryCreatePacksJSONRequestBody) (*OsqueryPackDetail, diag.Diagnostics) {
	resp, err := client.API.OsqueryCreatePacksWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag("HTTP request failed creating osquery pack", err)
	}

	createResp, diags := HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.SecurityOsqueryAPICreatePacksResponse { return resp.JSON200 })
	if diags.HasError() {
		return nil, diags
	}

	return osqueryPackDetailFromCreateResponse(createResp), nil
}

// GetOsqueryPack reads an Osquery pack by saved_object_id.
// Returns (nil, nil) when the pack is not found (HTTP 404) so the caller
// can remove the resource from state without treating absence as an error.
func GetOsqueryPack(ctx context.Context, client *Client, spaceID string, packID string) (*OsqueryPackDetail, diag.Diagnostics) {
	resp, err := client.API.OsqueryGetPacksDetailsWithResponse(ctx, packID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag(fmt.Sprintf("HTTP request failed reading osquery pack %q", packID), err)
	}

	findResp, diags := HandleGetTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.SecurityOsqueryAPIFindPackResponse { return resp.JSON200 })
	if diags.HasError() || findResp == nil {
		return nil, diags
	}

	return osqueryPackDetailFromFindResponse(findResp), nil
}

// UpdateOsqueryPack updates an existing Osquery pack.
// The returned detail is not authoritative final state. Callers must read-after-write
// via GetOsqueryPack for full detail and the prebuilt guard.
func UpdateOsqueryPack(ctx context.Context, client *Client, spaceID string, packID string, body kbapi.OsqueryUpdatePacksJSONRequestBody) (*OsqueryPackDetail, diag.Diagnostics) {
	resp, err := client.API.OsqueryUpdatePacksWithResponse(ctx, packID, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag(fmt.Sprintf("HTTP request failed updating osquery pack %q", packID), err)
	}

	updateResp, diags := HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.SecurityOsqueryAPIUpdatePacksResponse { return resp.JSON200 })
	if diags.HasError() {
		return nil, diags
	}
	if updateResp.Data == nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to parse response",
				"API returned success status but update response data was nil",
			),
		}
	}

	return osqueryPackDetailFromUpdateResponse(updateResp), nil
}

// DeleteOsqueryPack deletes an Osquery pack by saved_object_id.
// HTTP 404 is treated as success (already deleted).
func DeleteOsqueryPack(ctx context.Context, client *Client, spaceID string, packID string) diag.Diagnostics {
	resp, err := client.API.OsqueryDeletePacksWithResponse(ctx, packID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.ErrDiag(fmt.Sprintf("HTTP request failed deleting osquery pack %q", packID), err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}

func osqueryPackDetailFromFindResponse(resp *kbapi.SecurityOsqueryAPIFindPackResponse) *OsqueryPackDetail {
	if resp == nil {
		return nil
	}

	data := resp.Data
	return &OsqueryPackDetail{
		CreatedAt:           data.CreatedAt,
		CreatedBy:           data.CreatedBy,
		CreatedByProfileUID: data.CreatedByProfileUid,
		Description:         data.Description,
		Enabled:             data.Enabled,
		Name:                data.Name,
		Namespaces:          data.Namespaces,
		PolicyIDs:           data.PolicyIds,
		Queries:             data.Queries,
		ReadOnly:            data.ReadOnly,
		SavedObjectID:       data.SavedObjectId,
		Shards:              osqueryPackShardsFromMap(data.Shards),
		Type:                data.Type,
		UpdatedAt:           data.UpdatedAt,
		UpdatedBy:           data.UpdatedBy,
		UpdatedByProfileUID: data.UpdatedByProfileUid,
		Version:             data.Version,
	}
}

func osqueryPackDetailFromCreateResponse(resp *kbapi.SecurityOsqueryAPICreatePacksResponse) *OsqueryPackDetail {
	if resp == nil {
		return nil
	}

	data := resp.Data
	return &OsqueryPackDetail{
		CreatedAt:           data.CreatedAt,
		CreatedBy:           data.CreatedBy,
		CreatedByProfileUID: data.CreatedByProfileUid,
		Description:         data.Description,
		Enabled:             data.Enabled,
		Name:                data.Name,
		PolicyIDs:           data.PolicyIds,
		Queries:             data.Queries,
		SavedObjectID:       data.SavedObjectId,
		Shards:              osqueryPackShardsFromCreateArray(data.Shards),
		UpdatedAt:           data.UpdatedAt,
		UpdatedBy:           data.UpdatedBy,
		UpdatedByProfileUID: data.UpdatedByProfileUid,
		Version:             data.Version,
	}
}

func osqueryPackDetailFromUpdateResponse(resp *kbapi.SecurityOsqueryAPIUpdatePacksResponse) *OsqueryPackDetail {
	if resp == nil || resp.Data == nil {
		return nil
	}

	data := resp.Data
	detail := &OsqueryPackDetail{
		CreatedAt:           data.CreatedAt,
		CreatedBy:           data.CreatedBy,
		CreatedByProfileUID: data.CreatedByProfileUid,
		Description:         data.Description,
		Enabled:             data.Enabled,
		PolicyIDs:           data.PolicyIds,
		Queries:             data.Queries,
		Shards:              osqueryPackShardsFromMap(data.Shards),
		UpdatedAt:           data.UpdatedAt,
		UpdatedBy:           data.UpdatedBy,
		UpdatedByProfileUID: data.UpdatedByProfileUid,
		Version:             data.Version,
	}
	if data.Name != nil {
		detail.Name = *data.Name
	}
	if data.SavedObjectId != nil {
		detail.SavedObjectID = *data.SavedObjectId
	}
	return detail
}

func osqueryPackShardsFromMap(shards *kbapi.SecurityOsqueryAPIShards) OsqueryPackShards {
	if shards == nil || len(*shards) == 0 {
		return nil
	}

	result := make(OsqueryPackShards, len(*shards))
	for policyID, percent := range *shards {
		result[policyID] = float64(percent)
	}
	return result
}

func osqueryPackShardsFromCreateArray(shards *[]struct {
	Key   *string  `json:"key,omitempty"`
	Value *float32 `json:"value,omitempty"`
}) OsqueryPackShards {
	if shards == nil || len(*shards) == 0 {
		return nil
	}

	result := make(OsqueryPackShards, len(*shards))
	for _, entry := range *shards {
		if entry.Key == nil {
			continue
		}
		var value float64
		if entry.Value != nil {
			value = float64(*entry.Value)
		}
		result[*entry.Key] = value
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
