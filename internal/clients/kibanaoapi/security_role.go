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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// SecurityRoleESIndex represents an index entry in the elasticsearch.indices section of a Kibana role.
type SecurityRoleESIndex struct {
	AllowRestrictedIndices *bool                `json:"allow_restricted_indices,omitempty"`
	FieldSecurity          *map[string][]string `json:"field_security,omitempty"`
	Names                  []string             `json:"names"`
	Privileges             []string             `json:"privileges"`
	Query                  *string              `json:"query,omitempty"`
}

// SecurityRoleESRemoteIndex represents an entry in the elasticsearch.remote_indices section of a Kibana role.
type SecurityRoleESRemoteIndex struct {
	AllowRestrictedIndices *bool                `json:"allow_restricted_indices,omitempty"`
	Clusters               []string             `json:"clusters"`
	FieldSecurity          *map[string][]string `json:"field_security,omitempty"`
	Names                  []string             `json:"names"`
	Privileges             []string             `json:"privileges"`
	Query                  *string              `json:"query,omitempty"`
}

// SecurityRoleES holds the elasticsearch section of a Kibana role.
type SecurityRoleES struct {
	Cluster       *[]string                    `json:"cluster,omitempty"`
	Indices       *[]SecurityRoleESIndex       `json:"indices,omitempty"`
	RemoteIndices *[]SecurityRoleESRemoteIndex `json:"remote_indices,omitempty"`
	RunAs         *[]string                    `json:"run_as,omitempty"`
}

// SecurityRoleKibana holds one entry from the kibana section of a Kibana role.
// Base is kept as json.RawMessage to handle both []string and null gracefully,
// since the generated kbapi union type cannot unmarshal from the Kibana API response.
type SecurityRoleKibana struct {
	Base    json.RawMessage      `json:"base,omitempty"`
	Feature *map[string][]string `json:"feature,omitempty"`
	Spaces  *[]string            `json:"spaces,omitempty"`
}

// SecurityRole is a decodable representation of a Kibana role as returned by GET
// /api/security/role/{name}. It mirrors the PUT request body shape but uses
// named exported struct types and json.RawMessage for the kibana.base field.
type SecurityRole struct {
	Description   *string              `json:"description,omitempty"`
	Elasticsearch SecurityRoleES       `json:"elasticsearch"`
	Kibana        []SecurityRoleKibana `json:"kibana,omitempty"`
	Metadata      *map[string]any      `json:"metadata,omitempty"`
}

// SecurityRolePutBody is the request body for creating or updating a Kibana role.
// It uses plain Go types for all fields (no kbapi union structs) so that JSON
// marshaling produces the correct wire format.
type SecurityRolePutBody struct {
	Description   *string              `json:"description,omitempty"`
	Elasticsearch SecurityRoleES       `json:"elasticsearch"`
	Kibana        []SecurityRoleKibana `json:"kibana,omitempty"`
	Metadata      *map[string]any      `json:"metadata,omitempty"`
}

// GetSecurityRole retrieves a Kibana security role by name.
// Returns (nil, nil) when the role is not found (HTTP 404).
// Returns (nil, diags) on error.
func GetSecurityRole(ctx context.Context, client *Client, name string) (*SecurityRole, sdkdiag.Diagnostics) {
	params := &kbapi.GetSecurityRoleNameParams{}
	resp, err := client.API.GetSecurityRoleNameWithResponse(ctx, name, params)
	if err != nil {
		return nil, sdkdiag.Diagnostics{
			sdkdiag.Diagnostic{
				Severity: sdkdiag.Error,
				Summary:  "Failed to read Kibana security role",
				Detail:   err.Error(),
			},
		}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var role SecurityRole
		if err := json.Unmarshal(resp.Body, &role); err != nil {
			return nil, sdkdiag.Diagnostics{
				sdkdiag.Diagnostic{
					Severity: sdkdiag.Error,
					Summary:  "Failed to parse Kibana security role response",
					Detail:   fmt.Sprintf("JSON decode error: %s. Body: %s", err.Error(), string(resp.Body)),
				},
			}
		}
		return &role, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}
}

// PutSecurityRole creates or updates a Kibana security role using the supplied body.
// It uses the WithBody variant so that the kbapi union type for kibana.base is bypassed
// and the body is sent as-is.
func PutSecurityRole(ctx context.Context, client *Client, name string, params kbapi.PutSecurityRoleNameParams, body SecurityRolePutBody) sdkdiag.Diagnostics {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return sdkdiag.Diagnostics{
			sdkdiag.Diagnostic{
				Severity: sdkdiag.Error,
				Summary:  "Failed to serialize Kibana security role",
				Detail:   err.Error(),
			},
		}
	}

	resp, err := client.API.PutSecurityRoleNameWithBodyWithResponse(ctx, name, &params, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return sdkdiag.Diagnostics{
			sdkdiag.Diagnostic{
				Severity: sdkdiag.Error,
				Summary:  "Failed to write Kibana security role",
				Detail:   err.Error(),
			},
		}
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		return reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}
}

// DeleteSecurityRole deletes a Kibana security role by name.
func DeleteSecurityRole(ctx context.Context, client *Client, name string) sdkdiag.Diagnostics {
	resp, err := client.API.DeleteSecurityRoleNameWithResponse(ctx, name)
	if err != nil {
		return sdkdiag.Diagnostics{
			sdkdiag.Diagnostic{
				Severity: sdkdiag.Error,
				Summary:  "Failed to delete Kibana security role",
				Detail:   err.Error(),
			},
		}
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}
}
