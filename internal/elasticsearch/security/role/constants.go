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

package role

import "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"

// Aliases for role schema keys defined in the parent security package.
const (
	attrPrivileges             = security.RoleAttrPrivileges
	attrGrant                  = security.RoleAttrGrant
	attrExcept                 = security.RoleAttrExcept
	attrFieldSecurity          = security.RoleAttrFieldSecurity
	attrNames                  = security.RoleAttrNames
	attrQuery                  = security.RoleAttrQuery
	attrClusters               = security.RoleAttrClusters
	attrName                   = security.RoleAttrName
	attrDescription            = security.RoleAttrDescription
	attrGlobal                 = security.RoleAttrGlobal
	attrCluster                = security.RoleAttrCluster
	attrMetadata               = security.RoleAttrMetadata
	attrApplication            = security.RoleAttrApplication
	attrResources              = security.RoleAttrResources
	attrAllowRestrictedIndices = security.RoleAttrAllowRestrictedIndices

	blockApplications  = security.RoleBlockApplications
	blockIndices       = security.RoleBlockIndices
	blockRemoteIndices = security.RoleBlockRemoteIndices
)
