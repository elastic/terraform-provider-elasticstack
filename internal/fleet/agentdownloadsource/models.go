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
	Ssl              types.Object `tfsdk:"ssl"`
	Auth             *authModel   `tfsdk:"auth"`
	SpaceIDs         types.Set    `tfsdk:"space_ids"`         // > string
	KibanaConnection types.List   `tfsdk:"kibana_connection"` // > provider schema connection block
}

type authModel struct {
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PasswordWO types.String `tfsdk:"password_wo"`
	APIKey     types.String `tfsdk:"api_key"`
	APIKeyWO   types.String `tfsdk:"api_key_wo"`
	Headers    types.List   `tfsdk:"headers"`
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

// agentDownloadSourceSslMinVersion gates the ssl attribute on its own so
// configurations without ssl keep the 8.13.0 resource floor.
var agentDownloadSourceSslMinVersion = version.Must(version.NewVersion("9.1.0"))

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
	if typeutils.IsKnown(m.Ssl) {
		reqs = append(reqs, entitycore.NewAttributeVersionRequirement(
			path.Root("ssl"),
			*agentDownloadSourceSslMinVersion,
			fmt.Sprintf("ssl is only supported in Elastic Stack %s and above", agentDownloadSourceSslMinVersion),
		))
	}
	if m.Auth != nil {
		reqs = append(reqs, entitycore.NewAttributeVersionRequirement(
			path.Root("auth"),
			*agentDownloadSourceAuthMinVersion,
			fmt.Sprintf("`auth` is not supported on this version; it is only supported in Elastic Stack %s and above", agentDownloadSourceAuthMinVersion),
		))
	}

	return reqs, nil
}

// toAPICreateModel builds the create body. Create has no existing secrets to
// reference, so every secret is sent as a literal.
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

	if diags.HasError() {
		return kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{}, diags
	}

	ssl, sslDiags := objectValueToSSL(ctx, m.Ssl)
	diags.Append(sslDiags...)
	if diags.HasError() {
		return kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{}, diags
	}
	body.Ssl = sslToKibana(ssl)

	secrets, secretDiags := m.secretsPayload(ctx, literalSecrets)
	diags.Append(secretDiags...)
	if err := setSecrets(&body.Secrets, secrets); err != nil {
		diags.AddError("Failed to build secrets request", err.Error())
	}
	if diags.HasError() {
		return kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{}, diags
	}

	return body, diags
}

// withWriteOnlyFrom returns a copy of m whose write-only values are taken from
// config. The framework represents write-only values as null in the plan, so
// the write path must read them from configuration.
func (m model) withWriteOnlyFrom(ctx context.Context, config model) (model, diag.Diagnostics) {
	var diags diag.Diagnostics

	if typeutils.IsKnown(config.Ssl) {
		configSSL := typeutils.ObjectTypeAs[sslModel](ctx, config.Ssl, path.Root("ssl"), &diags)
		if diags.HasError() {
			return m, diags
		}
		ssl, sslDiags := withSSLWriteOnlyKey(ctx, m.Ssl, configSSL.KeyWO)
		diags.Append(sslDiags...)
		m.Ssl = ssl
	}

	if m.Auth != nil && config.Auth != nil {
		auth := *m.Auth
		auth.PasswordWO = config.Auth.PasswordWO
		auth.APIKeyWO = config.Auth.APIKeyWO
		m.Auth = &auth
	}

	return m, diags
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
	var diags diag.Diagnostics
	// SliceRef keeps a known empty list as "headers":[]. SliceNilIfEmpty would omit it.
	headers := typeutils.SliceRef(typeutils.ListTypeToSlice(ctx, m.Headers, path.Root("auth").AtName("headers"), &diags,
		func(header authHeaderModel, _ typeutils.ListMeta) downloadSourceAuthHeader {
			return downloadSourceAuthHeader{Key: header.Key.ValueString(), Value: header.Value.ValueString()}
		}))
	return headers, diags
}

func (m model) toAPIUpdateModel(ctx context.Context, prior model, resolve secretResolver) (kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody, diag.Diagnostics) {
	auth, diags := m.authForUpdate(ctx, prior)
	if diags.HasError() {
		return kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{}, diags
	}
	body := kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{
		Host:      m.Host.ValueString(),
		Name:      m.Name.ValueString(),
		IsDefault: m.Default.ValueBoolPointer(),
		ProxyId:   fleet.ProxyIDForUpdate(m.ProxyID, prior.ProxyID),
		Auth:      auth,
	}

	ssl, sslDiags := objectValueToSSLUpdate(ctx, m.Ssl, prior.Ssl)
	diags.Append(sslDiags...)
	if diags.HasError() {
		return kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{}, diags
	}
	body.Ssl = sslToKibana(ssl)

	secrets, secretDiags := m.secretsPayload(ctx, resolve)
	diags.Append(secretDiags...)
	if err := setSecrets(&body.Secrets, secrets); err != nil {
		diags.AddError("Failed to build secrets request", err.Error())
	}
	if diags.HasError() {
		return kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{}, diags
	}

	return body, diags
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
