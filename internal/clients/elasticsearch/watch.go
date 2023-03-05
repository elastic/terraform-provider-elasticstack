package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutWatch(ctx context.Context, apiClient *clients.ApiClient, watch *models.Watch) diag.Diagnostics {
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
	active := esClient.Watcher.PutWatch.WithActive(watch.Status.State.Active)
	res, err := esClient.Watcher.PutWatch(watch.WatchID, active, body, esClient.Watcher.PutWatch.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update watch"); diags.HasError() {
		return diags
	}

	// if watch.Active {
	// 	_, err := esClient.Watcher.ActivateWatch(watch.WatchID)
	// 	if err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// } else {
	// 	_, err := esClient.Watcher.DeactivateWatch(watch.WatchID)
	// 	if err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// }

	return diags
}

func GetWatch(ctx context.Context, apiClient *clients.ApiClient, watchID string) (*models.Watch, diag.Diagnostics) {
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
	if diags := utils.CheckError(res, "Unable to find watch on cluster."); diags.HasError() {
		return nil, diags
	}

	var watch models.Watch
	if err := json.NewDecoder(res.Body).Decode(&watch); err != nil {
		return nil, diag.FromErr(err)
	}

	watch.WatchID = watchID
	return &watch, diags
}

func DeleteWatch(ctx context.Context, apiClient *clients.ApiClient, watchID string) diag.Diagnostics {
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
	if diags := utils.CheckError(res, "Unable to delete watch"); diags.HasError() {
		return diags
	}
	return diags
}
