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

package security

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID       types.String         `tfsdk:"id"`
	Username types.String         `tfsdk:"username"`
	FullName types.String         `tfsdk:"full_name"`
	Email    types.String         `tfsdk:"email"`
	Roles    types.Set            `tfsdk:"roles"`
	Metadata jsontypes.Normalized `tfsdk:"metadata"`
	Enabled  types.Bool           `tfsdk:"enabled"`
}

func NewUserDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[userDataSourceModel](
		entitycore.ComponentElasticsearch,
		"security_user",
		getUserDataSourceSchema,
		readUserDataSource,
	)
}

func getUserDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: userDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "An identifier for the user",
				Required:            true,
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "The full name of the user.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the user.",
				Computed:            true,
			},
			"roles": schema.SetAttribute{
				MarkdownDescription: "A set of roles the user has. The roles determine the user's access permissions. Default is [].",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Arbitrary metadata that you want to associate with the user.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the user is enabled. The default value is true.",
				Computed:            true,
			},
		},
	}
}

func readUserDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config userDataSourceModel) (userDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	username := config.Username.ValueString()

	id, sdkDiags := esClient.ID(ctx, username)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}
	config.ID = types.StringValue(id.String())

	user, sdkDiags := elasticsearch.GetUser(ctx, esClient, username)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}

	if user == nil {
		config.ID = types.StringValue("")
		config.FullName = types.StringNull()
		config.Email = types.StringNull()
		config.Roles = types.SetNull(types.StringType)
		config.Metadata = jsontypes.NewNormalizedNull()
		config.Enabled = types.BoolNull()
		config.Username = types.StringValue(username)
		return config, diags
	}

	if user.Email != nil {
		config.Email = types.StringValue(*user.Email)
	} else {
		config.Email = types.StringValue("")
	}
	if user.FullName != nil {
		config.FullName = types.StringValue(*user.FullName)
	} else {
		config.FullName = types.StringValue("")
	}

	rolesSet, d := types.SetValueFrom(ctx, types.StringType, user.Roles)
	diags.Append(d...)
	if diags.HasError() {
		return config, diags
	}
	config.Roles = rolesSet

	metadata, err := json.Marshal(user.Metadata)
	if err != nil {
		diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata JSON: %s", err))
		return config, diags
	}
	config.Metadata = jsontypes.NewNormalizedValue(string(metadata))

	config.Enabled = types.BoolValue(user.Enabled)
	config.Username = types.StringValue(username)

	return config, diags
}
