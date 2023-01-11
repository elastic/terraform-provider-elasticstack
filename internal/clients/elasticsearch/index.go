package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutIlm(ctx context.Context, apiClient *clients.ApiClient, policy *models.Policy) diag.Diagnostics {
	var diags diag.Diagnostics
	policyBytes, err := json.Marshal(map[string]interface{}{"policy": policy})
	if err != nil {
		return diag.FromErr(err)
	}
	req := apiClient.GetESClient().ILM.PutLifecycle.WithBody(bytes.NewReader(policyBytes))
	res, err := apiClient.GetESClient().ILM.PutLifecycle(policy.Name, req, apiClient.GetESClient().ILM.PutLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the ILM policy"); diags.HasError() {
		return diags
	}
	return diags
}

func GetIlm(ctx context.Context, apiClient *clients.ApiClient, policyName string) (*models.PolicyDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := apiClient.GetESClient().ILM.GetLifecycle.WithPolicy(policyName)
	res, err := apiClient.GetESClient().ILM.GetLifecycle(req, apiClient.GetESClient().ILM.GetLifecycle.WithContext(ctx))
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

func DeleteIlm(ctx context.Context, apiClient *clients.ApiClient, policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := apiClient.GetESClient().ILM.DeleteLifecycle(policyName, apiClient.GetESClient().ILM.DeleteLifecycle.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete ILM policy."); diags.HasError() {
		return diags
	}
	return diags
}

func PutComponentTemplate(ctx context.Context, apiClient *clients.ApiClient, template *models.ComponentTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := apiClient.GetESClient().Cluster.PutComponentTemplate(template.Name, bytes.NewReader(templateBytes), apiClient.GetESClient().Cluster.PutComponentTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create component template"); diags.HasError() {
		return diags
	}

	return diags
}

func GetComponentTemplate(ctx context.Context, apiClient *clients.ApiClient, templateName string) (*models.ComponentTemplateResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := apiClient.GetESClient().Cluster.GetComponentTemplate.WithName(templateName)
	res, err := apiClient.GetESClient().Cluster.GetComponentTemplate(req, apiClient.GetESClient().Cluster.GetComponentTemplate.WithContext(ctx))
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

func DeleteComponentTemplate(ctx context.Context, apiClient *clients.ApiClient, templateName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := apiClient.GetESClient().Cluster.DeleteComponentTemplate(templateName, apiClient.GetESClient().Cluster.DeleteComponentTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete component template"); diags.HasError() {
		return diags
	}
	return diags
}

func PutIndexTemplate(ctx context.Context, apiClient *clients.ApiClient, template *models.IndexTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := apiClient.GetESClient().Indices.PutIndexTemplate(template.Name, bytes.NewReader(templateBytes), apiClient.GetESClient().Indices.PutIndexTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create index template"); diags.HasError() {
		return diags
	}

	return diags
}

func GetIndexTemplate(ctx context.Context, apiClient *clients.ApiClient, templateName string) (*models.IndexTemplateResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := apiClient.GetESClient().Indices.GetIndexTemplate.WithName(templateName)
	res, err := apiClient.GetESClient().Indices.GetIndexTemplate(req, apiClient.GetESClient().Indices.GetIndexTemplate.WithContext(ctx))
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
			Detail:   fmt.Sprintf("Elasticsearch API returned %d when requested '%s' template.", len(indexTemplates.IndexTemplates), templateName),
		})
		return nil, diags
	}
	tpl := indexTemplates.IndexTemplates[0]
	return &tpl, diags
}

func DeleteIndexTemplate(ctx context.Context, apiClient *clients.ApiClient, templateName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := apiClient.GetESClient().Indices.DeleteIndexTemplate(templateName, apiClient.GetESClient().Indices.DeleteIndexTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete index template"); diags.HasError() {
		return diags
	}
	return diags
}

func PutIndex(ctx context.Context, apiClient *clients.ApiClient, index *models.Index) diag.Diagnostics {
	var diags diag.Diagnostics
	indexBytes, err := json.Marshal(index)
	if err != nil {
		return diag.FromErr(err)
	}

	req := apiClient.GetESClient().Indices.Create.WithBody(bytes.NewReader(indexBytes))
	res, err := apiClient.GetESClient().Indices.Create(index.Name, req, apiClient.GetESClient().Indices.Create.WithContext(ctx))
	if err != nil {
		diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create index: %s", index.Name)); diags.HasError() {
		return diags
	}
	return diags
}

func DeleteIndex(ctx context.Context, apiClient *clients.ApiClient, name string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := apiClient.GetESClient().Indices.Delete([]string{name}, apiClient.GetESClient().Indices.Delete.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete the index: %s", name)); diags.HasError() {
		return diags
	}

	return diags
}

