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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// InstallSecurityEntityStore installs the entity store with the given types and options.
func InstallSecurityEntityStore(ctx context.Context, client *Client, spaceID string, body kbapi.PostSecurityEntityStoreInstallJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PostSecurityEntityStoreInstallWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusCreated)
}

// UpdateSecurityEntityStore updates the entity store log extraction configuration.
func UpdateSecurityEntityStore(ctx context.Context, client *Client, spaceID string, body kbapi.PutSecurityEntityStoreJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PutSecurityEntityStoreWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// UninstallSecurityEntityStore uninstalls the given entity types from the store.
func UninstallSecurityEntityStore(ctx context.Context, client *Client, spaceID string, body kbapi.PostSecurityEntityStoreUninstallJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PostSecurityEntityStoreUninstallWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// StartSecurityEntityStore starts the entity store engines.
func StartSecurityEntityStore(ctx context.Context, client *Client, spaceID string, body kbapi.PutSecurityEntityStoreStartJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PutSecurityEntityStoreStartWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// StopSecurityEntityStore stops the entity store engines.
func StopSecurityEntityStore(ctx context.Context, client *Client, spaceID string, body kbapi.PutSecurityEntityStoreStopJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PutSecurityEntityStoreStopWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// GetSecurityEntityStoreStatus reads the entity store status. The caller is responsible for unmarshaling the response body.
func GetSecurityEntityStoreStatus(ctx context.Context, client *Client, spaceID string, editors ...kbapi.RequestEditorFn) (*kbapi.GetSecurityEntityStoreStatusResponse, diag.Diagnostics) {
	allEditors := append([]kbapi.RequestEditorFn{kibanautil.SpaceAwarePathRequestEditor(spaceID)}, editors...)
	resp, err := client.API.GetSecurityEntityStoreStatusWithResponse(ctx, &kbapi.GetSecurityEntityStoreStatusParams{}, allEditors...)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if d := diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK); d.HasError() {
		return nil, d
	}
	return resp, nil
}
