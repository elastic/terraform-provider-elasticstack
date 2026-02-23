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

package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessorReroute_JSON(t *testing.T) {
	processor := &ProcessorReroute{
		Destination: "logs-generic-default",
		Dataset:     "generic",
		Namespace:   "default",
	}
	processor.IgnoreFailure = false

	// Marshal to JSON
	processorJSON, err := json.Marshal(map[string]*ProcessorReroute{"reroute": processor})
	require.NoError(t, err)

	// Unmarshal back to verify structure
	var result map[string]map[string]any
	err = json.Unmarshal(processorJSON, &result)
	require.NoError(t, err)

	// Verify the structure
	assert.Contains(t, result, "reroute")
	reroute := result["reroute"]
	assert.Equal(t, "logs-generic-default", reroute["destination"])
	assert.Equal(t, "generic", reroute["dataset"])
	assert.Equal(t, "default", reroute["namespace"])
	assert.Equal(t, false, reroute["ignore_failure"])
}
