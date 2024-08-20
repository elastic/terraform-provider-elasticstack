package private_location

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"testing"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/stretchr/testify/assert"
)

func Test_roundtrip(t *testing.T) {
	tests := []struct {
		name string
		id   string
		ns   string
		plc  kbapi.PrivateLocationConfig
	}{
		{
			name: "only required fields",
			id:   "id-1",
			ns:   "ns-1",
			plc: kbapi.PrivateLocationConfig{
				Label:         "label-1",
				AgentPolicyId: "agent-policy-id-1",
			},
		},
		{
			name: "all fields",
			id:   "id-2",
			ns:   "ns-2",
			plc: kbapi.PrivateLocationConfig{
				Label:         "label-2",
				AgentPolicyId: "agent-policy-id-2",
				Tags:          []string{"tag-1", "tag-2", "tag-3"},
				Geo: &kbapi.SyntheticGeoConfig{
					Lat: 43.2,
					Lon: 23.1,
				},
			},
		},
		{
			name: "only tags",
			id:   "id-3",
			ns:   "ns-3",
			plc: kbapi.PrivateLocationConfig{
				Label:         "label-3",
				AgentPolicyId: "agent-policy-id-3",
				Tags:          []string{"tag-1", "tag-2", "tag-3"},
				Geo:           nil,
			},
		},
		{
			name: "only geo",
			id:   "id-4",
			ns:   "ns-4",
			plc: kbapi.PrivateLocationConfig{
				Label:         "label-4",
				AgentPolicyId: "agent-policy-id-4",
				Geo: &kbapi.SyntheticGeoConfig{
					Lat: 43.2,
					Lon: 23.1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plc := tt.plc
			input := kbapi.PrivateLocation{
				Id:                    tt.id,
				Namespace:             tt.ns,
				PrivateLocationConfig: plc,
			}
			modelV0 := toModelV0(input)

			compositeId, _ := synthetics.GetCompositeId(modelV0.ID.ValueString())

			actual := kbapi.PrivateLocation{
				Id:                    compositeId.ResourceId,
				Namespace:             modelV0.SpaceID.ValueString(),
				PrivateLocationConfig: modelV0.toPrivateLocationConfig(),
			}
			assert.Equal(t, input, actual)
		})
	}
}
