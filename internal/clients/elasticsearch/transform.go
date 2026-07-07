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
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// PutTransform creates or updates a transform.
//
// We use .Raw() because the typed types.TransformDestination does not yet
// model the destination.aliases field. Passing the raw JSON preserves it.
func PutTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, transform *models.Transform, deferValidation bool, timeout time.Duration, enabled bool) fwdiag.Diagnostics {
	transformBytes, err := json.Marshal(transform)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient := apiClient.GetESClient()

	_, err = typedClient.Transform.PutTransform(transform.Name).
		Raw(bytes.NewReader(transformBytes)).
		Timeout(formatDuration(timeout)).
		DeferValidation(deferValidation).
		Do(ctx)
	if err != nil {
		return fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(fmt.Sprintf("Unable to create transform: %s", transform.Name), err.Error()),
		}
	}

	if enabled {
		if d := startTransform(ctx, apiClient, transform.Name, timeout); d.HasError() {
			return d
		}
	}

	return nil
}

// GetTransform retrieves a transform by name.
//
// We use .Perform() (raw *http.Response) instead of .Do() because the typed
// types.ReindexDestination does not yet model the destination.aliases field.
// We decode directly from the raw response body to preserve it.
func GetTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name *string) (*models.Transform, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Transform.GetTransform().TransformId(*name).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if d := diagutil.CheckHTTPErrorFromFW(res, fmt.Sprintf("Unable to get requested transform: %s", *name)); d.HasError() {
		return nil, d
	}

	var transformsResponse models.GetTransformResponse
	if err := json.NewDecoder(res.Body).Decode(&transformsResponse); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	var foundTransform *models.Transform
	for _, t := range transformsResponse.Transforms {
		if t.ID == *name {
			foundTransform = &t
			break
		}
	}

	if foundTransform == nil {
		return nil, fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(
				"Unable to find the transform in the cluster",
				fmt.Sprintf(`Unable to find "%s" transform in the cluster`, *name),
			),
		}
	}

	foundTransform.Name = *name
	return foundTransform, nil
}

func GetTransformStats(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name *string) (*types.TransformStats, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	statsRes, err := typedClient.Transform.GetTransformStats(*name).Do(ctx)
	if err != nil {
		return nil, fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(fmt.Sprintf("Unable to get transform stats: %s", *name), err.Error()),
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
		return nil, fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(
				"Unable to find the transform stats in the cluster",
				fmt.Sprintf(`Unable to find "%s" transform stats in the cluster`, *name),
			),
		}
	}

	return foundTransformStats, nil
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
) fwdiag.Diagnostics {
	transformBytes, err := json.Marshal(transform)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient := apiClient.GetESClient()

	_, err = typedClient.Transform.UpdateTransform(transform.Name).
		Raw(bytes.NewReader(transformBytes)).
		Timeout(formatDuration(timeout)).
		DeferValidation(deferValidation).
		Do(ctx)
	if err != nil {
		return fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(fmt.Sprintf("Unable to update transform: %s", transform.Name), err.Error()),
		}
	}

	if applyEnabled {
		if enabled {
			if d := startTransform(ctx, apiClient, transform.Name, timeout); d.HasError() {
				return d
			}
		} else {
			if d := stopTransform(ctx, apiClient, transform.Name, timeout); d.HasError() {
				return d
			}
		}
	}

	return nil
}

func DeleteTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name *string) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Transform.DeleteTransform(*name).Force(true).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(fmt.Sprintf("Unable to delete transform: %s", *name), err.Error()),
		}
	}

	return nil
}

func startTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, transformName string, timeout time.Duration) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Transform.StartTransform(transformName).Timeout(formatDuration(timeout)).Do(ctx)
	if err != nil {
		return fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(fmt.Sprintf("Unable to start transform: %s", transformName), err.Error()),
		}
	}

	return nil
}

func stopTransform(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, transformName string, timeout time.Duration) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Transform.StopTransform(transformName).Timeout(formatDuration(timeout)).Do(ctx)
	if err != nil {
		return fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(fmt.Sprintf("Unable to stop transform: %s", transformName), err.Error()),
		}
	}

	return nil
}
