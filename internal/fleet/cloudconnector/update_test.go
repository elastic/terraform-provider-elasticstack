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

package cloudconnector

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCloudConnector_skipsWhenNoMutationAndNoResubmit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := newResource()
	model := completeAWSCloudConnectorModel(t, types.StringValue("secret-a"))
	priv := newMapPrivateState()

	result, diags := r.updateCloudConnector(ctx, nil, entitycore.KibanaWriteRequest[cloudConnectorModel]{
		Plan:    model,
		Prior:   &model,
		Config:  model,
		Private: priv,
	})
	require.False(t, diags.HasError())
	assert.Equal(t, model.Name, result.Model.Name)
}

func TestCloudConnectorUpdate_skipGuard(t *testing.T) {
	t.Parallel()

	model := completeAWSCloudConnectorModel(t, types.StringNull())
	model.AWS = mustAWSBlockObjectWithRoleArn(t, "arn:aws:iam::123:role/x")
	config := completeAWSCloudConnectorModel(t, types.StringValue("new-secret"))
	config.AWS = mustAWSBlockObjectWithRoleArn(t, "arn:aws:iam::123:role/x")

	require.False(t, planMutatesAPIResource(model, model, config))

	shouldSkip := !planMutatesAPIResource(model, model, config) && len(map[string]struct{}{}) == 0
	assert.True(t, shouldSkip)

	resubmit := map[string]struct{}{writeOnlyAttributeAWSExternalID: {}}
	shouldSkipWithResubmit := !planMutatesAPIResource(model, model, config) && len(resubmit) == 0
	assert.False(t, shouldSkipWithResubmit)
}
