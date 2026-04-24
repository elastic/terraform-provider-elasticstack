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

package privatelocation

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_roundtrip verifies Terraform model → create request body mapping for private locations.
func Test_roundtrip(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		spaceID       string
		label         string
		agentPolicyID string
		tags          []string
		geo           *struct{ Lat, Lon float64 }
		wantTags      []string
		wantGeoNil    bool
	}{
		{
			name:          "only required fields",
			id:            "id-1",
			spaceID:       "",
			label:         "label-1",
			agentPolicyID: "agent-policy-id-1",
		},
		{
			name:          "all fields",
			id:            "id-2",
			spaceID:       "sample-space",
			label:         "label-2",
			agentPolicyID: "agent-policy-id-2",
			tags:          []string{"tag-1", "tag-2", "tag-3"},
			geo:           &struct{ Lat, Lon float64 }{Lat: 43.2, Lon: 23.1},
			wantTags:      []string{"tag-1", "tag-2", "tag-3"},
		},
		{
			name:          "only tags",
			id:            "id-3",
			spaceID:       "default",
			label:         "label-3",
			agentPolicyID: "agent-policy-id-3",
			tags:          []string{"tag-1", "tag-2", "tag-3"},
			wantTags:      []string{"tag-1", "tag-2", "tag-3"},
			wantGeoNil:    true,
		},
		{
			name:          "only geo",
			id:            "id-4",
			spaceID:       "",
			label:         "label-4",
			agentPolicyID: "agent-policy-id-4",
			geo:           &struct{ Lat, Lon float64 }{Lat: 43.2, Lon: 23.1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a tfModelV0 and convert to create body.
			var tagValues []types.String
			for _, tag := range tt.tags {
				tagValues = append(tagValues, types.StringValue(tag))
			}
			var geo *tfGeoConfigV0
			if tt.geo != nil {
				geo = &tfGeoConfigV0{
					Lat: NewFloat32PrecisionValue(float64(tt.geo.Lat)),
					Lon: NewFloat32PrecisionValue(float64(tt.geo.Lon)),
				}
			}
			model := tfModelV0{
				ID:            types.StringValue(tt.id),
				Label:         types.StringValue(tt.label),
				AgentPolicyID: types.StringValue(tt.agentPolicyID),
				SpaceID:       types.StringValue(tt.spaceID),
				Tags:          tagValues,
				Geo:           geo,
			}

			body := privateLocationToCreateBody(model)
			assert.Equal(t, tt.label, body.Label)
			assert.Equal(t, tt.agentPolicyID, body.AgentPolicyId)

			if len(tt.tags) > 0 {
				require.NotNil(t, body.Tags)
				assert.Equal(t, tt.tags, *body.Tags)
			} else {
				assert.Nil(t, body.Tags)
			}

			if tt.geo != nil {
				require.NotNil(t, body.Geo)
				assert.InDelta(t, float64(tt.geo.Lat), float64(body.Geo.Lat), 1e-4)
				assert.InDelta(t, float64(tt.geo.Lon), float64(body.Geo.Lon), 1e-4)
			} else {
				assert.Nil(t, body.Geo)
			}
		})
	}
}

// Test_privateLocationFromAPI_tags verifies that tags stored in AdditionalProperties round-trip correctly.
func Test_privateLocationFromAPI_tags(t *testing.T) {
	jsonPayload := `{
		"id": "loc-id",
		"label": "my-label",
		"agentPolicyId": "policy-id",
		"tags": ["env:prod", "region:us-east"]
	}`

	var loc kbapi.SyntheticsGetPrivateLocation
	require.NoError(t, json.Unmarshal([]byte(jsonPayload), &loc))

	model := privateLocationFromAPI(loc, "my-space", types.List{})

	assert.Equal(t, "loc-id", model.ID.ValueString())
	assert.Equal(t, "my-label", model.Label.ValueString())
	assert.Equal(t, "policy-id", model.AgentPolicyID.ValueString())
	assert.Equal(t, "my-space", model.SpaceID.ValueString())
	assert.Nil(t, model.Geo)

	require.Len(t, model.Tags, 2)
	assert.Equal(t, "env:prod", model.Tags[0].ValueString())
	assert.Equal(t, "region:us-east", model.Tags[1].ValueString())
}

