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
