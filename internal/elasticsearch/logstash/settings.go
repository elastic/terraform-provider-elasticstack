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
	"math"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// queueMaxBytesRegexp validates the queue.max_bytes setting value format.
var queueMaxBytesRegexp = regexp.MustCompile(`^[0-9]+[kmgtp]?b$`)

// expandSettings converts the typed Data model fields into the flat
// map[string]any API settings format used by the Logstash Pipelines API.
// Only non-null, non-unknown fields are included.
func expandSettings(data Data) map[string]any {
	settings := make(map[string]any)

	if typeutils.IsKnown(data.PipelineBatchDelay) {
		settings["pipeline.batch.delay"] = data.PipelineBatchDelay.ValueInt64()
	}
	if typeutils.IsKnown(data.PipelineBatchSize) {
		settings["pipeline.batch.size"] = data.PipelineBatchSize.ValueInt64()
	}
	if typeutils.IsKnown(data.PipelineEcsCompatibility) && data.PipelineEcsCompatibility.ValueString() != "" {
		settings["pipeline.ecs_compatibility"] = data.PipelineEcsCompatibility.ValueString()
	}
	if typeutils.IsKnown(data.PipelineOrdered) && data.PipelineOrdered.ValueString() != "" {
		settings["pipeline.ordered"] = data.PipelineOrdered.ValueString()
	}
	if typeutils.IsKnown(data.PipelinePluginClassloaders) {
		settings["pipeline.plugin_classloaders"] = data.PipelinePluginClassloaders.ValueBool()
	}
	if typeutils.IsKnown(data.PipelineUnsafeShutdown) {
		settings["pipeline.unsafe_shutdown"] = data.PipelineUnsafeShutdown.ValueBool()
	}
	if typeutils.IsKnown(data.PipelineWorkers) {
		settings["pipeline.workers"] = data.PipelineWorkers.ValueInt64()
	}
	if typeutils.IsKnown(data.QueueCheckpointAcks) {
		settings["queue.checkpoint.acks"] = data.QueueCheckpointAcks.ValueInt64()
	}
	if typeutils.IsKnown(data.QueueCheckpointRetry) {
		settings["queue.checkpoint.retry"] = data.QueueCheckpointRetry.ValueBool()
	}
	if typeutils.IsKnown(data.QueueCheckpointWrites) {
		settings["queue.checkpoint.writes"] = data.QueueCheckpointWrites.ValueInt64()
	}
	if typeutils.IsKnown(data.QueueDrain) {
		settings["queue.drain"] = data.QueueDrain.ValueBool()
	}
	if typeutils.IsKnown(data.QueueMaxBytes) && data.QueueMaxBytes.ValueString() != "" {
		settings["queue.max_bytes"] = data.QueueMaxBytes.ValueString()
	}
	if typeutils.IsKnown(data.QueueMaxEvents) {
		settings["queue.max_events"] = data.QueueMaxEvents.ValueInt64()
	}
	if typeutils.IsKnown(data.QueuePageCapacity) && data.QueuePageCapacity.ValueString() != "" {
		settings["queue.page_capacity"] = data.QueuePageCapacity.ValueString()
	}
	if typeutils.IsKnown(data.QueueType) && data.QueueType.ValueString() != "" {
		settings["queue.type"] = data.QueueType.ValueString()
	}

	return settings
}

// flattenSettings reads the flat map[string]any API settings response and
// populates the corresponding typed fields on *Data.
func flattenSettings(apiSettings map[string]any, data *Data) {
	if v, ok := apiSettings["pipeline.batch.delay"]; ok {
		data.PipelineBatchDelay = types.Int64Value(toInt64(v))
	}
	if v, ok := apiSettings["pipeline.batch.size"]; ok {
		data.PipelineBatchSize = types.Int64Value(toInt64(v))
	}
	if v, ok := apiSettings["pipeline.ecs_compatibility"]; ok {
		data.PipelineEcsCompatibility = types.StringValue(toString(v))
	}
	if v, ok := apiSettings["pipeline.ordered"]; ok {
		data.PipelineOrdered = types.StringValue(toString(v))
	}
	if v, ok := apiSettings["pipeline.plugin_classloaders"]; ok {
		data.PipelinePluginClassloaders = types.BoolValue(toBool(v))
	}
	if v, ok := apiSettings["pipeline.unsafe_shutdown"]; ok {
		data.PipelineUnsafeShutdown = types.BoolValue(toBool(v))
	}
	if v, ok := apiSettings["pipeline.workers"]; ok {
		data.PipelineWorkers = types.Int64Value(toInt64(v))
	}
	if v, ok := apiSettings["queue.checkpoint.acks"]; ok {
		data.QueueCheckpointAcks = types.Int64Value(toInt64(v))
	}
	if v, ok := apiSettings["queue.checkpoint.retry"]; ok {
		data.QueueCheckpointRetry = types.BoolValue(toBool(v))
	}
	if v, ok := apiSettings["queue.checkpoint.writes"]; ok {
		data.QueueCheckpointWrites = types.Int64Value(toInt64(v))
	}
	if v, ok := apiSettings["queue.drain"]; ok {
		data.QueueDrain = types.BoolValue(toBool(v))
	}
	if v, ok := apiSettings["queue.max_bytes"]; ok {
		data.QueueMaxBytes = types.StringValue(toString(v))
	}
	if v, ok := apiSettings["queue.max_events"]; ok {
		data.QueueMaxEvents = types.Int64Value(toInt64(v))
	}
	if v, ok := apiSettings["queue.page_capacity"]; ok {
		data.QueuePageCapacity = types.StringValue(toString(v))
	}
	if v, ok := apiSettings["queue.type"]; ok {
		data.QueueType = types.StringValue(toString(v))
	}
}

// toInt64 converts a value from the Logstash API (typically float64 from JSON
// decode) to int64.
func toInt64(v any) int64 {
	switch val := v.(type) {
	case float64:
		return int64(math.Round(val))
	case int64:
		return val
	case int:
		return int64(val)
	}
	return 0
}

// toString converts a value from the API response to string.
func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// toBool converts a value from the API response to bool.
func toBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
