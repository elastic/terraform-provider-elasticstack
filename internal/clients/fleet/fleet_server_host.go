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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetFleetServerHost reads a specific fleet server host from the API.
func GetFleetServerHost(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.GetFleetFleetServerHostsItemidWithResponse(ctx, id, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// CreateFleetServerHost creates a new fleet server host.
func CreateFleetServerHost(ctx context.Context, client *Client, spaceID string, req kbapi.PostFleetFleetServerHostsJSONRequestBody) (*kbapi.ServerHost, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.ServerHost, int, diag.Diagnostics) {
		resp, err := client.API.PostFleetFleetServerHostsWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			return &resp.JSON200.Item, resp.StatusCode(), nil
		default:
			return nil, resp.StatusCode(), diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
		}
	})
}

// UpdateFleetServerHost updates an existing fleet server host.
func UpdateFleetServerHost(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PutFleetFleetServerHostsItemidJSONRequestBody) (*kbapi.ServerHost, diag.Diagnostics) {
	return kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (*kbapi.ServerHost, int, diag.Diagnostics) {
		resp, err := client.API.PutFleetFleetServerHostsItemidWithResponse(ctx, id, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return nil, 0, diagutil.FrameworkDiagFromError(err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			return &resp.JSON200.Item, resp.StatusCode(), nil
		default:
			return nil, resp.StatusCode(), diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
		}
	})
}

// DeleteFleetServerHost deletes an existing fleet server host.
func DeleteFleetServerHost(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	_, diags := kibanautil.ConflictRetry(ctx, kibanautil.ConflictMaxAttempts, func() (struct{}, int, diag.Diagnostics) {
		resp, err := client.API.DeleteFleetFleetServerHostsItemidWithResponse(ctx, id, kibanautil.SpaceAwarePathRequestEditor(spaceID))
		if err != nil {
			return struct{}{}, 0, diagutil.FrameworkDiagFromError(err)
		}
		return struct{}{}, resp.StatusCode(), handleDeleteResponse(resp.StatusCode(), resp.Body)
	})
	return diags
}
