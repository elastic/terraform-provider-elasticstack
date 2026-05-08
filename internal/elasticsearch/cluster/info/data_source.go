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

package info

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewDataSource returns the Plugin Framework datasource.DataSource for
// elasticstack_elasticsearch_info.
func NewDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[dataSourceModel](
		entitycore.ComponentElasticsearch,
		"info",
		getDataSourceSchema,
		readDataSource,
	)
}

// versionAttrTypes returns the attr.Type map for a versionModel element,
// matching the tfsdk tags in versionModel.
func versionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"build_date":                          types.StringType,
		"build_flavor":                        types.StringType,
		"build_hash":                          types.StringType,
		"build_snapshot":                      types.BoolType,
		"build_type":                          types.StringType,
		"lucene_version":                      types.StringType,
		"minimum_index_compatibility_version": types.StringType,
		"minimum_wire_compatibility_version":  types.StringType,
		"number":                              types.StringType,
	}
}

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config dataSourceModel) (dataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	res, sdkDiags := elasticsearch.GetClusterInfo(ctx, esClient)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}

	config.ID = types.StringValue(res.ClusterUuid)
	config.ClusterUUID = types.StringValue(res.ClusterUuid)
	config.ClusterName = types.StringValue(res.ClusterName)
	config.Name = types.StringValue(res.Name)
	config.Tagline = types.StringValue(res.Tagline)

	// Build-date: the API type is DateTime (any) — Go's encoding/json decodes
	// JSON numbers into float64 when the target is any (not int64 as the old
	// SDK implementation incorrectly did), so we handle float64 here.
	var buildDate string
	switch v := res.Version.BuildDate.(type) {
	case string:
		buildDate = v
	case float64:
		// JSON numbers decode to float64; format as integer milliseconds string.
		buildDate = fmt.Sprintf("%d", int64(v))
	default:
		buildDate = fmt.Sprintf("%v", v)
	}

	verObj := versionModel{
		BuildDate:                        types.StringValue(buildDate),
		BuildFlavor:                      types.StringValue(res.Version.BuildFlavor),
		BuildHash:                        types.StringValue(res.Version.BuildHash),
		BuildSnapshot:                    types.BoolValue(res.Version.BuildSnapshot),
		BuildType:                        types.StringValue(res.Version.BuildType),
		LuceneVersion:                    types.StringValue(res.Version.LuceneVersion),
		MinimumIndexCompatibilityVersion: types.StringValue(res.Version.MinimumIndexCompatibilityVersion),
		MinimumWireCompatibilityVersion:  types.StringValue(res.Version.MinimumWireCompatibilityVersion),
		Number:                           types.StringValue(res.Version.Int),
	}

	versionList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: versionAttrTypes()}, []versionModel{verObj})
	diags.Append(listDiags...)
	if diags.HasError() {
		return config, diags
	}

	config.Version = versionList
	return config, diags
}
