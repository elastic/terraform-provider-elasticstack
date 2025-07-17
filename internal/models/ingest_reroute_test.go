package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessorReroute_JSON(t *testing.T) {
	processor := &ProcessorReroute{
		Destination: "logs-generic-default",
		Dataset:     "generic",
		Namespace:   "default",
	}
	processor.IgnoreFailure = false

	// Marshal to JSON
	processorJson, err := json.Marshal(map[string]*ProcessorReroute{"reroute": processor})
	assert.NoError(t, err)

	// Unmarshal back to verify structure
	var result map[string]map[string]interface{}
	err = json.Unmarshal(processorJson, &result)
	assert.NoError(t, err)

	// Verify the structure
	assert.Contains(t, result, "reroute")
	reroute := result["reroute"]
	assert.Equal(t, "logs-generic-default", reroute["destination"])
	assert.Equal(t, "generic", reroute["dataset"])
	assert.Equal(t, "default", reroute["namespace"])
	assert.Equal(t, false, reroute["ignore_failure"])
}
