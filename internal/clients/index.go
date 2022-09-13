package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func (a *ApiClient) PutElasticsearchIlm(ctx context.Context, policy *models.Policy) diag.Diagnostics {
	var diags diag.Diagnostics
	policyBytes, err := json.Marshal(map[string]interface{}{"policy": policy})
	if err != nil {
		return diag.FromErr(err)
	}
	req := a.es.ILM.PutLifecycle.WithBody(bytes.NewReader(policyBytes))
	res, err := a.es.ILM.PutLifecycle(policy.Name, req, a.es.ILM.PutLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the ILM policy"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchIlm(ctx context.Context, policyName string) (*models.PolicyDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.ILM.GetLifecycle.WithPolicy(policyName)
	res, err := a.es.ILM.GetLifecycle(req, a.es.ILM.GetLifecycle.WithContext(ctx))
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

func (a *ApiClient) DeleteElasticsearchIlm(ctx context.Context, policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.ILM.DeleteLifecycle(policyName, a.es.ILM.DeleteLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete ILM policy."); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchComponentTemplate(ctx context.Context, template *models.ComponentTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := a.es.Cluster.PutComponentTemplate(template.Name, bytes.NewReader(templateBytes), a.es.Cluster.PutComponentTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create component template"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchComponentTemplate(ctx context.Context, templateName string) (*models.ComponentTemplateResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Cluster.GetComponentTemplate.WithName(templateName)
	res, err := a.es.Cluster.GetComponentTemplate(req, a.es.Cluster.GetComponentTemplate.WithContext(ctx))
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

	var componentTemplates models.ComponentTemplatesResponse
	if err := json.NewDecoder(res.Body).Decode(&componentTemplates); err != nil {
		return nil, diag.FromErr(err)
	}

	// we requested only 1 template
	if len(componentTemplates.ComponentTemplates) != 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong number of templates returned",
			Detail:   fmt.Sprintf("Elasticsearch API returned %d when requested '%s' component template.", len(componentTemplates.ComponentTemplates), templateName),
		})
		return nil, diags
	}
	tpl := componentTemplates.ComponentTemplates[0]
	return &tpl, diags
}

func (a *ApiClient) DeleteElasticsearchComponentTemplate(ctx context.Context, templateName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Cluster.DeleteComponentTemplate(templateName, a.es.Cluster.DeleteComponentTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete component template"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchIndexTemplate(ctx context.Context, template *models.IndexTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := a.es.Indices.PutIndexTemplate(template.Name, bytes.NewReader(templateBytes), a.es.Indices.PutIndexTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create index template"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchIndexTemplate(ctx context.Context, templateName string) (*models.IndexTemplateResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Indices.GetIndexTemplate.WithName(templateName)
	res, err := a.es.Indices.GetIndexTemplate(req, a.es.Indices.GetIndexTemplate.WithContext(ctx))
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
	return &tpl, diags
}

func (a *ApiClient) DeleteElasticsearchIndexTemplate(ctx context.Context, templateName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Indices.DeleteIndexTemplate(templateName, a.es.Indices.DeleteIndexTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete index template"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchIndex(ctx context.Context, index *models.Index) diag.Diagnostics {
	var diags diag.Diagnostics
	indexBytes, err := json.Marshal(index)
	if err != nil {
		return diag.FromErr(err)
	}

	req := a.es.Indices.Create.WithBody(bytes.NewReader(indexBytes))
	res, err := a.es.Indices.Create(index.Name, req, a.es.Indices.Create.WithContext(ctx))
	if err != nil {
		diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create index: %s", index.Name)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) DeleteElasticsearchIndex(ctx context.Context, name string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.Indices.Delete([]string{name}, a.es.Indices.Delete.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete the index: %s", name)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchIndex(ctx context.Context, name string) (*models.Index, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := a.es.Indices.Get.WithFlatSettings(true)
	res, err := a.es.Indices.Get([]string{name}, req, a.es.Indices.Get.WithContext(ctx))
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

func (a *ApiClient) DeleteElasticsearchIndexAlias(ctx context.Context, index string, aliases []string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Indices.DeleteAlias([]string{index}, aliases, a.es.Indices.DeleteAlias.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete aliases '%v' for index '%s'", index, aliases)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) UpdateElasticsearchIndexAlias(ctx context.Context, index string, alias *models.IndexAlias) diag.Diagnostics {
	var diags diag.Diagnostics
	aliasBytes, err := json.Marshal(alias)
	if err != nil {
		diag.FromErr(err)
	}
	req := a.es.Indices.PutAlias.WithBody(bytes.NewReader(aliasBytes))
	res, err := a.es.Indices.PutAlias([]string{index}, alias.Name, req, a.es.Indices.PutAlias.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to update alias '%v' for index '%s'", index, alias.Name)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) UpdateElasticsearchIndexSettings(ctx context.Context, index string, settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		diag.FromErr(err)
	}
	req := a.es.Indices.PutSettings.WithIndex(index)
	res, err := a.es.Indices.PutSettings(bytes.NewReader(settingsBytes), req, a.es.Indices.PutSettings.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update index settings"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) UpdateElasticsearchIndexMappings(ctx context.Context, index, mappings string) diag.Diagnostics {
	var diags diag.Diagnostics
	req := a.es.Indices.PutMapping.WithIndex(index)
	res, err := a.es.Indices.PutMapping(strings.NewReader(mappings), req, a.es.Indices.PutMapping.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update index mappings"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchDataStream(ctx context.Context, dataStreamName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.Indices.CreateDataStream(dataStreamName, a.es.Indices.CreateDataStream.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create DataStream: %s", dataStreamName)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchDataStream(ctx context.Context, dataStreamName string) (*models.DataStream, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Indices.GetDataStream.WithName(dataStreamName)
	res, err := a.es.Indices.GetDataStream(req, a.es.Indices.GetDataStream.WithContext(ctx))
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
	// if the DataStream found in must be the first index in the data_stream object
	ds := dStreams["data_streams"][0]
	return &ds, diags
}

func (a *ApiClient) DeleteElasticsearchDataStream(ctx context.Context, dataStreamName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.Indices.DeleteDataStream([]string{dataStreamName}, a.es.Indices.DeleteDataStream.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete DataStream: %s", dataStreamName)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) PutElasticsearchIngestPipeline(ctx context.Context, pipeline *models.IngestPipeline) diag.Diagnostics {
	var diags diag.Diagnostics
	pipelineBytes, err := json.Marshal(pipeline)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := a.es.Ingest.PutPipeline(pipeline.Name, bytes.NewReader(pipelineBytes), a.es.Ingest.PutPipeline.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create or update ingest pipeline: %s", pipeline.Name)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchIngestPipeline(ctx context.Context, name *string) (*models.IngestPipeline, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Ingest.GetPipeline.WithPipelineID(*name)
	res, err := a.es.Ingest.GetPipeline(req, a.es.Ingest.GetPipeline.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get requested ingest pipeline: %s", *name)); diags.HasError() {
		return nil, diags
	}

	pipelines := make(map[string]models.IngestPipeline)
	if err := json.NewDecoder(res.Body).Decode(&pipelines); err != nil {
		return nil, diag.FromErr(err)
	}
	pipeline := pipelines[*name]
	pipeline.Name = *name

	return &pipeline, diags
}

func (a *ApiClient) DeleteElasticsearchIngestPipeline(ctx context.Context, name *string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.Ingest.DeletePipeline(*name, a.es.Ingest.DeletePipeline.WithContext(ctx))
	if err != nil {
		return diags
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete ingest pipeline: %s", *name)); diags.HasError() {
		return diags
	}
	return diags
}
