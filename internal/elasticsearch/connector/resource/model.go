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

package resource

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// ContentConnectorData is the Terraform state model for the content connector resource.
type ContentConnectorData struct {
	entitycore.ElasticsearchConnectionField
	ID                  fwtypes.String `tfsdk:"id"`
	ConnectorID         fwtypes.String `tfsdk:"connector_id"`
	ServiceType         fwtypes.String `tfsdk:"service_type"`
	Name                fwtypes.String `tfsdk:"name"`
	Description         fwtypes.String `tfsdk:"description"`
	IndexName           fwtypes.String `tfsdk:"index_name"`
	IsNative            fwtypes.Bool   `tfsdk:"is_native"`
	Language            fwtypes.String `tfsdk:"language"`
	APIKeyID            fwtypes.String `tfsdk:"api_key_id"`
	APIKeySecretID      fwtypes.String `tfsdk:"api_key_secret_id"`
	Pipeline            fwtypes.Object `tfsdk:"pipeline"`
	Scheduling          fwtypes.Object `tfsdk:"scheduling"`
	Features            fwtypes.Object `tfsdk:"features"`
	ConfigurationValues fwtypes.Map    `tfsdk:"configuration_values"`
}

func (data ContentConnectorData) GetID() fwtypes.String         { return data.ID }
func (data ContentConnectorData) GetResourceID() fwtypes.String { return data.ConnectorID }
func (data ContentConnectorData) GetElasticsearchConnection() fwtypes.List {
	return data.ElasticsearchConnection
}

var (
	_ entitycore.ElasticsearchResourceModel = ContentConnectorData{}
	_ entitycore.WithVersionRequirements    = ContentConnectorData{}
	_ entitycore.WithOptionalWriteIdentity  = ContentConnectorData{}
	_ entitycore.WithReadResourceID         = ContentConnectorData{}
)

// AllowsEmptyWriteIdentityOnCreate satisfies [entitycore.WithOptionalWriteIdentity].
func (ContentConnectorData) AllowsEmptyWriteIdentityOnCreate() bool { return true }

// GetReadResourceID satisfies [entitycore.WithReadResourceID].
func (data ContentConnectorData) GetReadResourceID() string {
	if typeutils.IsKnown(data.ConnectorID) {
		return data.ConnectorID.ValueString()
	}
	return ""
}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (data ContentConnectorData) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *connector.MinSupportedVersion,
		ErrorMessage: "elasticstack_elasticsearch_connector requires Elasticsearch 8.16.0 or later (the connector request bodies the typed client sends are rejected on 8.12.x–8.15.x).",
	}}, nil
}
