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

package security_role

// Terraform schema attribute keys shared across the security role resource
// schema, data source schema, attr types, and flatten helpers.
const (
	attrGrant                  = "grant"
	attrExcept                 = "except"
	attrNames                  = "names"
	attrPrivileges             = "privileges"
	attrQuery                  = "query"
	attrFieldSecurity          = "field_security"
	attrClusters               = "clusters"
	attrCluster                = "cluster"
	attrIndices                = "indices"
	attrRemoteIndices          = "remote_indices"
	attrRunAs                  = "run_as"
	attrName                   = "name"
	attrSpaces                 = "spaces"
	attrBase                   = "base"
	attrFeature                = "feature"
	attrAllowRestrictedIndices = "allow_restricted_indices"
)

// Schema descriptions repeated in resource and data source definitions.
const (
	descFieldSecurityGrant  = "List of the fields to grant the access to."
	descFieldSecurityExcept = "List of the fields to which the grants will not be applied."
	descFieldSecurityBlock  = "The document fields that the owners of the role have read access to."
	descIndexNames          = "A list of indices (or index name patterns) to which the permissions in this entry apply."
	descIndexPrivileges     = "The index level privileges that the owners of the role have on the specified indices."
	descIndexQuery          = "A search query that defines the documents the owners of the role have read access to."
)
