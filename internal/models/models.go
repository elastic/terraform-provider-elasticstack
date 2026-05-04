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
	"errors"
	"strings"
	"time"
)

type APIKeyRoleDescriptor struct {
	Name          string             `json:"-"`
	Applications  []Application      `json:"applications,omitempty"`
	Global        map[string]any     `json:"global,omitempty"`
	Cluster       []string           `json:"cluster,omitempty"`
	Indices       []IndexPerms       `json:"indices,omitempty"`
	RemoteIndices []RemoteIndexPerms `json:"remote_indices,omitempty"`
	Metadata      map[string]any     `json:"metadata,omitempty"`
	RunAs         []string           `json:"run_as,omitempty"`
	Restriction   *Restriction       `json:"restriction,omitempty"`
}

type Restriction struct {
	Workflows []string `json:"workflows,omitempty"`
}

type CrossClusterAPIKeyAccess struct {
	Search      []CrossClusterAPIKeyAccessEntry `json:"search,omitempty"`
	Replication []CrossClusterAPIKeyAccessEntry `json:"replication,omitempty"`
}

type CrossClusterAPIKeyAccessEntry struct {
	Names                  []string       `json:"names"`
	FieldSecurity          *FieldSecurity `json:"field_security,omitempty"`
	Query                  *string        `json:"query,omitempty"`
	AllowRestrictedIndices *bool          `json:"allow_restricted_indices,omitempty"`
}

type IndexPerms struct {
	FieldSecurity          *FieldSecurity `json:"field_security,omitempty"`
	Names                  []string       `json:"names"`
	Privileges             []string       `json:"privileges"`
	Query                  *string        `json:"query,omitempty"`
	AllowRestrictedIndices *bool          `json:"allow_restricted_indices,omitempty"`
}

type RemoteIndexPerms struct {
	IndexPerms
	Clusters []string `json:"clusters"`
}

type FieldSecurity struct {
	Grant  []string `json:"grant,omitempty"`
	Except []string `json:"except,omitempty"`
}

type Application struct {
	Name       string   `json:"application"`
	Privileges []string `json:"privileges,omitempty"`
	Resources  []string `json:"resources"`
}

type IndexTemplate struct {
	Name                            string              `json:"-"`
	Create                          bool                `json:"-"`
	Timeout                         string              `json:"-"`
	ComposedOf                      []string            `json:"composed_of"`
	IgnoreMissingComponentTemplates []string            `json:"ignore_missing_component_templates,omitempty"`
	DataStream                      *DataStreamSettings `json:"data_stream,omitempty"`
	IndexPatterns                   []string            `json:"index_patterns"`
	Meta                            map[string]any      `json:"_meta,omitempty"`
	Priority                        *int64              `json:"priority,omitempty"`
	Template                        *Template           `json:"template,omitempty"`
	Version                         *int64              `json:"version,omitempty"`
}

type DataStreamSettings struct {
	Hidden             *bool `json:"hidden,omitempty"`
	AllowCustomRouting *bool `json:"allow_custom_routing,omitempty"`
}

type DataStreamOptions struct {
	FailureStore *FailureStoreOptions `json:"failure_store,omitempty"`
}

type FailureStoreOptions struct {
	Enabled   bool                   `json:"enabled"`
	Lifecycle *FailureStoreLifecycle `json:"lifecycle,omitempty"`
}

type FailureStoreLifecycle struct {
	DataRetention string `json:"data_retention,omitempty"`
}

type Template struct {
	Aliases           map[string]IndexAlias `json:"aliases,omitempty"`
	Mappings          map[string]any        `json:"mappings,omitempty"`
	Settings          map[string]any        `json:"settings,omitempty"`
	Lifecycle         *LifecycleSettings    `json:"lifecycle,omitempty"`
	DataStreamOptions *DataStreamOptions    `json:"data_stream_options,omitempty"`
}

