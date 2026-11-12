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
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// addInternalOriginEditor identifies the request as internal. Kibana serves the
// advanced settings routes as internal routes and, from 9.0, rejects requests to
// them that lack this header.
func addInternalOriginEditor() func(ctx context.Context, req *http.Request) error {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("x-elastic-internal-origin", "Kibana")
		return nil
	}
}

type advancedSettingsResponse struct {
	Settings map[string]struct {
		UserValue json.RawMessage `json:"userValue"`
	} `json:"settings"`
}

// GetAdvancedSettings returns the settings a user has set in the space, keyed by
// setting name, with each value as the JSON Kibana returned.
// Returns (nil, nil) when Kibana responds with HTTP 404.
func GetAdvancedSettings(ctx context.Context, client *Client, spaceID string) (map[string]json.RawMessage, diag.Diagnostics) {
	const summary = "Failed to read Kibana advanced settings"

	resp, err := client.API.GetKibanaSettingsWithResponse(ctx,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
		addInternalOriginEditor(),
	)
	if err != nil {
		return nil, diagutil.ErrDiag(summary, err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var body advancedSettingsResponse
		if err := json.Unmarshal(resp.Body, &body); err != nil {
			return nil, diagutil.ErrDiag(summary, err)
		}
		settings := make(map[string]json.RawMessage, len(body.Settings))
		for key, entry := range body.Settings {
			settings[key] = entry.UserValue
		}
		return settings, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportKibanaBoomHTTPError(resp.StatusCode(), summary, resp.Body)
	}
}

// ApplyAdvancedSettings changes settings in the space with one bulk request. A
// nil value resets the setting to its Kibana default. The returned bool is true
// when Kibana responds with HTTP 404.
func ApplyAdvancedSettings(ctx context.Context, client *Client, spaceID string, changes map[string]any) (bool, diag.Diagnostics) {
	const summary = "Failed to apply Kibana advanced settings"

	resp, err := client.API.PostKibanaSettingsWithResponse(ctx,
		kbapi.PostKibanaSettingsJSONRequestBody{Changes: changes},
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
		addInternalOriginEditor(),
	)
	if err != nil {
		return false, diagutil.ErrDiag(summary, err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return false, nil
	case http.StatusNotFound:
		return true, nil
	default:
		return false, diagutil.ReportKibanaBoomHTTPError(resp.StatusCode(), summary, resp.Body)
	}
}
