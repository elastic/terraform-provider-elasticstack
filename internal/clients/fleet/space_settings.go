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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SpaceSettings holds the Fleet settings of a single Kibana space.
type SpaceSettings struct {
	AllowedNamespacePrefixes []string
	ManagedBy                *string
}

// GetSpaceSettings reads the Fleet settings of a space. Returns (nil, nil) on HTTP 404.
func GetSpaceSettings(ctx context.Context, client *Client, spaceID string) (*SpaceSettings, diag.Diagnostics) {
	resp, err := client.API.GetFleetSpaceSettingsWithResponse(ctx, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return kibanaoapi.HandleGetTypedResponse(resp.StatusCode(), resp.Body, func() *SpaceSettings {
		if resp.JSON200 == nil {
			return nil
		}
		return &SpaceSettings{
			AllowedNamespacePrefixes: resp.JSON200.Item.AllowedNamespacePrefixes,
			ManagedBy:                resp.JSON200.Item.ManagedBy,
		}
	})
}

// PutSpaceSettings writes the allowed namespace prefixes of a space.
func PutSpaceSettings(ctx context.Context, client *Client, spaceID string, allowedNamespacePrefixes []string) diag.Diagnostics {
	return putSpaceSettings(ctx, client, spaceID, allowedNamespacePrefixes, http.StatusOK)
}

// ResetSpaceSettings clears the allowed namespace prefixes of a space. HTTP 404 counts as success.
func ResetSpaceSettings(ctx context.Context, client *Client, spaceID string) diag.Diagnostics {
	return putSpaceSettings(ctx, client, spaceID, []string{}, http.StatusOK, http.StatusNotFound)
}

func putSpaceSettings(ctx context.Context, client *Client, spaceID string, allowedNamespacePrefixes []string, successCodes ...int) diag.Diagnostics {
	body := kbapi.PutFleetSpaceSettingsJSONRequestBody{AllowedNamespacePrefixes: &allowedNamespacePrefixes}
	resp, err := client.API.PutFleetSpaceSettingsWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, successCodes...)
}
