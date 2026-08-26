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
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// OpenMLJob opens a machine learning job
func OpenMLJob(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, jobID string) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Ml.OpenJob(jobID).Do(ctx)
	if err != nil {
		diags.AddError("Failed to open ML job", fmt.Sprintf("Unable to open ML job: %s — %s", jobID, err.Error()))
		return diags
	}

	return diags
}

// CloseMLJob closes a machine learning job
func CloseMLJob(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, jobID string, force bool, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient := apiClient.GetESClient()

	req := typedClient.Ml.CloseJob(jobID).
		Force(force).
		AllowNoMatch(true)

	if timeout > 0 {
		req.Timeout(typeutils.DurationToElasticsearchTimeoutString(timeout))
	}

	_, err := req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to close ML job", fmt.Sprintf("Unable to close ML job: %s — %s", jobID, err.Error()))
		return diags
	}

	return diags
}

// GetMLJobStats retrieves the stats for a specific machine learning job
func GetMLJobStats(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, jobID string) (*types.JobStats, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Ml.GetJobStats().JobId(jobID).AllowNoMatch(true).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Failed to get ML job stats", fmt.Sprintf("Unable to get ML job stats: %s — %s", jobID, err.Error()))
		return nil, diags
	}

	for i := range res.Jobs {
		if res.Jobs[i].JobId == jobID {
			return &res.Jobs[i], diags
		}
	}

	return nil, diags
}
