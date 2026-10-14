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
	SpaceIDs         types.Set    `tfsdk:"space_ids"`         // > string
	KibanaConnection types.List   `tfsdk:"kibana_connection"` // > provider schema connection block
}

var agentDownloadSourceMinVersion = version.Must(version.NewVersion("8.13.0"))

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
	return reqs, nil
}

func (m model) toAPICreateModel(ctx context.Context) (kbapi.PostFleetAgentDownloadSourcesJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{
		Host:      m.Host.ValueString(),
		Name:      m.Name.ValueString(),
		IsDefault: m.Default.ValueBoolPointer(),
		ProxyId:   m.ProxyID.ValueStringPointer(),
	}

	// The API allows setting a custom id only at creation time.
	if typeutils.IsKnown(m.SourceID) && m.SourceID.ValueString() != "" {
		id := m.SourceID.ValueString()
		body.Id = &id
	}

	ssl, sslDiags := objectValueToSSL(ctx, m.Ssl)
	if sslDiags.HasError() {
		return kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{}, sslDiags
	}
	body.Ssl = sslToKibana(ssl)

	return body, nil
}

func (m model) toAPIUpdateModel(ctx context.Context, prior model) (kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{
		Host:      m.Host.ValueString(),
		Name:      m.Name.ValueString(),
		IsDefault: m.Default.ValueBoolPointer(),
		ProxyId:   fleet.ProxyIDForUpdate(m.ProxyID, prior.ProxyID),
	}

	ssl, sslDiags := objectValueToSSLUpdate(ctx, m.Ssl, prior.Ssl)
	if sslDiags.HasError() {
		return kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{}, sslDiags
	}
	body.Ssl = sslToKibana(ssl)

	return body, nil
}
