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
	"testing"

	kboapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestModel_GetResourceID_compositeID(t *testing.T) {
	t.Parallel()

	m := Model{ID: types.StringValue("ops-team/param-uuid")}
	require.Equal(t, "param-uuid", m.GetResourceID().ValueString())
}

func TestModel_GetResourceID_legacyBareUUID(t *testing.T) {
	t.Parallel()

	m := Model{ID: types.StringValue("legacy-uuid-only")}
	require.Equal(t, "legacy-uuid-only", m.GetResourceID().ValueString())
}

func TestModel_GetSpaceID(t *testing.T) {
	t.Parallel()

	m := Model{SpaceID: types.StringValue(clients.DefaultSpaceID)}
	require.Equal(t, clients.DefaultSpaceID, m.GetSpaceID().ValueString())
}

func TestModel_toParameterRequest_updateOmitsShareAcrossSpaces(t *testing.T) {
	t.Parallel()

	m := Model{ShareAcrossSpaces: types.BoolValue(true)}
	require.Nil(t, m.toParameterRequest(true).ShareAcrossSpaces)
}

func TestModelFromOAPI_setsCompositeIDAndSpaceID(t *testing.T) {
	t.Parallel()

	paramUUID := "abc-123"
	m := modelFromOAPI(kboapi.SyntheticsGetParameterResponse{
		Id:  &paramUUID,
		Key: new("my-key"),
	}, "my-space")

	require.Equal(t, "my-space/abc-123", m.ID.ValueString())
	require.Equal(t, "my-space", m.SpaceID.ValueString())
	require.Equal(t, "abc-123", m.GetResourceID().ValueString())
}

func TestModelFromOAPI_emptySpaceIDDefaultsToDefaultSpace(t *testing.T) {
	t.Parallel()

	paramUUID := "abc-123"
	m := modelFromOAPI(kboapi.SyntheticsGetParameterResponse{
		Id:  &paramUUID,
		Key: new("my-key"),
	}, "")

	require.Equal(t, clients.DefaultSpaceID, m.SpaceID.ValueString())
	require.Equal(t, clients.DefaultSpaceID+"/abc-123", m.ID.ValueString())
}

func TestModel_setCompositeIdentity_emptySpaceDefaultsToDefault(t *testing.T) {
	t.Parallel()

	var m Model
	m.setCompositeIdentity("", "uuid-1")

	require.Equal(t, clients.DefaultSpaceID, m.SpaceID.ValueString())
	require.Equal(t, clients.DefaultSpaceID+"/uuid-1", m.ID.ValueString())
}

func TestModel_setCompositeIdentity_namedSpace(t *testing.T) {
	t.Parallel()

	var m Model
	m.setCompositeIdentity("ops-team", "uuid-1")

	require.Equal(t, "ops-team", m.SpaceID.ValueString())
	require.Equal(t, "ops-team/uuid-1", m.ID.ValueString())
}
