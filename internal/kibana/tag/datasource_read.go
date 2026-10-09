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
	_ string,
	spaceID string,
	config tagsDataSourceModel,
) (tagsDataSourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	spaceID = clients.EffectiveSpaceID(spaceID)
	config.SpaceID = types.StringValue(spaceID)

	oapiClient := client.GetKibanaOapiClient()
	query := typeutils.ValueStringPointer(config.Query)
	tags, listDiags := listAllTags(ctx, oapiClient, spaceID, query, kibanaoapi.ListTags)
	diags.Append(listDiags...)
	if diags.HasError() {
		return config, false, diags
	}

	diags.Append(config.setTags(ctx, tags)...)
	return config, !diags.HasError(), diags
}

type listTagsPageFunc func(context.Context, *kibanaoapi.Client, string, *kbapi.GetTagsParams) (*kibanaoapi.TagListResult, diag.Diagnostics)

func listAllTags(ctx context.Context, client *kibanaoapi.Client, spaceID string, query *string, listPage listTagsPageFunc) ([]kibanaoapi.TagDetail, diag.Diagnostics) {
	fetchPage := func(page float32) ([]kibanaoapi.TagDetail, float32, diag.Diagnostics) {
		perPage := kibanaoapi.TagListMaxPerPage()
		params := &kbapi.GetTagsParams{
			Page:    &page,
			PerPage: &perPage,
		}
		if query != nil && *query != "" {
			params.Query = query
		}

		result, diags := listPage(ctx, client, spaceID, params)
		if diags.HasError() {
			return nil, 0, diags
		}

		return result.Tags, result.Total, nil
	}

	return kibanaoapi.CollectAllPages(fetchPage)
}
