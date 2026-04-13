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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type model struct {
	ID       types.String `tfsdk:"id"`
	SourceID types.String `tfsdk:"source_id"`
	Name     types.String `tfsdk:"name"`
	Host     types.String `tfsdk:"host"`
	Default  types.Bool   `tfsdk:"default"`
	ProxyID  types.String `tfsdk:"proxy_id"`
	SpaceIDs types.Set    `tfsdk:"space_ids"` // > string
}

func (m model) toAPICreateModel(_ context.Context) kbapi.PostFleetAgentDownloadSourcesJSONRequestBody {
	body := kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{
		Host:      m.Host.ValueString(),
		Name:      m.Name.ValueString(),
		IsDefault: m.Default.ValueBoolPointer(),
		ProxyId:   m.ProxyID.ValueStringPointer(),
	}

	// The API allows setting a custom id only at creation time.
	if !m.SourceID.IsNull() && !m.SourceID.IsUnknown() && m.SourceID.ValueString() != "" {
		id := m.SourceID.ValueString()
		body.Id = &id
	}

	return body
}

func (m model) toAPIUpdateModel(_ context.Context) kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody {
	return kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{
		Host:      m.Host.ValueString(),
		Name:      m.Name.ValueString(),
		IsDefault: m.Default.ValueBoolPointer(),
		ProxyId:   m.ProxyID.ValueStringPointer(),
	}
}
