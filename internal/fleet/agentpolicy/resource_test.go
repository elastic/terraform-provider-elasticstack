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

package agentpolicy

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, err)
			assert.Equal(t, tt.expected, MonitoringRuntimeExperimentalSupported(v))
		})
	}
}
