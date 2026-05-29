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

package fleet

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CloudConnectorItem is the provider-facing representation of a Fleet cloud connector.
// The kbapi client uses anonymous struct types in its responses; this type is populated
// by copying fields out of those response structs.
type CloudConnectorItem struct {
	ID                    string
	Name                  string
	CloudProvider         string
	AccountType           *string
	Namespace             *string
	PackagePolicyCount    float32
	Vars                  map[string]any
	VerificationStatus    *string
	VerificationStartedAt *string
	VerificationFailedAt  *string
	CreatedAt             string
	UpdatedAt             string
}

// GetCloudConnector reads a specific Fleet cloud connector from the API. Returns (nil, nil) on HTTP 404.
func GetCloudConnector(ctx context.Context, client *Client, spaceID, cloudConnectorID string) (*CloudConnectorItem, diag.Diagnostics) {
	resp, err := client.API.GetFleetCloudConnectorsCloudconnectoridWithResponse(ctx, cloudConnectorID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return kibanaoapi.HandleGetTypedResponse(resp.StatusCode(), resp.Body, func() *CloudConnectorItem {
		if resp.JSON200 == nil {
			return nil
		}
		item := cloudConnectorItemFromAPIFields(
			resp.JSON200.Item.Id,
			resp.JSON200.Item.Name,
			resp.JSON200.Item.CloudProvider,
			resp.JSON200.Item.AccountType,
			resp.JSON200.Item.Namespace,
			resp.JSON200.Item.PackagePolicyCount,
			resp.JSON200.Item.Vars,
			resp.JSON200.Item.VerificationStatus,
			resp.JSON200.Item.VerificationStartedAt,
			resp.JSON200.Item.VerificationFailedAt,
			resp.JSON200.Item.CreatedAt,
			resp.JSON200.Item.UpdatedAt,
		)
		return &item
	})
}

// CreateCloudConnector creates a new Fleet cloud connector.
func CreateCloudConnector(ctx context.Context, client *Client, spaceID string, body kbapi.PostFleetCloudConnectorsJSONRequestBody) (*CloudConnectorItem, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*CloudConnectorItem, int, diag.Diagnostics) {
		resp, err := client.API.PostFleetCloudConnectorsWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		result, diags := kibanaoapi.HandleMutateTypedResponse(resp.StatusCode(), resp.Body, func() *CloudConnectorItem {
			if resp.JSON200 == nil {
				return nil
			}
			item := cloudConnectorItemFromAPIFields(
				resp.JSON200.Item.Id,
				resp.JSON200.Item.Name,
				resp.JSON200.Item.CloudProvider,
				resp.JSON200.Item.AccountType,
				resp.JSON200.Item.Namespace,
				resp.JSON200.Item.PackagePolicyCount,
				resp.JSON200.Item.Vars,
				resp.JSON200.Item.VerificationStatus,
				resp.JSON200.Item.VerificationStartedAt,
				resp.JSON200.Item.VerificationFailedAt,
				resp.JSON200.Item.CreatedAt,
				resp.JSON200.Item.UpdatedAt,
			)
			return &item
		})
		return result, resp.StatusCode(), diags
	})
}

// UpdateCloudConnector updates an existing Fleet cloud connector.
func UpdateCloudConnector(ctx context.Context, client *Client, spaceID, cloudConnectorID string, body kbapi.PutFleetCloudConnectorsCloudconnectoridJSONRequestBody) (*CloudConnectorItem, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*CloudConnectorItem, int, diag.Diagnostics) {
		resp, err := client.API.PutFleetCloudConnectorsCloudconnectoridWithResponse(ctx, cloudConnectorID, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		result, diags := kibanaoapi.HandleMutateTypedResponse(resp.StatusCode(), resp.Body, func() *CloudConnectorItem {
			if resp.JSON200 == nil {
				return nil
			}
			item := cloudConnectorItemFromAPIFields(
				resp.JSON200.Item.Id,
				resp.JSON200.Item.Name,
				resp.JSON200.Item.CloudProvider,
				resp.JSON200.Item.AccountType,
				resp.JSON200.Item.Namespace,
				resp.JSON200.Item.PackagePolicyCount,
				resp.JSON200.Item.Vars,
				resp.JSON200.Item.VerificationStatus,
				resp.JSON200.Item.VerificationStartedAt,
				resp.JSON200.Item.VerificationFailedAt,
				resp.JSON200.Item.CreatedAt,
				resp.JSON200.Item.UpdatedAt,
			)
			return &item
		})
		return result, resp.StatusCode(), diags
	})
}

