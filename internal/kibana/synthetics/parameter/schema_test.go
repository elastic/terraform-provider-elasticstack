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
			namespaces: []string{"ns1"},
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
				Description:       utils.Pointer("description-2"),
				Tags:              utils.Pointer([]string{"tag-1", "tag-2", "tag-3"}),
				ShareAcrossSpaces: utils.Pointer(true),
			},
		},
		{
			name:       "only description",
			id:         "id-3",
			namespaces: []string{"ns3"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-3",
				Value:       "value-3",
				Description: utils.Pointer("description-3"),
			},
		},
		{
			name:       "only tags",
			id:         "id-4",
			namespaces: []string{"ns4"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-4",
				Value:       "value-4",
				Description: utils.Pointer("description-4"),
			},
		},
		{
			name:       "all namespaces",
			id:         "id-5",
			namespaces: []string{"ns5"},
			request: kboapi.SyntheticsParameterRequest{
				Key:         "key-5",
				Value:       "value-5",
				Description: utils.Pointer("description-5"),
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
			assert.Equal(t, utils.DefaultIfNil(tt.request.Description), utils.DefaultIfNil(actual.Description))
			assert.Equal(t, utils.NonNilSlice(utils.DefaultIfNil(tt.request.Tags)), utils.NonNilSlice(utils.DefaultIfNil(actual.Tags)))
			assert.Equal(t, utils.DefaultIfNil(tt.request.ShareAcrossSpaces), utils.DefaultIfNil(actual.ShareAcrossSpaces))
		})
	}
}
