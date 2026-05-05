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
	return PutWatchBodyJSON(ctx, apiClient, watch.WatchID, watch.Active, watchBodyBytes)
}

// PutWatchBodyJSON sends a pre-encoded watch document (the JSON object under the watch id) to Put Watch.
//
// We use .Raw() because the provider stores watch body fields (trigger, input,
// condition, actions, metadata, transform) as normalized JSON strings in the
// Terraform state, then unmarshals them into map[string]any for transport. The
// typed types.Watch uses strongly-typed unboxed structs that do not align with
// this map[string]any shape, so passing through .Raw() preserves the exact
// JSON produced by the resource layer while still using the typed client for
// transport.
func PutWatchBodyJSON(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, watchID string, active bool, watchBodyJSON []byte) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Watcher.PutWatch(watchID).Active(active).Raw(bytes.NewReader(watchBodyJSON)).Do(ctx)
	if err != nil {
		diags.AddError("Unable to put watch '"+watchID+"'", err.Error())
		return diags
	}
	return diags
}

func GetWatch(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, watchID string) (*models.Watch, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	// We use .Perform() (raw *http.Response) instead of .Do() which would
	// return *getwatch.Response containing *types.Watch. The typed
	// types.Watch has trigger/input/condition/actions as strongly-typed
	// unboxed structs, while the provider's models.Watch stores them as
	// map[string]any under Body (with json:"watch"). The JSON shapes do
	// not align for a simple Marshal/Unmarshal round-trip, so we decode
	// directly from the raw response body into models.Watch.
	res, err := typedClient.Watcher.GetWatch(watchID).Perform(ctx)
	if err != nil {
		diags.AddError("Unable to get watch", err.Error())
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if d := diagutil.CheckHTTPErrorFromFW(res, "Unable to get watch from cluster."); d.HasError() {
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

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Watcher.DeleteWatch(watchID).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return diags // already gone, treat as success
		}
		diags.AddError("Unable to delete watch", err.Error())
		return diags
	}
	return diags
}
