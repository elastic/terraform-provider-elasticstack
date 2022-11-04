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

func (a *ApiClient) PutElasticsearchSnapshotRepository(ctx context.Context, repository *models.SnapshotRepository) diag.Diagnostics {
	var diags diag.Diagnostics
	snapRepoBytes, err := json.Marshal(repository)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := a.es.Snapshot.CreateRepository(repository.Name, bytes.NewReader(snapRepoBytes), a.es.Snapshot.CreateRepository.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the snapshot repository"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchSnapshotRepository(ctx context.Context, name string) (*models.SnapshotRepository, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Snapshot.GetRepository.WithRepository(name)
	res, err := a.es.Snapshot.GetRepository(req, a.es.Snapshot.GetRepository.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get the information about snapshot repository: %s", name)); diags.HasError() {
		return nil, diags
	}
	snapRepoResponse := make(map[string]models.SnapshotRepository)
	if err := json.NewDecoder(res.Body).Decode(&snapRepoResponse); err != nil {
		return nil, diag.FromErr(err)
	}

	if currentRepo, ok := snapRepoResponse[name]; ok {
		if len(currentRepo.Name) <= 0 {
			currentRepo.Name = name
		}
		return &currentRepo, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find requested repository",
		Detail:   fmt.Sprintf(`Repository "%s" is missing in the ES API response`, name),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchSnapshotRepository(ctx context.Context, name string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Snapshot.DeleteRepository([]string{name}, a.es.Snapshot.DeleteRepository.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete snapshot repository: %s", name)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchSlm(ctx context.Context, slm *models.SnapshotPolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	slmBytes, err := json.Marshal(slm)
	if err != nil {
		return diag.FromErr(err)
	}
	req := a.es.SlmPutLifecycle.WithBody(bytes.NewReader(slmBytes))
	res, err := a.es.SlmPutLifecycle(slm.Id, req, a.es.SlmPutLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the SLM"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchSlm(ctx context.Context, slmName string) (*models.SnapshotPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.SlmGetLifecycle.WithPolicyID(slmName)
	res, err := a.es.SlmGetLifecycle(req, a.es.SlmGetLifecycle.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, "Unable to get SLM policy from ES API"); diags.HasError() {
		return nil, diags
	}
	type SlmResponse = map[string]struct {
		Policy models.SnapshotPolicy `json:"policy"`
	}
	var slmResponse SlmResponse
	if err := json.NewDecoder(res.Body).Decode(&slmResponse); err != nil {
		return nil, diag.FromErr(err)
	}
	if slm, ok := slmResponse[slmName]; ok {
		return &slm.Policy, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find the SLM policy in the response",
		Detail:   fmt.Sprintf(`Unable to find "%s" policy in the ES API response.`, slmName),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchSlm(ctx context.Context, slmName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.SlmDeleteLifecycle(slmName, a.es.SlmDeleteLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete SLM policy: %s", slmName)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) PutElasticsearchSettings(ctx context.Context, settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		diag.FromErr(err)
	}
	res, err := a.es.Cluster.PutSettings(bytes.NewReader(settingsBytes), a.es.Cluster.PutSettings.WithContext(ctx))
	if err != nil {
		diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update cluster settings."); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchSettings(ctx context.Context) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Cluster.GetSettings.WithFlatSettings(true)
	res, err := a.es.Cluster.GetSettings(req, a.es.Cluster.GetSettings.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to read cluster settings."); diags.HasError() {
		return nil, diags
	}

	clusterSettings := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&clusterSettings); err != nil {
		return nil, diag.FromErr(err)
	}
	return clusterSettings, diags
}

func (a *ApiClient) GetElasticsearchScript(ctx context.Context, id string) (*models.Script, diag.Diagnostics) {
	res, err := a.es.GetScript(id, a.es.GetScript.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get stored script: %s", id)); diags.HasError() {
		return nil, diags
	}
	var scriptResponse struct {
		Script *models.Script `json:"script"`
	}
	if err := json.NewDecoder(res.Body).Decode(&scriptResponse); err != nil {
		return nil, diag.FromErr(err)
	}

	return scriptResponse.Script, nil
}

func (a *ApiClient) PutElasticsearchScript(ctx context.Context, script *models.Script) diag.Diagnostics {
	req := struct {
		Script *models.Script `json:"script"`
	}{
		script,
	}
	scriptBytes, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := a.es.PutScript(script.ID, bytes.NewReader(scriptBytes), a.es.PutScript.WithContext(ctx), a.es.PutScript.WithScriptContext(script.Context))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to put stored script"); diags.HasError() {
		return diags
	}
	return nil
}

func (a *ApiClient) DeleteElasticsearchScript(ctx context.Context, id string) diag.Diagnostics {
	res, err := a.es.DeleteScript(id, a.es.DeleteScript.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete script: %s", id)); diags.HasError() {
		return diags
	}
	return nil
}
