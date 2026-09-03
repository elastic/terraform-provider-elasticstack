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
	"reflect"
	"strings"
	"testing"

	kboapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_roundtrip(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		namespaces     []string
		omitNamespaces bool
		request        kboapi.SyntheticsParameterRequest
	}{
		{
			name:       "only required fields",
			id:         "id-1",
			namespaces: []string{"ns-1"},
			request: kboapi.SyntheticsParameterRequest{
				Key:   "key-1",
				Value: "value-1",
			},
		},
		{
			name:       "all fields",
			id:         "id-2",
			namespaces: []string{"*"},
			request: kboapi.SyntheticsParameterRequest{
				Key:               "key-2",
				Value:             "value-2",
				Description:       new("description-2"),
				Tags:              new([]string{"tag-1", "tag-2", "tag-3"}),
				ShareAcrossSpaces: new(true),
			},
		},
		{
			name:       "only description",
			id:         "id-3",
			namespaces: []string{"ns-3"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-3",
				Value:       "value-3",
				Description: new("description-3"),
			},
		},
		{
			name:       "only tags",
			id:         "id-4",
			namespaces: []string{"ns-4"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-4",
				Value:       "value-4",
				Description: new("description-4"),
			},
		},
		{
			name:       "all namespaces",
			id:         "id-5",
			namespaces: []string{"ns-5"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-5",
				Value:       "value-5",
				Description: new("description-5"),
			},
		},
		{
			// Kibana omits `namespaces` for callers without read-only permissions.
			// modelFromOAPI must not dereference a nil pointer in that case.
			name:           "namespaces omitted",
			id:             "id-6",
			omitNamespaces: true,
			request: kboapi.SyntheticsParameterRequest{
				Key:   "key-6",
				Value: "value-6",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := kboapi.SyntheticsGetParameterResponse{
				Id:          &tt.id,
				Key:         &tt.request.Key,
				Value:       &tt.request.Value,
				Description: tt.request.Description,
				Tags:        tt.request.Tags,
			}
			if !tt.omitNamespaces {
				response.Namespaces = &tt.namespaces
			}
			m := modelFromOAPI(response, clients.DefaultSpaceID)

			assert.Equal(t, clients.DefaultSpaceID, m.SpaceID.ValueString())
			assert.Equal(t, clients.DefaultSpaceID+"/"+tt.id, m.ID.ValueString())
			assert.Equal(t, tt.id, m.GetResourceID().ValueString())

			actual := m.toParameterRequest(false)

			assert.Equal(t, tt.request.Key, actual.Key)
			assert.Equal(t, tt.request.Value, actual.Value)
			assert.Equal(t, typeutils.Deref(tt.request.Description), typeutils.Deref(actual.Description))
			assert.Equal(t, typeutils.NonNilSlice(typeutils.Deref(tt.request.Tags)), typeutils.NonNilSlice(typeutils.Deref(actual.Tags)))
			assert.Equal(t, typeutils.Deref(tt.request.ShareAcrossSpaces), typeutils.Deref(actual.ShareAcrossSpaces))
		})
	}
}

func TestSchema_hasSpaceIDAttribute(t *testing.T) {
	t.Parallel()

	canonical := kbschema.ResourceSpaceIDAttribute()
	spaceIDAttr, ok := getSchema(context.Background()).Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.Equal(t, canonical.MarkdownDescription, spaceIDAttr.MarkdownDescription)
	assert.True(t, spaceIDAttr.IsOptional())
	assert.True(t, spaceIDAttr.IsComputed())
	assertHasStringPlanModifier(t, spaceIDAttr.PlanModifiers, "useStateForUnknown")
	assertHasStringPlanModifier(t, spaceIDAttr.PlanModifiers, "requiresReplace")
}

func TestSchema_spaceIDDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	spaceIDAttr, ok := getSchema(ctx).Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)

	defaultHandler := spaceIDAttr.StringDefaultValue()
	require.NotNil(t, defaultHandler)

	var resp defaults.StringResponse
	defaultHandler.DefaultString(ctx, defaults.StringRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, clients.DefaultSpaceID, resp.PlanValue.ValueString())
}

func assertHasStringPlanModifier(t *testing.T, modifiers []planmodifier.String, suffix string) {
	t.Helper()
	for i := range modifiers {
		if strings.Contains(reflect.TypeOf(modifiers[i]).String(), suffix) {
			return
		}
	}
	t.Fatalf("expected plan modifier containing %q among %d modifiers", suffix, len(modifiers))
}
