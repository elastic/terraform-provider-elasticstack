package agentpolicy

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestMonitoringRuntimeExperimentalSupported(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"8.17.0", "8.17.0", false},
		{"8.18.8", "8.18.8", false},
		{"8.19.0", "8.19.0", true},
		{"8.19.5", "8.19.5", true},
		{"9.0.0", "9.0.0", false},
		{"9.0.8", "9.0.8", false},
		{"9.1.0", "9.1.0", true},
		{"9.4.0", "9.4.0", true},
		{"10.0.0", "10.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := version.NewVersion(tt.version)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, MonitoringRuntimeExperimentalSupported(v))
		})
	}
}
