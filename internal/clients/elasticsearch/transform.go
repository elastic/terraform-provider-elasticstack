package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var apiOperationTimeoutParamMinSupportedVersion = version.Must(version.NewVersion("7.17.0"))

func PutTransform(ctx context.Context, apiClient *clients.ApiClient, transform *models.Transform, params *models.PutTransformParams) diag.Diagnostics {

	var diags diag.Diagnostics
	transformBytes, err := json.Marshal(transform)
	if err != nil {
		return diag.FromErr(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	serverVersion, diags := apiClient.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	withTimeout := serverVersion.GreaterThanOrEqual(apiOperationTimeoutParamMinSupportedVersion)

	putOptions := []func(*esapi.TransformPutTransformRequest){
		esClient.TransformPutTransform.WithContext(ctx),
		esClient.TransformPutTransform.WithDeferValidation(params.DeferValidation),
	}

	if withTimeout {
		putOptions = append(putOptions, esClient.TransformPutTransform.WithTimeout(params.Timeout))
	}

	res, err := esClient.TransformPutTransform(bytes.NewReader(transformBytes), transform.Name, putOptions...)
	if err != nil {
		return diag.FromErr(err)
	}

	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to create transform: %s", transform.Name)); diags.HasError() {
		return diags
	}

	if params.Enabled {

		var timeout time.Duration
		if withTimeout {
			timeout = params.Timeout
		} else {
			timeout = 0
		}

		if diags := startTransform(ctx, esClient, transform.Name, timeout); diags.HasError() {
			return diags
		}
	}

	return diags
}

func GetTransform(ctx context.Context, apiClient *clients.ApiClient, name *string) (*models.Transform, diag.Diagnostics) {

	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	req := esClient.TransformGetTransform.WithTransformID(*name)
	res, err := esClient.TransformGetTransform(req, esClient.TransformGetTransform.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get requested transform: %s", *name)); diags.HasError() {
		return nil, diags
	}

	transformsResponse := models.GetTransformResponse{}
	if err := json.NewDecoder(res.Body).Decode(&transformsResponse); err != nil {
		return nil, diag.FromErr(err)
	}

	for _, t := range transformsResponse.Transforms {
		if t.Id == *name {
			t.Name = *name
			return &t, diags
		}
	}

	return nil, diags
}

func UpdateTransform(ctx context.Context, apiClient *clients.ApiClient, transform *models.Transform, params *models.UpdateTransformParams) diag.Diagnostics {

	var diags diag.Diagnostics
	transformBytes, err := json.Marshal(transform)
	if err != nil {
		return diag.FromErr(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	serverVersion, diags := apiClient.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	withTimeout := serverVersion.GreaterThanOrEqual(apiOperationTimeoutParamMinSupportedVersion)

	updateOptions := []func(*esapi.TransformUpdateTransformRequest){
		esClient.TransformUpdateTransform.WithContext(ctx),
		esClient.TransformUpdateTransform.WithDeferValidation(params.DeferValidation),
	}

	if withTimeout {
		updateOptions = append(updateOptions, esClient.TransformUpdateTransform.WithTimeout(params.Timeout))
	}

	res, err := esClient.TransformUpdateTransform(bytes.NewReader(transformBytes), transform.Name, updateOptions...)
	if err != nil {
		return diag.FromErr(err)
	}

	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to update transform: %s", transform.Name)); diags.HasError() {
		return diags
	}

	var timeout time.Duration
	if withTimeout {
		timeout = params.Timeout
	} else {
		timeout = 0
	}

	if params.Enabled {
		if diags := startTransform(ctx, esClient, transform.Name, timeout); diags.HasError() {
			return diags
		}
	} else {
		if diags := stopTransform(ctx, esClient, transform.Name, timeout); diags.HasError() {
			return diags
		}
	}

	return diags
}

func DeleteTransform(ctx context.Context, apiClient *clients.ApiClient, name *string) diag.Diagnostics {

	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := esClient.TransformDeleteTransform(*name, esClient.TransformDeleteTransform.WithForce(true), esClient.TransformDeleteTransform.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete transform: %s", *name)); diags.HasError() {
		return diags
	}

	return diags
}

func startTransform(ctx context.Context, esClient *elasticsearch.Client, transformName string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	startOptions := []func(*esapi.TransformStartTransformRequest){
		esClient.TransformStartTransform.WithContext(ctx),
	}

	if timeout > 0 {
		startOptions = append(startOptions, esClient.TransformStartTransform.WithTimeout(timeout))
	}

	startRes, err := esClient.TransformStartTransform(transformName, startOptions...)
	if err != nil {
		return diag.FromErr(err)
	}

	defer startRes.Body.Close()
	if diags := utils.CheckError(startRes, fmt.Sprintf("Unable to start transform: %s", transformName)); diags.HasError() {
		return diags
	}

	return diags
}

func stopTransform(ctx context.Context, esClient *elasticsearch.Client, transformName string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	stopOptions := []func(*esapi.TransformStopTransformRequest){
		esClient.TransformStopTransform.WithContext(ctx),
	}

	if timeout > 0 {
		stopOptions = append(stopOptions, esClient.TransformStopTransform.WithTimeout(timeout))
	}

	startRes, err := esClient.TransformStopTransform(transformName, stopOptions...)
	if err != nil {
		return diag.FromErr(err)
	}

	defer startRes.Body.Close()
	if diags := utils.CheckError(startRes, fmt.Sprintf("Unable to stop transform: %s", transformName)); diags.HasError() {
		return diags
	}

	return diags
}
