package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func (a *ApiClient) PutElasticsearchSnapshotRepository(repository *models.SnapshotRepository) diag.Diagnostics {
	var diags diag.Diagnostics
	snapRepoBytes, err := json.Marshal(repository)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending snapshot repository definition to ES API: %s", snapRepoBytes)
	res, err := a.es.Snapshot.CreateRepository(repository.Name, bytes.NewReader(snapRepoBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the snapshot repository"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchSnapshotRepository(name string) (*models.SnapshotRepository, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Snapshot.GetRepository.WithRepository(name)
	res, err := a.es.Snapshot.GetRepository(req)
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
	log.Printf("[TRACE] response ES API snapshot repository: %+v", snapRepoResponse)

	if currentRepo, ok := snapRepoResponse[name]; ok {
		return &currentRepo, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find requested repository",
		Detail:   fmt.Sprintf(`Repository "%s" is missing in the ES API response`, name),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchSnapshotRepository(name string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Snapshot.DeleteRepository([]string{name})
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete snapshot repository: %s", name)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchSlm(slm *models.SnapshotPolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	slmBytes, err := json.Marshal(slm)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending SLM to ES API: %s", slmBytes)
	req := a.es.SlmPutLifecycle.WithBody(bytes.NewReader(slmBytes))
	res, err := a.es.SlmPutLifecycle(slm.Id, req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the SLM"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchSlm(slmName string) (*models.SnapshotPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.SlmGetLifecycle.WithPolicyID(slmName)
	res, err := a.es.SlmGetLifecycle(req)
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
	type SlmReponse = map[string]struct {
		Policy models.SnapshotPolicy `json:"policy"`
	}
	var slmResponse SlmReponse
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

func (a *ApiClient) DeleteElasticsearchSlm(slmName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.SlmDeleteLifecycle(slmName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete SLM policy: %s", slmName)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) PutElasticsearchSettings(settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		diag.FromErr(err)
	}
	log.Printf("[TRACE] settings to set: %s", settingsBytes)
	res, err := a.es.Cluster.PutSettings(bytes.NewReader(settingsBytes))
	if err != nil {
		diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update cluster settings."); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchSettings() (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Cluster.GetSettings.WithFlatSettings(true)
	res, err := a.es.Cluster.GetSettings(req)
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