// DeleteCloudConnector deletes an existing Fleet cloud connector. When force is true, the API
// receives ?force=true. HTTP 404 is treated as success (idempotent delete).
//
// Non-success responses (including in-use conflicts when force is false) are surfaced via
// diagutil.ReportUnknownHTTPError with the raw response body. The resource layer (Task 6) can
// pattern-match that body to attach package_policy_count guidance; no sentinel is defined here.
func DeleteCloudConnector(ctx context.Context, client *Client, spaceID, cloudConnectorID string, force bool) diag.Diagnostics {
	params := &kbapi.DeleteFleetCloudConnectorsCloudconnectoridParams{}
	if force {
		forceParam := true
		params.Force = &forceParam
	}

	_, diags := kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (struct{}, int, diag.Diagnostics) {
		resp, err := client.API.DeleteFleetCloudConnectorsCloudconnectoridWithResponse(ctx, cloudConnectorID, params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return struct{}{}, 0, diagutil.FrameworkDiagFromError(err)
		}
		return struct{}{}, resp.StatusCode(), handleDeleteResponse(resp.StatusCode(), resp.Body)
	})
	return diags
}

// ListCloudConnectors lists Fleet cloud connectors, optionally filtered and paginated via params.
func ListCloudConnectors(ctx context.Context, client *Client, spaceID string, params kbapi.GetFleetCloudConnectorsParams) ([]CloudConnectorItem, diag.Diagnostics) {
	resp, err := client.API.GetFleetCloudConnectorsWithResponse(ctx, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	result, diags := kibanaoapi.HandleGetTypedResponse(resp.StatusCode(), resp.Body, func() *[]CloudConnectorItem {
		if resp.JSON200 == nil {
			return nil
		}
		items := make([]CloudConnectorItem, len(resp.JSON200.Items))
		for i, apiItem := range resp.JSON200.Items {
			items[i] = cloudConnectorItemFromAPIFields(
				apiItem.Id,
				apiItem.Name,
				apiItem.CloudProvider,
				apiItem.AccountType,
				apiItem.Namespace,
				apiItem.PackagePolicyCount,
				apiItem.Vars,
				apiItem.VerificationStatus,
				apiItem.VerificationStartedAt,
				apiItem.VerificationFailedAt,
				apiItem.CreatedAt,
				apiItem.UpdatedAt,
			)
		}
		return &items
	})
	if result == nil {
		return nil, diags
	}
	return *result, diags
}

func cloudConnectorItemFromAPIFields(
	id string,
	name string,
	cloudProvider string,
	accountType *string,
	namespace *string,
	packagePolicyCount float32,
	vars map[string]*interface{},
	verificationStatus *string,
	verificationStartedAt *string,
	verificationFailedAt *string,
	createdAt string,
	updatedAt string,
) CloudConnectorItem {
	return CloudConnectorItem{
		ID:                    id,
		Name:                  name,
		CloudProvider:         cloudProvider,
		AccountType:           accountType,
		Namespace:             namespace,
		PackagePolicyCount:    packagePolicyCount,
		Vars:                  varsMapFromAPI(vars),
		VerificationStatus:    verificationStatus,
		VerificationStartedAt: verificationStartedAt,
		VerificationFailedAt:  verificationFailedAt,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}
}

func varsMapFromAPI(vars map[string]*interface{}) map[string]any {
	if vars == nil {
		return nil
	}
	out := make(map[string]any, len(vars))
	for k, v := range vars {
		if v != nil {
			out[k] = *v
		} else {
			out[k] = nil
		}
	}
	return out
}
