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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// formatDuration converts duration to a string in the format accepted by
// Elasticsearch, matching the legacy esapi behavior (milliseconds).
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return strconv.FormatInt(int64(d), 10) + "nanos"
	}
	return strconv.FormatInt(int64(d)/int64(time.Millisecond), 10) + "ms"
}

// PutTransform creates or updates a transform.
//
// We use .Raw() because the typed types.TransformDestination does not yet
// model the destination.aliases field. Passing the raw JSON preserves it.
func PutTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, transform *models.Transform, deferValidation bool, timeout time.Duration, enabled bool) diag.Diagnostics {

	var diags diag.Diagnostics
	transformBytes, err := json.Marshal(transform)
	if err != nil {
		return diag.FromErr(err)
	}

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Transform.PutTransform(transform.Name).
		Raw(bytes.NewReader(transformBytes)).
		Timeout(formatDuration(timeout)).
		DeferValidation(deferValidation).
		Do(ctx)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to create transform: %s", transform.Name),
				Detail:   err.Error(),
			},
		}
	}

	if enabled {
		if diags := startTransform(ctx, apiClient, transform.Name, timeout); diags.HasError() {
			return diags
		}
	}

	return diags
}

// GetTransform retrieves a transform by name.
//
// We use .Perform() (raw *http.Response) instead of .Do() because the typed
// types.ReindexDestination does not yet model the destination.aliases field.
// We decode directly from the raw response body to preserve it.
func GetTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name *string) (*models.Transform, diag.Diagnostics) {

	var diags diag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := typedClient.Transform.GetTransform().TransformId(*name).Perform(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if d := diagutil.CheckHTTPError(res, fmt.Sprintf("Unable to get requested transform: %s", *name)); d.HasError() {
		return nil, d
	}

	var transformsResponse models.GetTransformResponse
	if err := json.NewDecoder(res.Body).Decode(&transformsResponse); err != nil {
		return nil, diag.FromErr(err)
	}

	var foundTransform *models.Transform
	for _, t := range transformsResponse.Transforms {
		if t.ID == *name {
			foundTransform = &t
			break
		}
	}

	if foundTransform == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find the transform in the cluster",
			Detail:   fmt.Sprintf(`Unable to find "%s" transform in the cluster`, *name),
		})

		return nil, diags
	}

	foundTransform.Name = *name
	return foundTransform, diags
}

func GetTransformStats(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name *string) (*types.TransformStats, diag.Diagnostics) {
	var diags diag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	statsRes, err := typedClient.Transform.GetTransformStats(*name).Do(ctx)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to get transform stats: %s", *name),
				Detail:   err.Error(),
			},
		}
	}

	var foundTransformStats *types.TransformStats
	for _, ts := range statsRes.Transforms {
		if ts.Id == *name {
			foundTransformStats = &ts
			break
		}
	}

	if foundTransformStats == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find the transform stats in the cluster",
			Detail:   fmt.Sprintf(`Unable to find "%s" transform stats in the cluster`, *name),
		})
		return nil, diags
	}

	return foundTransformStats, diags
}

// UpdateTransform updates an existing transform.
//
// We use .Raw() because the typed types.TransformDestination does not yet
// model the destination.aliases field. Passing the raw JSON preserves it.
func UpdateTransform(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	transform *models.Transform,
	deferValidation bool,
	timeout time.Duration,
	enabled bool,
	applyEnabled bool,
) diag.Diagnostics {

	var diags diag.Diagnostics
	transformBytes, err := json.Marshal(transform)
	if err != nil {
		return diag.FromErr(err)
	}

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Transform.UpdateTransform(transform.Name).
		Raw(bytes.NewReader(transformBytes)).
		Timeout(formatDuration(timeout)).
		DeferValidation(deferValidation).
		Do(ctx)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to update transform: %s", transform.Name),
				Detail:   err.Error(),
			},
		}
	}

	if applyEnabled {
		if enabled {
			if diags := startTransform(ctx, apiClient, transform.Name, timeout); diags.HasError() {
				return diags
			}
		} else {
			if diags := stopTransform(ctx, apiClient, transform.Name, timeout); diags.HasError() {
				return diags
			}
		}
	}

	return diags
}

func DeleteTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name *string) diag.Diagnostics {

	var diags diag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Transform.DeleteTransform(*name).Force(true).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return diags
		}
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to delete transform: %s", *name),
				Detail:   err.Error(),
			},
		}
	}

	return diags
}

func startTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, transformName string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Transform.StartTransform(transformName).Timeout(formatDuration(timeout)).Do(ctx)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to start transform: %s", transformName),
				Detail:   err.Error(),
			},
		}
	}

	return diags
}

func stopTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, transformName string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Transform.StopTransform(transformName).Timeout(formatDuration(timeout)).Do(ctx)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to stop transform: %s", transformName),
				Detail:   err.Error(),
			},
		}
	}

	return diags
}
