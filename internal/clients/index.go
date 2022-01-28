package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func (a *ApiClient) PutElasticsearchIlm(policy *models.Policy) diag.Diagnostics {
	var diags diag.Diagnostics
	policyBytes, err := json.Marshal(map[string]interface{}{"policy": policy})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending new ILM policy to ES API: %s", policyBytes)
	req := a.es.ILM.PutLifecycle.WithBody(bytes.NewReader(policyBytes))
	res, err := a.es.ILM.PutLifecycle(policy.Name, req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the ILM policy"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchIlm(policyName string) (*models.PolicyDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.ILM.GetLifecycle.WithPolicy(policyName)
	res, err := a.es.ILM.GetLifecycle(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, "Unable to fetch ILM policy from the cluster."); diags.HasError() {
		return nil, diags
	}

	// our API response
	ilm := make(map[string]models.PolicyDefinition)
	if err := json.NewDecoder(res.Body).Decode(&ilm); err != nil {
		return nil, diag.FromErr(err)
	}

	if ilm, ok := ilm[policyName]; ok {
		return &ilm, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find a ILM policy in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" ILM policy in the cluster`, policyName),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchIlm(policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.ILM.DeleteLifecycle(policyName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete ILM policy."); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchIndexTemplate(template *models.IndexTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending request to ES: %s to create template '%s' ", templateBytes, template.Name)

	res, err := a.es.Indices.PutIndexTemplate(template.Name, bytes.NewReader(templateBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create index template"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchIndexTemplate(templateName string) (*models.IndexTemplateResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Indices.GetIndexTemplate.WithName(templateName)
	res, err := a.es.Indices.GetIndexTemplate(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, "Unable to request index template."); diags.HasError() {
		return nil, diags
	}

	var indexTemplates models.IndexTemplatesResponse
	if err := json.NewDecoder(res.Body).Decode(&indexTemplates); err != nil {
		return nil, diag.FromErr(err)
	}

	// we requested only 1 template
	if len(indexTemplates.IndexTemplates) != 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong number of templates returned",
			Detail:   fmt.Sprintf("Elasticsearch API returned %d when requsted '%s' template.", len(indexTemplates.IndexTemplates), templateName),
		})
		return nil, diags
	}
	tpl := indexTemplates.IndexTemplates[0]
	log.Printf("[TRACE] read index template from API: %+v", tpl)
	return &tpl, diags
}

func (a *ApiClient) DeleteElasticsearchIndexTemplate(templateName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Indices.DeleteIndexTemplate(templateName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete index template"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchIndex(index *models.Index) diag.Diagnostics {
	var diags diag.Diagnostics
	indexBytes, err := json.Marshal(index)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] index definition: %s", indexBytes)

	req := a.es.Indices.Create.WithBody(bytes.NewReader(indexBytes))
	res, err := a.es.Indices.Create(index.Name, req)
	if err != nil {
		diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create index: %s", index.Name)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) DeleteElasticsearchIndex(name string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.Indices.Delete([]string{name})
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete the index: %s", name)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchIndex(name string) (*models.Index, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := a.es.Indices.Get.WithFlatSettings(true)
	res, err := a.es.Indices.Get([]string{name}, req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	// if there is no index found, return the empty struct, which should force the creation of the index
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get requested index: %s", name)); diags.HasError() {
		return nil, diags
	}

	indices := make(map[string]models.Index)
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, diag.FromErr(err)
	}
	index := indices[name]
	return &index, diags
}

func (a *ApiClient) DeleteElasticsearchIndexAlias(index string, aliases []string) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[TRACE] Deleting aliases for index %s: %v", index, aliases)
	res, err := a.es.Indices.DeleteAlias([]string{index}, aliases)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete aliases '%v' for index '%s'", index, aliases)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) UpdateElasticsearchIndexAlias(index string, alias *models.IndexAlias) diag.Diagnostics {
	var diags diag.Diagnostics
	aliasBytes, err := json.Marshal(alias)
	if err != nil {
		diag.FromErr(err)
	}
	log.Printf("[TRACE] updaing index %s alias: %s", index, aliasBytes)
	req := a.es.Indices.PutAlias.WithBody(bytes.NewReader(aliasBytes))
	res, err := a.es.Indices.PutAlias([]string{index}, alias.Name, req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to update alias '%v' for index '%s'", index, alias.Name)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) UpdateElasticsearchIndexSettings(index string, settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		diag.FromErr(err)
	}
	log.Printf("[TRACE] updaing index %s settings: %s", index, settingsBytes)
	req := a.es.Indices.PutSettings.WithIndex(index)
	res, err := a.es.Indices.PutSettings(bytes.NewReader(settingsBytes), req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update index settings"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) UpdateElasticsearchIndexMappings(index, mappings string) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[TRACE] updaing index %s mappings: %s", index, mappings)
	req := a.es.Indices.PutMapping.WithIndex(index)
	res, err := a.es.Indices.PutMapping(strings.NewReader(mappings), req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update index mappings"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchDataStream(dataStreamName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.Indices.CreateDataStream(dataStreamName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create DataStream: %s", dataStreamName)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchDataStream(dataStreamName string) (*models.DataStream, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Indices.GetDataStream.WithName(dataStreamName)
	res, err := a.es.Indices.GetDataStream(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get requested DataStream: %s", dataStreamName)); diags.HasError() {
		return nil, diags
	}

	dStreams := make(map[string][]models.DataStream)
	if err := json.NewDecoder(res.Body).Decode(&dStreams); err != nil {
		return nil, diag.FromErr(err)
	}
	log.Printf("[TRACE] get data stream '%v' from ES api: %+v", dataStreamName, dStreams)
	// if the DataStream found in must be the first index in the data_stream object
	ds := dStreams["data_streams"][0]
	return &ds, diags
}

func (a *ApiClient) DeleteElasticsearchDataStream(dataStreamName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.Indices.DeleteDataStream([]string{dataStreamName})
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete DataStream: %s", dataStreamName)); diags.HasError() {
		return diags
	}

	return diags
}
