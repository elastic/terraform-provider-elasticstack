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

package anomalydetectionjob

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/jobstate"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// waitForJobClosed polls the job's state until it reports "closed" or is no
// longer found. A nil stats result (job not found) is treated as settled.
// The wait is bounded by the Terraform operation context (delete timeout).
func waitForJobClosed(ctx context.Context, client *clients.ElasticsearchScopedClient, jobID string) fwdiags.Diagnostics {
	_, diags := ml.WaitForResourceState(ctx, "ml_job", jobID, ml.WaitForResourceStateConfig[jobstate.JobState]{
		Get: func(ctx context.Context) (*jobstate.JobState, fwdiags.Diagnostics) {
			stats, diags := elasticsearch.GetMLJobStats(ctx, client, jobID)
			if diags.HasError() || stats == nil {
				return nil, diags
			}
			return &stats.State, diags
		},
		Desired:           jobstate.Closed,
		NotFoundIsDesired: true,
	})
	return diags
}
