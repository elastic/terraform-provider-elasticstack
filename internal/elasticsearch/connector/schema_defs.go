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

package connector

// LeafAttr describes a single leaf schema attribute shared between the resource and data source.
// IsString true means a string attribute; false means a bool attribute.
type LeafAttr struct {
	Name        string
	Description string
	IsString    bool
}

// Base description constants for nested attributes shared between the resource and data source.
// Resource builders may append operation-specific context (e.g. "Changes trigger ...").
const (
	PipelineNestedDesc   = "Ingest pipeline settings applied to synced documents."
	SchedulingNestedDesc = "Sync scheduling for full, incremental, and access-control jobs."
	FeaturesNestedDesc   = "Connector feature flags."
	SyncRulesNestedDesc  = "Sync rules feature flags."
)

// PipelineLeafAttrs returns the shared leaf attribute definitions for the pipeline nested attribute.
func PipelineLeafAttrs() []LeafAttr {
	return []LeafAttr{
		{Name: NameAttr, Description: "Ingest pipeline name.", IsString: true},
		{Name: ExtractBinaryContentAttr, Description: "Whether to extract binary content during ingestion."},
		{Name: ReduceWhitespaceAttr, Description: "Whether to reduce whitespace in extracted text."},
		{Name: RunMlInferenceAttr, Description: "Whether to run ML inference during ingestion."},
	}
}

// ScheduleEntryLeafAttrs returns the shared leaf attribute definitions for a schedule entry.
func ScheduleEntryLeafAttrs() []LeafAttr {
	return []LeafAttr{
		{Name: EnabledAttr, Description: "Whether this scheduled job type is enabled."},
		{Name: IntervalAttr, Description: "Cron expression accepted by the Elasticsearch scheduler.", IsString: true},
	}
}

// FeatureFlagLeafAttrs returns the shared leaf attribute definitions for a feature flag.
func FeatureFlagLeafAttrs() []LeafAttr {
	return []LeafAttr{
		{Name: EnabledAttr, Description: "Whether the feature is enabled."},
	}
}
