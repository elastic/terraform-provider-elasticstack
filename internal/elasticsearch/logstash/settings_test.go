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

package logstash

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestFlattenSettings(t *testing.T) {
	t.Parallel()

	apiSettings := map[string]any{
		"pipeline.batch.delay":         float64(50),
		"pipeline.batch.size":          float64(125),
		"pipeline.ecs_compatibility":   "v1",
		"pipeline.ordered":             "auto",
		"pipeline.plugin_classloaders": true,
		"pipeline.unsafe_shutdown":     false,
		"pipeline.workers":             float64(4),
		"queue.checkpoint.acks":        float64(1024),
		"queue.checkpoint.retry":       true,
		"queue.checkpoint.writes":      float64(1024),
		"queue.drain":                  false,
		"queue.max_bytes":              "1gb",
		"queue.max_events":             float64(0),
		"queue.page_capacity":          "64mb",
		"queue.type":                   "memory",
	}

	var data Data
	flattenSettings(apiSettings, &data)

	require.Equal(t, types.Int64Value(50), data.PipelineBatchDelay)
	require.Equal(t, types.Int64Value(125), data.PipelineBatchSize)
	require.Equal(t, types.StringValue("v1"), data.PipelineEcsCompatibility)
	require.Equal(t, types.StringValue("auto"), data.PipelineOrdered)
	require.Equal(t, types.BoolValue(true), data.PipelinePluginClassloaders)
	require.Equal(t, types.BoolValue(false), data.PipelineUnsafeShutdown)
	require.Equal(t, types.Int64Value(4), data.PipelineWorkers)
	require.Equal(t, types.Int64Value(1024), data.QueueCheckpointAcks)
	require.Equal(t, types.BoolValue(true), data.QueueCheckpointRetry)
	require.Equal(t, types.Int64Value(1024), data.QueueCheckpointWrites)
	require.Equal(t, types.BoolValue(false), data.QueueDrain)
	require.Equal(t, types.StringValue("1gb"), data.QueueMaxBytes)
	require.Equal(t, types.Int64Value(0), data.QueueMaxEvents)
	require.Equal(t, types.StringValue("64mb"), data.QueuePageCapacity)
	require.Equal(t, types.StringValue("memory"), data.QueueType)
}

func TestFlattenSettings_MissingKeysLeaveFieldsUnset(t *testing.T) {
	t.Parallel()

	var data Data
	flattenSettings(map[string]any{}, &data)

	require.True(t, data.PipelineBatchDelay.IsNull())
	require.True(t, data.PipelineEcsCompatibility.IsNull())
	require.True(t, data.PipelinePluginClassloaders.IsNull())
}

func TestExpandSettings(t *testing.T) {
	t.Parallel()

	data := Data{
		PipelineBatchDelay:         types.Int64Value(50),
		PipelineBatchSize:          types.Int64Value(125),
		PipelineEcsCompatibility:   types.StringValue("v1"),
		PipelineOrdered:            types.StringValue("auto"),
		PipelinePluginClassloaders: types.BoolValue(true),
		PipelineUnsafeShutdown:     types.BoolValue(false),
		PipelineWorkers:            types.Int64Value(4),
		QueueCheckpointAcks:        types.Int64Value(1024),
		QueueCheckpointRetry:       types.BoolValue(true),
		QueueCheckpointWrites:      types.Int64Value(1024),
		QueueDrain:                 types.BoolValue(false),
		QueueMaxBytes:              types.StringValue("1gb"),
		QueueMaxEvents:             types.Int64Value(0),
		QueuePageCapacity:          types.StringValue("64mb"),
		QueueType:                  types.StringValue("memory"),
	}

	settings := expandSettings(data)

	require.Equal(t, int64(50), settings["pipeline.batch.delay"])
	require.Equal(t, int64(125), settings["pipeline.batch.size"])
	require.Equal(t, "v1", settings["pipeline.ecs_compatibility"])
	require.Equal(t, "auto", settings["pipeline.ordered"])
	require.Equal(t, true, settings["pipeline.plugin_classloaders"])
	require.Equal(t, false, settings["pipeline.unsafe_shutdown"])
	require.Equal(t, int64(4), settings["pipeline.workers"])
	require.Equal(t, int64(1024), settings["queue.checkpoint.acks"])
	require.Equal(t, true, settings["queue.checkpoint.retry"])
	require.Equal(t, int64(1024), settings["queue.checkpoint.writes"])
	require.Equal(t, false, settings["queue.drain"])
	require.Equal(t, "1gb", settings["queue.max_bytes"])
	require.Equal(t, int64(0), settings["queue.max_events"])
	require.Equal(t, "64mb", settings["queue.page_capacity"])
	require.Equal(t, "memory", settings["queue.type"])
}

func TestExpandSettings_UnknownAndEmptyFieldsOmitted(t *testing.T) {
	t.Parallel()

	data := Data{
		PipelineBatchDelay:       types.Int64Unknown(),
		PipelineEcsCompatibility: types.StringValue(""),
	}

	settings := expandSettings(data)

	require.NotContains(t, settings, "pipeline.batch.delay")
	require.NotContains(t, settings, "pipeline.ecs_compatibility")
}
