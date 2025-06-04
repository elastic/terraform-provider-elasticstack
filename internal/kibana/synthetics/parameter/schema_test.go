package parameter

import (
	"testing"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/stretchr/testify/assert"
)

func boolPointer(b bool) *bool {
	return &b
}

func Test_roundtrip(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		namespaces []string
		config     kbapi.ParameterConfig
	}{
		{
			name:       "only required fields",
			id:         "id-1",
			namespaces: []string{"ns-1"},
			config: kbapi.ParameterConfig{
				Key:   "key-1",
				Value: "value-1",
			},
		},
		{
			name:       "all fields",
			id:         "id-2",
			namespaces: []string{"*"},
			config: kbapi.ParameterConfig{
				Key:               "key-2",
				Value:             "value-2",
				Description:       "description-2",
				Tags:              []string{"tag-1", "tag-2", "tag-3"},
				ShareAcrossSpaces: boolPointer(true),
			},
		},
		{
			name:       "only description",
			id:         "id-3",
			namespaces: []string{"ns-3"},
			config: kbapi.ParameterConfig{
				Key:         "key-3",
				Value:       "value-3",
				Description: "description-3",
			},
		},
		{
			name:       "only tags",
			id:         "id-4",
			namespaces: []string{"ns-4"},
			config: kbapi.ParameterConfig{
				Key:         "key-4",
				Value:       "value-4",
				Description: "description-4",
			},
		},
		{
			name:       "all namespaces",
			id:         "id-5",
			namespaces: []string{"ns-5"},
			config: kbapi.ParameterConfig{
				Key:         "key-5",
				Value:       "value-5",
				Description: "description-5",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			input := kbapi.Parameter{
				Id:              tt.id,
				Namespaces:      tt.namespaces,
				ParameterConfig: config,
			}
			modelV0 := toModelV0(input)

			actual := modelV0.toParameterConfig(false)
			if config.ShareAcrossSpaces == nil {
				// The conversion always sets ShareAcrossSpaces.
				config.ShareAcrossSpaces = boolPointer(false)
			}
			assert.Equal(t, config, actual)
		})
	}
}
