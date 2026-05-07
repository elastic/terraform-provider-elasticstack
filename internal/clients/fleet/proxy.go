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

// GetProxy reads a specific Fleet proxy from the API. Returns (nil, nil) on HTTP 404.
func GetProxy(ctx context.Context, client *Client, spaceID, proxyID string) (*kbapi.FleetProxyItem, diag.Diagnostics) {
	resp, err := client.API.GetFleetProxiesItemidWithResponse(ctx, proxyID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
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

// CreateProxy creates a new Fleet proxy.
func CreateProxy(ctx context.Context, client *Client, spaceID string, body kbapi.PostFleetProxiesJSONRequestBody) (*kbapi.FleetProxyItem, diag.Diagnostics) {
	resp, err := client.API.PostFleetProxiesWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// UpdateProxy updates an existing Fleet proxy.
func UpdateProxy(ctx context.Context, client *Client, spaceID, proxyID string, body kbapi.PutFleetProxiesItemidJSONRequestBody) (*kbapi.FleetProxyItem, diag.Diagnostics) {
	resp, err := client.API.PutFleetProxiesItemidWithResponse(ctx, proxyID, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// DeleteProxy deletes an existing Fleet proxy.
func DeleteProxy(ctx context.Context, client *Client, spaceID, proxyID string) diag.Diagnostics {
	resp, err := client.API.DeleteFleetProxiesItemidWithResponse(ctx, proxyID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleDeleteResponse(resp.StatusCode(), resp.Body)
}
