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

package cloudconnector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Returns Fleet cloud connectors visible in a Kibana space. " +
			"Cloud connectors are a preview feature in Kibana; this data source is experimental and may change in future provider releases. " +
			"See the [Fleet Cloud Connectors API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-fleet-cloud-connectors) for more information.",
		Attributes: map[string]dsschema.Attribute{
			attrID: dsschema.StringAttribute{
				MarkdownDescription: "Internal identifier for this data source read.",
				Computed:            true,
			},
			attrSpaceID: dsschema.StringAttribute{
				MarkdownDescription: "The Kibana space ID to scope the request to. When not specified, the default space is used.",
				Optional:            true,
			},
			attrKuery: dsschema.StringAttribute{
				MarkdownDescription: "Optional KQL filter passed to the Fleet API as the `kuery` query parameter.",
				Optional:            true,
			},
			attrPage: dsschema.Int64Attribute{
				MarkdownDescription: "Optional page number for API pagination, passed as the `page` query parameter.",
				Optional:            true,
			},
			attrPerPage: dsschema.Int64Attribute{
				MarkdownDescription: "Optional page size for API pagination, passed as the `perPage` query parameter.",
				Optional:            true,
			},
			attrCloudConnectors: dsschema.ListNestedAttribute{
				MarkdownDescription: "Cloud connectors returned by the Fleet list API. Secret configuration (`vars`) is omitted from each entry.",
				Computed:            true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						attrID: dsschema.StringAttribute{
							MarkdownDescription: "The composite ID of the cloud connector: `<space_id>/<cloud_connector_id>`.",
							Computed:            true,
						},
						attrCloudConnectorID: dsschema.StringAttribute{
							MarkdownDescription: "The API-assigned cloud connector ID.",
							Computed:            true,
						},
						attrSpaceID: dsschema.StringAttribute{
							MarkdownDescription: "The Kibana space ID where this cloud connector is available.",
							Computed:            true,
						},
						attrName: dsschema.StringAttribute{
							MarkdownDescription: "The cloud connector name.",
							Computed:            true,
						},
						attrCloudProvider: dsschema.StringAttribute{
							MarkdownDescription: "The cloud provider for this connector. One of `aws`, `azure`, or `gcp`.",
							Computed:            true,
						},
						attrAccountType: dsschema.StringAttribute{
							MarkdownDescription: "The account type: `single-account` or `organization-account`.",
							Computed:            true,
						},
						attrNamespace: dsschema.StringAttribute{
							MarkdownDescription: "The namespace assigned to this cloud connector.",
							Computed:            true,
						},
						attrPackagePolicyCount: dsschema.Int64Attribute{
							MarkdownDescription: "The number of package policies using this cloud connector.",
							Computed:            true,
						},
						attrVerificationStatus: dsschema.StringAttribute{
							MarkdownDescription: "The connector verification status. May be null on first read because verification is asynchronous.",
							Computed:            true,
						},
						attrVerificationStartedAt: dsschema.StringAttribute{
							MarkdownDescription: "When connector verification started. May be null on first read because verification is asynchronous.",
							Computed:            true,
						},
						attrVerificationFailedAt: dsschema.StringAttribute{
							MarkdownDescription: "When connector verification failed, if applicable. May be null on first read because verification is asynchronous.",
							Computed:            true,
						},
						attrCreatedAt: dsschema.StringAttribute{
							MarkdownDescription: "When the cloud connector was created, in ISO 8601 format.",
							Computed:            true,
						},
						attrUpdatedAt: dsschema.StringAttribute{
							MarkdownDescription: "When the cloud connector was last updated, in ISO 8601 format.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func getCloudConnectorListItemElemType(ctx context.Context) attr.Type {
	return getDataSourceSchema(ctx).Attributes[attrCloudConnectors].GetType().(attr.TypeWithElementType).ElementType()
}
