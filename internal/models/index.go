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

import "time"

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
	AllowAutoCreate                 *bool               `json:"allow_auto_create,omitempty"`
}

type DataStreamSettings struct {
	Hidden             *bool `json:"hidden,omitempty"`
	AllowCustomRouting *bool `json:"allow_custom_routing,omitempty"`
}

type DataStreamOptions struct {
	FailureStore *FailureStoreOptions `json:"failure_store,omitempty"`
}

type FailureStoreOptions struct {
	Enabled   *bool                  `json:"enabled,omitempty"`
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
	Version  *int64         `json:"version,omitempty"`
}

type ComponentTemplateResponse struct {
	Name              string            `json:"name"`
	ComponentTemplate ComponentTemplate `json:"component_template"`
}

// IndexTemplatesResponse mirrors the GET /_index_template/<name> body so the read
// path can decode index template settings as raw map[string]any rather than through
// the typed go-elasticsearch structs, which silently drop fields they do not model
// (e.g. index.search.slowlog.include) and coerce string-encoded values such as
// index.lifecycle.parse_origination_date. See issue #3124.
type IndexTemplatesResponse struct {
	IndexTemplates []IndexTemplateResponse `json:"index_templates"`
}

type IndexTemplateResponse struct {
	Name          string        `json:"name"`
	IndexTemplate IndexTemplate `json:"index_template"`
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

type DataStreamLifecycleResponse struct {
	DataStreams []DataStreamLifecycle `json:"data_streams"`
}

type DataStreamLifecycle struct {
	Name      string            `json:"name"`
	Lifecycle LifecycleSettings `json:"lifecycle"`
}

type Downsampling struct {
	After         string `json:"after,omitempty"`
	FixedInterval string `json:"fixed_interval,omitempty"`
}