type ComponentTemplate struct {
	Name     string         `json:"-"`
	Meta     map[string]any `json:"_meta,omitempty"`
	Template *Template      `json:"template,omitempty"`
	Version  *int           `json:"version,omitempty"`
}

type ComponentTemplateResponse struct {
	Name              string            `json:"name"`
	ComponentTemplate ComponentTemplate `json:"component_template"`
}

type Policy struct {
	Name     string           `json:"-"`
	Metadata map[string]any   `json:"_meta,omitempty"`
	Phases   map[string]Phase `json:"phases"`
}

type Phase struct {
	MinAge  string            `json:"min_age,omitempty"`
	Actions map[string]Action `json:"actions"`
}

type Action map[string]any

type StringSliceOrCSV []string

var ErrInvalidStringSliceOrCSV = errors.New("expected array of strings, or a csv string")

func (i *StringSliceOrCSV) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// First try to parse as an array
	var sliceResult []string
	if err := json.Unmarshal(data, &sliceResult); err == nil {
		*i = StringSliceOrCSV(sliceResult)
		return nil
	}

	var stringResult string
	if err := json.Unmarshal(data, &stringResult); err == nil {
		*i = StringSliceOrCSV(strings.Split(stringResult, ","))
		return nil
	}

	return ErrInvalidStringSliceOrCSV
}

type Index struct {
	Name     string                `json:"-"`
	Aliases  map[string]IndexAlias `json:"aliases,omitempty"`
	Mappings map[string]any        `json:"mappings,omitempty"`
	Settings map[string]any        `json:"settings,omitempty"`
}

type PutIndexParams struct {
	WaitForActiveShards string
	MasterTimeout       time.Duration
	Timeout             time.Duration
}

type IndexAlias struct {
	Name          string         `json:"-"`
	Filter        map[string]any `json:"filter,omitempty"`
	IndexRouting  string         `json:"index_routing,omitempty"`
	IsHidden      bool           `json:"is_hidden,omitempty"`
	IsWriteIndex  bool           `json:"is_write_index,omitempty"`
	Routing       string         `json:"routing,omitempty"`
	SearchRouting string         `json:"search_routing,omitempty"`
}

type LifecycleSettings struct {
	DataRetention string         `json:"data_retention,omitempty"`
	Enabled       bool           `json:"enabled,omitempty"`
	Downsampling  []Downsampling `json:"downsampling,omitempty"`
}

type Downsampling struct {
	After         string `json:"after,omitempty"`
	FixedInterval string `json:"fixed_interval,omitempty"`
}

type DataStreamLifecycle struct {
	Name      string            `json:"name"`
	Lifecycle LifecycleSettings `json:"lifecycle,omitzero"`
}

type TimestampField struct {
	Name string `json:"name"`
}

type LogstashPipeline struct {
	PipelineID       string         `json:"-"`
	Description      string         `json:"description,omitempty"`
	LastModified     string         `json:"last_modified"`
	Pipeline         string         `json:"pipeline"`
	PipelineMetadata map[string]any `json:"pipeline_metadata"`
	PipelineSettings map[string]any `json:"pipeline_settings"`
	Username         string         `json:"username"`
}

type Watch struct {
	WatchID string `json:"-"`
	Status  struct {
		State struct {
			Active bool `json:"active"`
		} `json:"state"`
	} `json:"status"`
	Body WatchBody `json:"watch"`
}

type PutWatch struct {
	WatchID string
	Active  bool
	Body    WatchBody
}

type WatchBody struct {
	Trigger                map[string]any `json:"trigger"`
	Input                  map[string]any `json:"input"`
	Condition              map[string]any `json:"condition"`
	Actions                map[string]any `json:"actions"`
	Metadata               map[string]any `json:"metadata"`
	Transform              map[string]any `json:"transform,omitempty"`
	ThrottlePeriodInMillis int            `json:"throttle_period_in_millis,omitempty"`
}
