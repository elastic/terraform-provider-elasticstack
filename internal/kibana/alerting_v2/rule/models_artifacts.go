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

package rule

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func artifactsToAPI(artifacts []artifactModel, diags *diag.Diagnostics) *[]kbapi.KibanaHTTPAPIsAlertingRuleArtifact {
	if artifacts == nil {
		return nil
	}
	out := make([]kbapi.KibanaHTTPAPIsAlertingRuleArtifact, 0, len(artifacts))
	for i, artifact := range artifacts {
		out = append(out, kbapi.KibanaHTTPAPIsAlertingRuleArtifact{
			Id:   artifact.ID.ValueString(),
			Type: artifact.Type.ValueString(),
			Data: typeutils.NormalizedTypeToMap[any](artifact.Data, artifactDataPath(i), diags),
		})
	}
	return &out
}

func artifactsFromAPI(artifacts *[]kbapi.KibanaHTTPAPIsAlertingRuleArtifact, diags *diag.Diagnostics) []artifactModel {
	if artifacts == nil {
		return nil
	}
	out := make([]artifactModel, 0, len(*artifacts))
	for i, artifact := range *artifacts {
		out = append(out, artifactModel{
			ID:   types.StringValue(artifact.Id),
			Type: types.StringValue(artifact.Type),
			Data: typeutils.MarshalToNormalized(artifact.Data, artifactDataPath(i), diags),
		})
	}
	return out
}

func artifactDataPath(index int) path.Path {
	return artifactsPath.AtListIndex(index).AtName("data")
}
