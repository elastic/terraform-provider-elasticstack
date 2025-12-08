package agent_policy

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMergeAgentFeature(t *testing.T) {
	tests := []struct {
		name       string
		existing   []apiAgentFeature
		newFeature *apiAgentFeature
		want       *[]apiAgentFeature
	}{
		{
			name:       "nil new feature with empty existing returns nil",
			existing:   nil,
			newFeature: nil,
			want:       nil,
		},
		{
			name:       "nil new feature with empty slice returns nil",
			existing:   []apiAgentFeature{},
			newFeature: nil,
			want:       nil,
		},
		{
			name: "nil new feature preserves existing features",
			existing: []apiAgentFeature{
				{Name: "feature1", Enabled: true},
				{Name: "feature2", Enabled: false},
			},
			newFeature: nil,
			want: &[]apiAgentFeature{
				{Name: "feature1", Enabled: true},
				{Name: "feature2", Enabled: false},
			},
		},
		{
			name:       "new feature added to empty existing",
			existing:   nil,
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: true},
			},
		},
		{
			name: "new feature added when not present",
			existing: []apiAgentFeature{
				{Name: "other", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "other", Enabled: true},
				{Name: "fqdn", Enabled: true},
			},
		},
		{
			name: "existing feature replaced",
			existing: []apiAgentFeature{
				{Name: "fqdn", Enabled: false},
				{Name: "other", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: true},
				{Name: "other", Enabled: true},
			},
		},
		{
			name: "feature disabled replaces enabled",
			existing: []apiAgentFeature{
				{Name: "fqdn", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: false},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeAgentFeature(tt.existing, tt.newFeature)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

func TestConvertHostNameFormatToAgentFeature(t *testing.T) {
	tests := []struct {
		name           string
		hostNameFormat types.String
		want           *apiAgentFeature
	}{
		{
			name:           "null host_name_format returns nil",
			hostNameFormat: types.StringNull(),
			want:           nil,
		},
		{
			name:           "unknown host_name_format returns nil",
			hostNameFormat: types.StringUnknown(),
			want:           nil,
		},
		{
			name:           "fqdn returns enabled feature",
			hostNameFormat: types.StringValue(HostNameFormatFQDN),
			want:           &apiAgentFeature{Name: agentFeatureFQDN, Enabled: true},
		},
		{
			name:           "hostname returns disabled feature",
			hostNameFormat: types.StringValue(HostNameFormatHostname),
			want:           &apiAgentFeature{Name: agentFeatureFQDN, Enabled: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &agentPolicyModel{
				HostNameFormat: tt.hostNameFormat,
			}

			got := model.convertHostNameFormatToAgentFeature()

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Enabled, got.Enabled)
		})
	}
}