// Test_privateLocationFromAPI_geo verifies that geo coordinates round-trip through the generated client model.
func Test_privateLocationFromAPI_geo(t *testing.T) {
	jsonPayload := `{
		"id": "geo-loc",
		"label": "geo-label",
		"agentPolicyId": "geo-policy",
		"geo": {"lat": 48.8566, "lon": 2.3522}
	}`

	var loc kbapi.SyntheticsGetPrivateLocation
	require.NoError(t, json.Unmarshal([]byte(jsonPayload), &loc))

	model := privateLocationFromAPI(loc, "", types.List{})

	require.NotNil(t, model.Geo)
	assert.InDelta(t, 48.8566, model.Geo.Lat.ValueFloat64(), 1e-9)
	assert.InDelta(t, 2.3522, model.Geo.Lon.ValueFloat64(), 1e-9)
	assert.Empty(t, model.Tags)
}

// Test_privateLocationFromAPI_tagsAndGeo verifies that both tags and geo fields are correctly extracted.
func Test_privateLocationFromAPI_tagsAndGeo(t *testing.T) {
	jsonPayload := `{
		"id": "full-loc",
		"label": "full-label",
		"agentPolicyId": "full-policy",
		"tags": ["a", "b"],
		"geo": {"lat": 10.5, "lon": -20.3}
	}`

	var loc kbapi.SyntheticsGetPrivateLocation
	require.NoError(t, json.Unmarshal([]byte(jsonPayload), &loc))

	model := privateLocationFromAPI(loc, "space-x", types.List{})

	assert.Equal(t, "full-loc", model.ID.ValueString())
	require.Len(t, model.Tags, 2)
	assert.Equal(t, "a", model.Tags[0].ValueString())
	assert.Equal(t, "b", model.Tags[1].ValueString())
	require.NotNil(t, model.Geo)
	assert.InDelta(t, 10.5, model.Geo.Lat.ValueFloat64(), 1e-9)
	assert.InDelta(t, -20.3, model.Geo.Lon.ValueFloat64(), 1e-9)
}

// Test_normalizePostResponse verifies that a POST response body (map[string]interface{}) can be
// re-encoded and decoded into SyntheticsGetPrivateLocation, matching the create normalization path.
func Test_normalizePostResponse(t *testing.T) {
	// Simulate the POST response map that would come back from the generated client.
	postResponseMap := map[string]any{
		"id":            "new-loc-id",
		"label":         "new-label",
		"agentPolicyId": "new-policy",
		"tags":          []any{"created", "fresh"},
		"geo":           map[string]any{"lat": float64(51.5074), "lon": float64(-0.1278)},
	}

	rawJSON, err := json.Marshal(postResponseMap)
	require.NoError(t, err)

	var loc kbapi.SyntheticsGetPrivateLocation
	require.NoError(t, json.Unmarshal(rawJSON, &loc))

	assert.Equal(t, "new-loc-id", *loc.Id)
	assert.Equal(t, "new-label", *loc.Label)
	assert.Equal(t, "new-policy", *loc.AgentPolicyId)

	// tags end up in AdditionalProperties
	tags := tagsFromAdditionalProperties(loc)
	require.Len(t, tags, 2)
	assert.Equal(t, "created", tags[0])
	assert.Equal(t, "fresh", tags[1])

	require.NotNil(t, loc.Geo)
	assert.InDelta(t, 51.5074, loc.Geo.Lat, 1e-9)
	assert.InDelta(t, -0.1278, loc.Geo.Lon, 1e-9)
}
