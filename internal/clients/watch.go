package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func (a *ApiClient) PutWatch(ctx context.Context, watch *models.Watch) diag.Diagnostics {
	var diags diag.Diagnostics
	watchBytes, err := json.Marshal(watch.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	body := a.es.Watcher.PutWatch.WithBody(bytes.NewReader(watchBytes))
	active := a.es.Watcher.PutWatch.WithActive(watch.Active)

	res, err := a.es.Watcher.PutWatch(watch.WatchID, body, active, a.es.Watcher.PutWatch.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update watch"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetWatch(ctx context.Context, watchID string) (*models.Watch, diag.Diagnostics) {
	var diags diag.Diagnostics
	res, err := a.es.Watcher.GetWatch(watchID, a.es.Watcher.GetWatch.WithContext(ctx))
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

	watch := make(map[string]models.Watch)
	if err := json.NewDecoder(res.Body).Decode(&watch); err != nil {
		return nil, diag.FromErr(err)
	}

	if watch, ok := watch[watchID]; ok {
		watch.WatchID = watchID
		return &watch, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find watch in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" watch in the cluster`, watchID),
	})
	return nil, diags
}

func (a *ApiClient) DeleteWatch(ctx context.Context, watchID string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Watcher.DeleteWatch(watchID, a.es.Watcher.DeleteWatch.WithContext(ctx))

	if err != nil && res.IsError() {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete watch"); diags.HasError() {
		return diags
	}
	return diags
}
