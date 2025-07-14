package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessorReroute_JSON(t *testing.T) {
	processor := &ProcessorReroute{
		Field:         "routing_field",
		IgnoreMissing: false,
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
	assert.Equal(t, "routing_field", reroute["field"])
	assert.Equal(t, false, reroute["ignore_failure"])
	assert.Equal(t, false, reroute["ignore_missing"])
}
