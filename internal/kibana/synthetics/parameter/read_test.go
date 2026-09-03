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

package parameter

import (
	"context"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadParameter_legacyBareUUIDUsesDefaultSpaceAndRewritesIdentity(t *testing.T) {
	originalGetParameterAPI := getParameterAPI
	t.Cleanup(func() {
		getParameterAPI = originalGetParameterAPI
	})

	const parameterUUID = "legacy-uuid"
	legacy := Model{
		ID:      types.StringValue(parameterUUID),
		SpaceID: types.StringNull(),
	}

	getParameterAPI = func(_ context.Context, _ *clients.KibanaScopedClient, resourceID, spaceID string) (*kbapi.GetParameterResponse, error) {
		assert.Equal(t, parameterUUID, resourceID)
		assert.Empty(t, spaceID)
		return &kbapi.GetParameterResponse{
			HTTPResponse: &http.Response{StatusCode: http.StatusOK},
			JSON200: &kbapi.SyntheticsGetParameterResponse{
				Id:  new(parameterUUID),
				Key: new("key"),
			},
		}, nil
	}

	result, found, diags := readParameter(
		context.Background(),
		&clients.KibanaScopedClient{},
		legacy.GetResourceID().ValueString(),
		legacy.GetSpaceID().ValueString(),
		legacy,
	)

	require.False(t, diags.HasError())
	require.True(t, found)
	assert.Equal(t, clients.DefaultSpaceID, result.SpaceID.ValueString())
	assert.Equal(t, clients.DefaultSpaceID+"/"+parameterUUID, result.ID.ValueString())
}
