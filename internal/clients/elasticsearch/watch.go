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
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutWatch(ctx context.Context, apiClient *clients.APIClient, watch *models.PutWatch) diag.Diagnostics {
	var diags diag.Diagnostics
	watchBodyBytes, err := json.Marshal(watch.Body)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	body := esClient.Watcher.PutWatch.WithBody(bytes.NewReader(watchBodyBytes))
	active := esClient.Watcher.PutWatch.WithActive(watch.Active)
	res, err := esClient.Watcher.PutWatch(watch.WatchID, active, body, esClient.Watcher.PutWatch.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to create or update watch"); diags.HasError() {
		return diags
	}

	return diags
}

func GetWatch(ctx context.Context, apiClient *clients.APIClient, watchID string) (*models.Watch, diag.Diagnostics) {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := esClient.Watcher.GetWatch(watchID, esClient.Watcher.GetWatch.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := diagutil.CheckError(res, "Unable to find watch on cluster."); diags.HasError() {
		return nil, diags
	}

	var watch models.Watch
	if err := json.NewDecoder(res.Body).Decode(&watch); err != nil {
		return nil, diag.FromErr(err)
	}

	watch.WatchID = watchID
	return &watch, diags
}

func DeleteWatch(ctx context.Context, apiClient *clients.APIClient, watchID string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Watcher.DeleteWatch(watchID, esClient.Watcher.DeleteWatch.WithContext(ctx))

	if err != nil && res.IsError() {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to delete watch"); diags.HasError() {
		return diags
	}
	return diags
}
