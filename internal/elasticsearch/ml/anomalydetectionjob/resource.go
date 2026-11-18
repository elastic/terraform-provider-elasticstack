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
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = newAnomalyDetectionJobResource()
	_ resource.ResourceWithConfigure      = newAnomalyDetectionJobResource()
	_ resource.ResourceWithImportState    = newAnomalyDetectionJobResource()
	_ resource.ResourceWithValidateConfig = newAnomalyDetectionJobResource()
)

type anomalyDetectionJobResource struct {
	*entitycore.ElasticsearchResource[TFModel]
	*entitycore.CompositeIDImporter
}

func newAnomalyDetectionJobResource() *anomalyDetectionJobResource {
	return &anomalyDetectionJobResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[TFModel]("ml_anomaly_detection_job", entitycore.ElasticsearchResourceOptions[TFModel]{
			Schema: getSchema,
			Read:   readAnomalyDetectionJob,
			Delete: deleteAnomalyDetectionJob,
			Create: createAnomalyDetectionJob,
			Update: updateAnomalyDetectionJob,
			Timeouts: entitycore.ResourceTimeouts{
				Delete: 20 * time.Minute,
			},
		}),
		// Import is intentionally sparse: only IDs are set. Everything else is populated by Read().
		CompositeIDImporter: entitycore.NewCompositeIDImporter(path.Root("id"), path.Root("job_id")),
	}
}

func NewAnomalyDetectionJobResource() resource.Resource {
	return newAnomalyDetectionJobResource()
}

// ValidateConfig rejects custom rules with no scope and no conditions when both are known at plan time.
func (r *anomalyDetectionJobResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config TFModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateConfigCustomRules(ctx, &config)...)
}
