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

package agentbuilderskill

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var createSkill = entitycore.SimpleKibanaCreate[skillModel, kbapi.PostAgentBuilderSkillsJSONRequestBody, models.Skill](
	skillModel.toAPICreateModel,
	kibanaoapi.CreateSkill,
	(*skillModel).setWriteSpaceID,
)

// setWriteSpaceID sets SpaceID explicitly so the returned model carries the
// resolved space for the envelope's read-after-write step. Shared by
// createSkill and updateSkill.
func (model *skillModel) setWriteSpaceID(_ context.Context, spaceID string, _ *models.Skill) diag.Diagnostics {
	model.SpaceID = types.StringValue(spaceID)
	return nil
}
