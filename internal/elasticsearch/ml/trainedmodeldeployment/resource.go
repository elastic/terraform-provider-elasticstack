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

package trainedmodeldeployment

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newTrainedModelDeploymentResource()
	_ resource.ResourceWithConfigure   = newTrainedModelDeploymentResource()
	_ resource.ResourceWithImportState = newTrainedModelDeploymentResource()
)

type trainedModelDeploymentResource struct {
	*entitycore.ElasticsearchResource[TrainedModelDeploymentData]
	*entitycore.CompositeIDImporter
}

func newTrainedModelDeploymentResource() *trainedModelDeploymentResource {
	return &trainedModelDeploymentResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[TrainedModelDeploymentData]("ml_trained_model_deployment", entitycore.ElasticsearchResourceOptions[TrainedModelDeploymentData]{
			Schema: GetSchema,
			Read:   readTrainedModelDeployment,
			Delete: deleteTrainedModelDeployment,
			Create: createTrainedModelDeployment,
			Update: updateTrainedModelDeployment,
		}),
		CompositeIDImporter: entitycore.NewCompositeIDImporter(path.Root("id"), path.Root("deployment_id"), path.Root("model_id")),
	}
}

func NewTrainedModelDeploymentResource() resource.Resource {
	return newTrainedModelDeploymentResource()
}
