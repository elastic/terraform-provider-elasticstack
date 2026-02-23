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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/stretchr/testify/assert"
)

func Test_roundtrip(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		namespaces []string
		request    kboapi.SyntheticsParameterRequest
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
				Description:       schemautil.Pointer("description-2"),
				Tags:              schemautil.Pointer([]string{"tag-1", "tag-2", "tag-3"}),
				ShareAcrossSpaces: schemautil.Pointer(true),
			},
		},
		{
			name:       "only description",
			id:         "id-3",
			namespaces: []string{"ns-3"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-3",
				Value:       "value-3",
				Description: schemautil.Pointer("description-3"),
			},
		},
		{
			name:       "only tags",
			id:         "id-4",
			namespaces: []string{"ns-4"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-4",
				Value:       "value-4",
				Description: schemautil.Pointer("description-4"),
			},
		},
		{
			name:       "all namespaces",
			id:         "id-5",
			namespaces: []string{"ns-5"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-5",
				Value:       "value-5",
				Description: schemautil.Pointer("description-5"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := kboapi.SyntheticsGetParameterResponse{
				Id:          &tt.id,
				Namespaces:  &tt.namespaces,
				Key:         &tt.request.Key,
				Value:       &tt.request.Value,
				Description: tt.request.Description,
				Tags:        tt.request.Tags,
			}
			modelV0 := modelV0FromOAPI(response)

			actual := modelV0.toParameterRequest(false)

			assert.Equal(t, tt.request.Key, actual.Key)
			assert.Equal(t, tt.request.Value, actual.Value)
			assert.Equal(t, schemautil.DefaultIfNil(tt.request.Description), schemautil.DefaultIfNil(actual.Description))
			assert.Equal(t, schemautil.NonNilSlice(schemautil.DefaultIfNil(tt.request.Tags)), schemautil.NonNilSlice(schemautil.DefaultIfNil(actual.Tags)))
			assert.Equal(t, schemautil.DefaultIfNil(tt.request.ShareAcrossSpaces), schemautil.DefaultIfNil(actual.ShareAcrossSpaces))
		})
	}
}
