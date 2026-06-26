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

package tag

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readTagsDataSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	config tagsDataSourceModel,
) (tagsDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	spaceID := clients.DefaultSpaceID
	if typeutils.IsKnown(config.SpaceID) && config.SpaceID.ValueString() != "" {
		spaceID = config.SpaceID.ValueString()
	}
	config.SpaceID = types.StringValue(spaceID)

	oapiClient := client.GetKibanaOapiClient()
	query := typeutils.ValueStringPointer(config.Query)
	tags, listDiags := listAllTags(ctx, oapiClient, spaceID, query)
	diags.Append(listDiags...)
	if diags.HasError() {
		return config, diags
	}

	diags.Append(config.setTags(ctx, tags)...)
	return config, diags
}

func listAllTags(ctx context.Context, client *kibanaoapi.Client, spaceID string, query *string) ([]kibanaoapi.TagDetail, diag.Diagnostics) {
	var (
		collected []kibanaoapi.TagDetail
		page      float32 = 1
		total     float32
	)

	for {
		perPage := kibanaoapi.TagListMaxPerPage()
		params := &kbapi.GetTagsParams{
			Page:    &page,
			PerPage: &perPage,
		}
		if query != nil && *query != "" {
			params.Query = query
		}

		result, diags := kibanaoapi.ListTags(ctx, client, spaceID, params)
		if diags.HasError() {
			return nil, diags
		}

		collected = append(collected, result.Tags...)
		if total == 0 {
			total = result.Total
		}

		if len(collected) >= int(total) || len(result.Tags) == 0 {
			break
		}
		page++
	}

	return collected, nil
}
