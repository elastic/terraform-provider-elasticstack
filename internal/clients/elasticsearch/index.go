package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutIlm(ctx context.Context, apiClient *clients.ApiClient, policy *models.Policy) diag.Diagnostics {
	var diags diag.Diagnostics
	policyBytes, err := json.Marshal(map[string]interface{}{"policy": policy})
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	req := esClient.ILM.PutLifecycle.WithBody(bytes.NewReader(policyBytes))
	res, err := esClient.ILM.PutLifecycle(policy.Name, req, esClient.ILM.PutLifecycle.WithContext(ctx))
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
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.ILM.GetLifecycle.WithPolicy(policyName)
	res, err := esClient.ILM.GetLifecycle(req, esClient.ILM.GetLifecycle.WithContext(ctx))
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

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.ILM.DeleteLifecycle(policyName, esClient.ILM.DeleteLifecycle.WithContext(ctx))
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

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Cluster.PutComponentTemplate(template.Name, bytes.NewReader(templateBytes), esClient.Cluster.PutComponentTemplate.WithContext(ctx))
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
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Cluster.GetComponentTemplate.WithName(templateName)
	res, err := esClient.Cluster.GetComponentTemplate(req, esClient.Cluster.GetComponentTemplate.WithContext(ctx))
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
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Cluster.DeleteComponentTemplate(templateName, esClient.Cluster.DeleteComponentTemplate.WithContext(ctx))
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

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Indices.PutIndexTemplate(template.Name, bytes.NewReader(templateBytes), esClient.Indices.PutIndexTemplate.WithContext(ctx))
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
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Indices.GetIndexTemplate.WithName(templateName)
	res, err := esClient.Indices.GetIndexTemplate(req, esClient.Indices.GetIndexTemplate.WithContext(ctx))
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
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Indices.DeleteIndexTemplate(templateName, esClient.Indices.DeleteIndexTemplate.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete index template"); diags.HasError() {
		return diags
	}
	return diags
}

func PutIndex(ctx context.Context, apiClient *clients.ApiClient, index *models.Index, params *models.PutIndexParams) fwdiags.Diagnostics {
	indexBytes, err := json.Marshal(index)
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	opts := []func(*esapi.IndicesCreateRequest){
		esClient.Indices.Create.WithBody(bytes.NewReader(indexBytes)),
		esClient.Indices.Create.WithContext(ctx),
		esClient.Indices.Create.WithWaitForActiveShards(params.WaitForActiveShards),
		esClient.Indices.Create.WithMasterTimeout(params.MasterTimeout),
		esClient.Indices.Create.WithTimeout(params.Timeout),
	}
	res, err := esClient.Indices.Create(
		index.Name,
		opts...,
	)
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	defer res.Body.Close()
	diags := utils.CheckError(res, fmt.Sprintf("Unable to create index: %s", index.Name))
	return utils.FrameworkDiagsFromSDK(diags)
}

func DeleteIndex(ctx context.Context, apiClient *clients.ApiClient, name string) fwdiags.Diagnostics {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	res, err := esClient.Indices.Delete([]string{name}, esClient.Indices.Delete.WithContext(ctx))
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	defer res.Body.Close()
	diags := utils.CheckError(res, fmt.Sprintf("Unable to delete the index: %s", name))
	return utils.FrameworkDiagsFromSDK(diags)
}

func GetIndex(ctx context.Context, apiClient *clients.ApiClient, name string) (*models.Index, fwdiags.Diagnostics) {
	indices, diags := GetIndices(ctx, apiClient, name)
	if diags.HasError() {
		return nil, diags
	}

	if index, ok := indices[name]; ok {
		return &index, nil
	}

	return nil, nil
}

func GetIndices(ctx context.Context, apiClient *clients.ApiClient, name string) (map[string]models.Index, fwdiags.Diagnostics) {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	req := esClient.Indices.Get.WithFlatSettings(true)
	res, err := esClient.Indices.Get([]string{name}, req, esClient.Indices.Get.WithContext(ctx))
	if err != nil {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	defer res.Body.Close()
	// if there is no index found, return the empty struct, which should force the creation of the index
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get requested index: %s", name)); diags.HasError() {
		return nil, utils.FrameworkDiagsFromSDK(diags)
	}

	indices := make(map[string]models.Index)
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}

	return indices, nil
}

func DeleteIndexAlias(ctx context.Context, apiClient *clients.ApiClient, index string, aliases []string) fwdiags.Diagnostics {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	res, err := esClient.Indices.DeleteAlias([]string{index}, aliases, esClient.Indices.DeleteAlias.WithContext(ctx))
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	defer res.Body.Close()
	diags := utils.CheckError(res, fmt.Sprintf("Unable to delete aliases '%v' for index '%s'", index, aliases))
	return utils.FrameworkDiagsFromSDK(diags)
}

func UpdateIndexAlias(ctx context.Context, apiClient *clients.ApiClient, index string, alias *models.IndexAlias) fwdiags.Diagnostics {
	aliasBytes, err := json.Marshal(alias)
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	req := esClient.Indices.PutAlias.WithBody(bytes.NewReader(aliasBytes))
	res, err := esClient.Indices.PutAlias([]string{index}, alias.Name, req, esClient.Indices.PutAlias.WithContext(ctx))
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	defer res.Body.Close()
	diags := utils.CheckError(res, fmt.Sprintf("Unable to update alias '%v' for index '%s'", index, alias.Name))
	return utils.FrameworkDiagsFromSDK(diags)
}

func UpdateIndexSettings(ctx context.Context, apiClient *clients.ApiClient, index string, settings map[string]interface{}) fwdiags.Diagnostics {
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	req := esClient.Indices.PutSettings.WithIndex(index)
	res, err := esClient.Indices.PutSettings(bytes.NewReader(settingsBytes), req, esClient.Indices.PutSettings.WithContext(ctx))
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	defer res.Body.Close()
	diags := utils.CheckError(res, "Unable to update index settings")
	return utils.FrameworkDiagsFromSDK(diags)
}

func UpdateIndexMappings(ctx context.Context, apiClient *clients.ApiClient, index, mappings string) fwdiags.Diagnostics {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	res, err := esClient.Indices.PutMapping([]string{index}, strings.NewReader(mappings), esClient.Indices.PutMapping.WithContext(ctx))
	if err != nil {
		return fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(err.Error(), err.Error()),
		}
	}
	defer res.Body.Close()
	diags := utils.CheckError(res, "Unable to update index mappings")
	return utils.FrameworkDiagsFromSDK(diags)
}

func PutDataStream(ctx context.Context, apiClient *clients.ApiClient, dataStreamName string) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Indices.CreateDataStream(dataStreamName, esClient.Indices.CreateDataStream.WithContext(ctx))
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
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Indices.GetDataStream.WithName(dataStreamName)
	res, err := esClient.Indices.GetDataStream(req, esClient.Indices.GetDataStream.WithContext(ctx))
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

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Indices.DeleteDataStream([]string{dataStreamName}, esClient.Indices.DeleteDataStream.WithContext(ctx))
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

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Ingest.PutPipeline(pipeline.Name, bytes.NewReader(pipelineBytes), esClient.Ingest.PutPipeline.WithContext(ctx))
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
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Ingest.GetPipeline.WithPipelineID(*name)
	res, err := esClient.Ingest.GetPipeline(req, esClient.Ingest.GetPipeline.WithContext(ctx))
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

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Ingest.DeletePipeline(*name, esClient.Ingest.DeletePipeline.WithContext(ctx))
	if err != nil {
		return diags
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete ingest pipeline: %s", *name)); diags.HasError() {
		return diags
	}
	return diags
}
