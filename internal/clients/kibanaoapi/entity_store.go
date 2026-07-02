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
	"io"
	"net/http"

	kbapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

// CreateSecurityEntityStoreEntity creates a single entity record.
func CreateSecurityEntityStoreEntity(ctx context.Context, client *Client, spaceID string, entityType string, body io.Reader) diag.Diagnostics {
	statusCode, respBody, err := CreateSecurityEntityStoreEntityStatus(ctx, client, spaceID, entityType, body)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(statusCode, respBody, http.StatusOK)
}

// CreateSecurityEntityStoreEntityStatus creates a single entity record and
// returns the raw HTTP status code and response body without collapsing them
// into diagnostics. This lets callers implement status-specific behavior such
// as retrying on HTTP 500 while the entity store is still initializing.
func CreateSecurityEntityStoreEntityStatus(ctx context.Context, client *Client, spaceID string, entityType string, body io.Reader) (int, []byte, error) {
	resp, err := client.API.PostSecurityEntityStoreEntitiesEntitytypeWithBodyWithResponse(
		ctx,
		kbapi.PostSecurityEntityStoreEntitiesEntitytypeParamsEntityType(entityType),
		"application/json",
		body,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode(), resp.Body, nil
}

// UpdateSecurityEntityStoreEntity updates a single entity record.
func UpdateSecurityEntityStoreEntity(ctx context.Context, client *Client, spaceID string, entityType string, body io.Reader, force bool) diag.Diagnostics {
	var editors []kbapi.RequestEditorFn
	editors = append(editors, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if force {
		editors = append(editors, func(_ context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Set("force", "true")
			req.URL.RawQuery = q.Encode()
			return nil
		})
	}

	resp, err := client.API.PutSecurityEntityStoreEntitiesEntitytypeWithBodyWithResponse(
		ctx,
		kbapi.PutSecurityEntityStoreEntitiesEntitytypeParamsEntityType(entityType),
		nil, // params – force is handled via request editor
		"application/json",
		body,
		editors...,
	)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// DeleteSecurityEntityStoreEntity deletes a single entity record.
func DeleteSecurityEntityStoreEntity(ctx context.Context, client *Client, spaceID string, body io.Reader) diag.Diagnostics {
	resp, err := client.API.DeleteSecurityEntityStoreEntitiesWithBodyWithResponse(
		ctx,
		"application/json",
		body,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// ListSecurityEntityStoreEntities lists/search entity records.
func ListSecurityEntityStoreEntities(
	ctx context.Context,
	client *Client,
	spaceID string,
	params *kbapi.GetSecurityEntityStoreEntitiesParams,
) (*kbapi.GetSecurityEntityStoreEntitiesResponse, diag.Diagnostics) {
	resp, err := client.API.GetSecurityEntityStoreEntitiesWithResponse(
		ctx,
		params,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if d := diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK); d.HasError() {
		return nil, d
	}
	return resp, nil
}

// GetSecurityEntityStoreStatus reads the entity store status. The caller is responsible for unmarshaling the response body.
func GetSecurityEntityStoreStatus(ctx context.Context, client *Client, spaceID string, includeComponents bool) (*kbapi.GetSecurityEntityStoreStatusResponse, diag.Diagnostics) {
	var editors []kbapi.RequestEditorFn
	if includeComponents {
		editors = append(editors, func(_ context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Set("include_components", "true")
			req.URL.RawQuery = q.Encode()
			return nil
		})
	}

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
