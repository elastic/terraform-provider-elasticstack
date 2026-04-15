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
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutWatch(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, watch *models.PutWatch) fwdiag.Diagnostics {
	watchBodyBytes, err := json.Marshal(watch.Body)
	if err != nil {
		var diags fwdiag.Diagnostics
		diags.AddError("Unable to marshal watch body", err.Error())
		return diags
	}
	return putWatchBytes(ctx, apiClient, watch.WatchID, watch.Active, watchBodyBytes)
}

// PutWatchBodyJSON sends a pre-encoded watch document (the JSON object under the watch id) to Put Watch.
func PutWatchBodyJSON(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, watchID string, active bool, watchBodyJSON []byte) fwdiag.Diagnostics {
	return putWatchBytes(ctx, apiClient, watchID, active, watchBodyJSON)
}

func putWatchBytes(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, watchID string, active bool, watchBodyBytes []byte) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	body := esClient.Watcher.PutWatch.WithBody(bytes.NewReader(watchBodyBytes))
	putActive := esClient.Watcher.PutWatch.WithActive(active)
	res, err := esClient.Watcher.PutWatch(watchID, putActive, body, esClient.Watcher.PutWatch.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to create or update watch", err.Error())
		return diags
	}
	defer res.Body.Close()

	return diagutil.CheckErrorFromFW(res, "Unable to create or update watch")
}

func GetWatch(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, watchID string) (*models.Watch, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := esClient.Watcher.GetWatch(watchID, esClient.Watcher.GetWatch.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to get watch", err.Error())
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if d := diagutil.CheckErrorFromFW(res, "Unable to find watch on cluster."); d.HasError() {
		return nil, d
	}

	var watch models.Watch
	if err := json.NewDecoder(res.Body).Decode(&watch); err != nil {
		diags.AddError("Unable to decode watch response", err.Error())
		return nil, diags
	}

	watch.WatchID = watchID
	return &watch, diags
}

func DeleteWatch(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, watchID string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	res, err := esClient.Watcher.DeleteWatch(watchID, esClient.Watcher.DeleteWatch.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to delete watch", err.Error())
		return diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return diags // already gone, treat as success
	}

	return diagutil.CheckErrorFromFW(res, "Unable to delete watch")
}