func GetIndex(ctx context.Context, apiClient *clients.ApiClient, name string) (*models.Index, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := apiClient.GetESClient().Indices.Get.WithFlatSettings(true)
	res, err := apiClient.GetESClient().Indices.Get([]string{name}, req, apiClient.GetESClient().Indices.Get.WithContext(ctx))
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

func DeleteIndexAlias(ctx context.Context, apiClient *clients.ApiClient, index string, aliases []string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := apiClient.GetESClient().Indices.DeleteAlias([]string{index}, aliases, apiClient.GetESClient().Indices.DeleteAlias.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete aliases '%v' for index '%s'", index, aliases)); diags.HasError() {
		return diags
	}
	return diags
}

func UpdateIndexAlias(ctx context.Context, apiClient *clients.ApiClient, index string, alias *models.IndexAlias) diag.Diagnostics {
	var diags diag.Diagnostics
	aliasBytes, err := json.Marshal(alias)
	if err != nil {
		diag.FromErr(err)
	}
	req := apiClient.GetESClient().Indices.PutAlias.WithBody(bytes.NewReader(aliasBytes))
	res, err := apiClient.GetESClient().Indices.PutAlias([]string{index}, alias.Name, req, apiClient.GetESClient().Indices.PutAlias.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to update alias '%v' for index '%s'", index, alias.Name)); diags.HasError() {
		return diags
	}
	return diags
}

func UpdateIndexSettings(ctx context.Context, apiClient *clients.ApiClient, index string, settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		diag.FromErr(err)
	}
	req := apiClient.GetESClient().Indices.PutSettings.WithIndex(index)
	res, err := apiClient.GetESClient().Indices.PutSettings(bytes.NewReader(settingsBytes), req, apiClient.GetESClient().Indices.PutSettings.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update index settings"); diags.HasError() {
		return diags
	}
	return diags
}

func UpdateIndexMappings(ctx context.Context, apiClient *clients.ApiClient, index, mappings string) diag.Diagnostics {
	var diags diag.Diagnostics
	req := apiClient.GetESClient().Indices.PutMapping.WithIndex(index)
	res, err := apiClient.GetESClient().Indices.PutMapping(strings.NewReader(mappings), req, apiClient.GetESClient().Indices.PutMapping.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update index mappings"); diags.HasError() {
		return diags
	}
	return diags
}

func PutDataStream(ctx context.Context, apiClient *clients.ApiClient, dataStreamName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := apiClient.GetESClient().Indices.CreateDataStream(dataStreamName, apiClient.GetESClient().Indices.CreateDataStream.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create DataStream: %s", dataStreamName)); diags.HasError() {
		return diags
	}

	return diags
}

func GetDataStream(ctx context.Context, apiClient *clients.ApiClient, dataStreamName string) (*models.DataStream, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := apiClient.GetESClient().Indices.GetDataStream.WithName(dataStreamName)
	res, err := apiClient.GetESClient().Indices.GetDataStream(req, apiClient.GetESClient().Indices.GetDataStream.WithContext(ctx))
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

func DeleteDataStream(ctx context.Context, apiClient *clients.ApiClient, dataStreamName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := apiClient.GetESClient().Indices.DeleteDataStream([]string{dataStreamName}, apiClient.GetESClient().Indices.DeleteDataStream.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete DataStream: %s", dataStreamName)); diags.HasError() {
		return diags
	}

	return diags
}

func PutIngestPipeline(ctx context.Context, apiClient *clients.ApiClient, pipeline *models.IngestPipeline) diag.Diagnostics {
	var diags diag.Diagnostics
	pipelineBytes, err := json.Marshal(pipeline)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := apiClient.GetESClient().Ingest.PutPipeline(pipeline.Name, bytes.NewReader(pipelineBytes), apiClient.GetESClient().Ingest.PutPipeline.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create or update ingest pipeline: %s", pipeline.Name)); diags.HasError() {
		return diags
	}

	return diags
}

func GetIngestPipeline(ctx context.Context, apiClient *clients.ApiClient, name *string) (*models.IngestPipeline, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := apiClient.GetESClient().Ingest.GetPipeline.WithPipelineID(*name)
	res, err := apiClient.GetESClient().Ingest.GetPipeline(req, apiClient.GetESClient().Ingest.GetPipeline.WithContext(ctx))
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

func DeleteIngestPipeline(ctx context.Context, apiClient *clients.ApiClient, name *string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := apiClient.GetESClient().Ingest.DeletePipeline(*name, apiClient.GetESClient().Ingest.DeletePipeline.WithContext(ctx))
	if err != nil {
		return diags
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete ingest pipeline: %s", *name)); diags.HasError() {
		return diags
	}
	return diags
}
