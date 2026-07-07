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

package elasticsearch

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/security/createapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/security/createcrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/security/invalidateapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/security/updateapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/security/updatecrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func CreateAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, req *createapikey.Request) (*createapikey.Response, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.CreateApiKey().Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to create apikey", err.Error())
		return nil, diags
	}

	return res, diags
}

func UpdateAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, req *updateapikey.Request) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.UpdateApiKey(id).Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to update apikey", err.Error())
		return diags
	}

	return diags
}

func GetAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) (*types.ApiKey, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.GetApiKey().Id(id).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Unable to get an apikey", err.Error())
		return nil, diags
	}

	if len(res.ApiKeys) != 1 {
		diags.AddError(
			"Unable to find an apikey in the cluster",
			fmt.Sprintf(`Unable to find "%s" apikey in the cluster`, id),
		)
		return nil, diags
	}

	apiKey := res.ApiKeys[0]
	return &apiKey, diags
}

func DeleteAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.InvalidateApiKey().Request(&invalidateapikey.Request{
		Ids: []string{id},
	}).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete an apikey", err.Error())
		return diags
	}

	return diags
}

func CreateCrossClusterAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, req *createcrossclusterapikey.Request) (*createcrossclusterapikey.Response, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.CreateCrossClusterApiKey().Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to create cross cluster apikey", err.Error())
		return nil, diags
	}

	return res, diags
}

func UpdateCrossClusterAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, req *updatecrossclusterapikey.Request) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.UpdateCrossClusterApiKey(id).Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to update cross cluster apikey", err.Error())
		return diags
	}

	return diags
}
