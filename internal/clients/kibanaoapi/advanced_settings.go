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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// internalOriginHeader marks a request as originating from Kibana itself.
const internalOriginHeader = "x-elastic-internal-origin"

// withInternalOrigin sets the internal origin header on advanced settings
// requests. Kibana registers /api/kibana/settings and
// /api/kibana/global_settings without public access, and Kibana 9.0+ rejects
// requests to such routes (server.restrictInternalApis defaults to true)
// unless they carry this header. Remove once Kibana exposes these routes
// publicly (https://github.com/elastic/kibana/issues/279134).
func withInternalOrigin(_ context.Context, req *http.Request) error {
	req.Header.Set(internalOriginHeader, "Kibana")
	return nil
}

// AdvancedSetting is a user-provided Kibana advanced setting value.
type AdvancedSetting struct {
	Value        any
	IsOverridden bool
}

// advancedSettingsBody mirrors the response body shared by the space and
// global advanced settings endpoints.
type advancedSettingsBody = struct {
	Settings map[string]struct {
		IsOverridden *bool `json:"isOverridden,omitempty"`
		UserValue    any   `json:"userValue,omitempty"`
	} `json:"settings"`
}

func advancedSettingsFromBody(body *advancedSettingsBody) map[string]AdvancedSetting {
	settings := make(map[string]AdvancedSetting, len(body.Settings))
	for key, setting := range body.Settings {
		settings[key] = AdvancedSetting{
			Value:        setting.UserValue,
			IsOverridden: setting.IsOverridden != nil && *setting.IsOverridden,
		}
	}
	return settings
}

// GetAdvancedSettings returns the user-provided advanced settings for a space,
// or for the global scope when global is true (spaceID is then ignored).
// Returns (nil, nil) when Kibana responds with 404.
func GetAdvancedSettings(ctx context.Context, client *Client, spaceID string, global bool) (map[string]AdvancedSetting, diag.Diagnostics) {
	var (
		statusCode int
		rawBody    []byte
		extract    func() *advancedSettingsBody
	)

	if global {
		resp, err := client.API.GetKibanaGlobalSettingsWithResponse(ctx, withInternalOrigin)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		statusCode, rawBody, extract = resp.StatusCode(), resp.Body, func() *advancedSettingsBody { return resp.JSON200 }
	} else {
		resp, err := client.API.GetKibanaSettingsWithResponse(ctx, spaceID, withInternalOrigin)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		statusCode, rawBody, extract = resp.StatusCode(), resp.Body, func() *advancedSettingsBody { return resp.JSON200 }
	}

	body, diags := HandleGetTypedResponse(statusCode, rawBody, extract)
	if diags.HasError() || body == nil {
		return nil, diags
	}
	return advancedSettingsFromBody(body), diags
}

// UpdateAdvancedSettings applies changes to the advanced settings for a space,
// or for the global scope when global is true (spaceID is then ignored). A nil
// value in changes resets that setting to its Kibana default. The returned map
// holds the user-provided settings reported by Kibana after the update.
func UpdateAdvancedSettings(ctx context.Context, client *Client, spaceID string, global bool, changes map[string]any) (map[string]AdvancedSetting, diag.Diagnostics) {
	var (
		statusCode int
		rawBody    []byte
		extract    func() *advancedSettingsBody
	)

	if global {
		resp, err := client.API.PostKibanaGlobalSettingsWithResponse(ctx, kbapi.PostKibanaGlobalSettingsJSONRequestBody{Changes: changes}, withInternalOrigin)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		statusCode, rawBody, extract = resp.StatusCode(), resp.Body, func() *advancedSettingsBody { return resp.JSON200 }
	} else {
		resp, err := client.API.PostKibanaSettingsWithResponse(ctx, spaceID, kbapi.PostKibanaSettingsJSONRequestBody{Changes: changes}, withInternalOrigin)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		statusCode, rawBody, extract = resp.StatusCode(), resp.Body, func() *advancedSettingsBody { return resp.JSON200 }
	}

	body, diags := HandleMutateTypedResponse(statusCode, rawBody, extract)
	if diags.HasError() {
		if global && statusCode == http.StatusInternalServerError {
			diags.AddError(
				"Failed to update global advanced settings",
				"Kibana responds with an internal server error when a setting is not registered as a global setting. "+
					"Check that every setting is a global advanced setting, or manage it per space instead.",
			)
		}
		return nil, diags
	}
	return advancedSettingsFromBody(body), diags
}
