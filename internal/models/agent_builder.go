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

type Agent struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Description   *string            `json:"description,omitempty"`
	AvatarColor   *string            `json:"avatar_color,omitempty"`
	AvatarSymbol  *string            `json:"avatar_symbol,omitempty"`
	Labels        []string           `json:"labels,omitempty"`
	Configuration AgentConfiguration `json:"configuration"`
}

type AgentConfiguration struct {
	Instructions *string            `json:"instructions,omitempty"`
	Tools        []AgentToolsConfig `json:"tools,omitempty"`
	SkillIDs     []string           `json:"skill_ids,omitempty"`
}

type AgentToolsConfig struct {
	ToolIDs []string `json:"tool_ids"`
}

type Tool struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Description   *string        `json:"description,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	ReadOnly      bool           `json:"readonly"`
	Configuration map[string]any `json:"configuration"`
}

type Workflow struct {
	ID          string  `json:"id"`
	Yaml        string  `json:"yaml"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Enabled     bool    `json:"enabled"`
	Valid       bool    `json:"valid"`
}

type Skill struct {
	ID                string                   `json:"id"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description"`
	Content           string                   `json:"content"`
	ToolIDs           []string                 `json:"tool_ids,omitempty"`
	ReferencedContent []SkillReferencedContent `json:"referenced_content,omitempty"`
}

type SkillReferencedContent struct {
	Name         string `json:"name"`
	RelativePath string `json:"relativePath"`
	Content      string `json:"content"`
}
