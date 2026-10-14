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

package agentdownloadsource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type model struct {
	entitycore.ResourceTimeoutsField
	ID               types.String `tfsdk:"id"`
	SourceID         types.String `tfsdk:"source_id"`
	Name             types.String `tfsdk:"name"`
	Host             types.String `tfsdk:"host"`
	Default          types.Bool   `tfsdk:"default"`
	ProxyID          types.String `tfsdk:"proxy_id"`
	Auth             *authModel   `tfsdk:"auth"`
	SpaceIDs         types.Set    `tfsdk:"space_ids"`         // > string
	KibanaConnection types.List   `tfsdk:"kibana_connection"` // > provider schema connection block
}

type authModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	APIKey   types.String `tfsdk:"api_key"`
	Headers  types.List   `tfsdk:"headers"`
}

type authHeaderModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// downloadSourceAuthHeader is one generated Fleet download-source auth header.
type downloadSourceAuthHeader = struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// downloadSourceAuth is the generated Fleet download-source auth object, shared by
// the create request, update request, and read response.
type downloadSourceAuth = struct {
	ApiKey   *string                     `json:"api_key,omitempty"` //nolint:revive // matches the generated kbapi field name
	Headers  *[]downloadSourceAuthHeader `json:"headers,omitempty"`
	Password *string                     `json:"password,omitempty"`
	Username *string                     `json:"username,omitempty"`
}

var (
	agentDownloadSourceMinVersion     = version.Must(version.NewVersion("8.13.0"))
	agentDownloadSourceAuthMinVersion = version.Must(version.NewVersion("9.4.0"))
)

func (m model) GetID() types.String             { return m.ID }
func (m model) GetResourceID() types.String     { return m.SourceID }
func (m model) GetKibanaConnection() types.List { return m.KibanaConnection }

func (m model) GetSpaceID() types.String {
	return fleet.SpaceIDFromSetOrDefault(m.SpaceIDs, clients.DefaultSpaceID)
}

func (m model) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	reqs := entitycore.SingleVersionRequirement(
		*agentDownloadSourceMinVersion,
		fmt.Sprintf("This resource requires stack version %s or higher.", agentDownloadSourceMinVersion),
	)

	addMinimum := func(configured bool, attr path.Path, minVersion *version.Version, detail string) {
		if !configured {
			return
		}
		reqs = append(reqs, entitycore.NewAttributeVersionRequirement(
			attr,
			*minVersion,
			detail,
		))
	}

	addMinimum(
		m.Auth != nil,
		path.Root("auth"),
		agentDownloadSourceAuthMinVersion,
		fmt.Sprintf("`auth` is not supported on this version; it is only supported in Elastic Stack %s and above", agentDownloadSourceAuthMinVersion),
	)

	return reqs, nil
}

func (m model) toAPICreateModel(ctx context.Context) (kbapi.PostFleetAgentDownloadSourcesJSONRequestBody, diag.Diagnostics) {
	auth, diags := m.authRequest(ctx)
	body := kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{
		Host:      m.Host.ValueString(),
		Name:      m.Name.ValueString(),
		IsDefault: m.Default.ValueBoolPointer(),
		ProxyId:   m.ProxyID.ValueStringPointer(),
		Auth:      auth,
	}

	// The API allows setting a custom id only at creation time.
	if typeutils.IsKnown(m.SourceID) && m.SourceID.ValueString() != "" {
		id := m.SourceID.ValueString()
		body.Id = &id
	}

	return body, diags
}

func (m model) authRequest(ctx context.Context) (*downloadSourceAuth, diag.Diagnostics) {
	if m.Auth == nil {
		return nil, nil
	}

	headers, diags := m.Auth.toAPIHeaders(ctx)
	return &downloadSourceAuth{
		ApiKey:   m.Auth.APIKey.ValueStringPointer(),
		Password: m.Auth.Password.ValueStringPointer(),
		Username: m.Auth.Username.ValueStringPointer(),
		Headers:  headers,
	}, diags
}

func (m authModel) toAPIHeaders(ctx context.Context) (*[]downloadSourceAuthHeader, diag.Diagnostics) {
	if !typeutils.IsKnown(m.Headers) {
		return nil, nil
	}

	var diags diag.Diagnostics
	headerModels := typeutils.ListTypeAs[authHeaderModel](ctx, m.Headers, path.Root("auth").AtName("headers"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	// make returns a non-nil slice even when headerModels is empty, so the
	// request JSON is "headers":[] rather than an omitted field.
	headers := make([]downloadSourceAuthHeader, len(headerModels))
	for i, header := range headerModels {
		headers[i] = downloadSourceAuthHeader{
			Key:   header.Key.ValueString(),
			Value: header.Value.ValueString(),
		}
	}
	return &headers, nil
}

func (m model) toAPIUpdateModel(ctx context.Context, prior model) (kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody, diag.Diagnostics) {
	auth, diags := m.authForUpdate(ctx, prior)
	return kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{
		Host:      m.Host.ValueString(),
		Name:      m.Name.ValueString(),
		IsDefault: m.Default.ValueBoolPointer(),
		ProxyId:   fleet.ProxyIDForUpdate(m.ProxyID, prior.ProxyID),
		Auth:      auth,
	}, diags
}

// authForUpdate builds the update body's auth field. A planned auth block is
// sent in full. Removing a previously configured auth block sends an empty
// object: the generated `json:"auth,omitempty"` tag drops nil, and Fleet
// treats an omitted field as a merge that leaves the prior credentials in
// place. When auth was already absent, nil is returned so the field stays omitted.
func (m model) authForUpdate(ctx context.Context, prior model) (*downloadSourceAuth, diag.Diagnostics) {
	if m.Auth != nil {
		return m.authRequest(ctx)
	}
	if prior.Auth != nil {
		return &downloadSourceAuth{}, nil
	}
	return nil, nil
}
