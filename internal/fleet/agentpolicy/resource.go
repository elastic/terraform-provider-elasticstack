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

package agentpolicy

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newAgentPolicyResource()
	_ resource.ResourceWithConfigure   = newAgentPolicyResource()
	_ resource.ResourceWithImportState = newAgentPolicyResource()
	_ resource.ResourceWithModifyPlan  = newAgentPolicyResource()
)

var (
	MinVersionGlobalDataTags      = version.Must(version.NewVersion("8.15.0"))
	MinSupportsAgentlessVersion   = version.Must(version.NewVersion("8.15.0"))
	MinVersionInactivityTimeout   = version.Must(version.NewVersion("8.7.0"))
	MinVersionUnenrollmentTimeout = version.Must(version.NewVersion("8.15.0"))
	MinVersionSpaceIDs            = version.Must(version.NewVersion("9.1.0"))
	MinVersionRequiredVersions    = version.Must(version.NewVersion("9.1.0"))
	MinVersionAgentFeatures       = version.Must(version.NewVersion("8.7.0"))
	MinVersionAdvancedMonitoring  = version.Must(version.NewVersion("8.16.0"))
	MinVersionAdvancedSettings    = version.Must(version.NewVersion("8.17.0"))
	// MinVersionTamperProtection is the minimum stack version for setting agent policy tamper protection (is_protected).
	MinVersionTamperProtection = version.Must(version.NewVersion("8.10.0"))
)

// MonitoringRuntimeExperimentalSupported returns true if the given version
// supports the monitoring_runtime_experimental advanced setting: 8.19.x or 9.1.0+.
func MonitoringRuntimeExperimentalSupported(v *version.Version) bool {
	if v == nil {
		return false
	}
	segments := v.Segments()
	if len(segments) < 2 {
		return false
	}
	major, minor := segments[0], segments[1]
	switch {
	case major >= 10:
		return true
	case major == 9 && minor >= 1:
		return true
	case major == 8 && minor >= 19:
		return true
	default:
		return false
	}
}

type agentPolicyResource struct {
	*entitycore.ResourceBase
	*fleet.SpaceImporter
}

func newAgentPolicyResource() *agentPolicyResource {
	return &agentPolicyResource{
		ResourceBase:  entitycore.NewResourceBase(entitycore.ComponentFleet, "agent_policy"),
		SpaceImporter: fleet.NewSpaceImporter(path.Root("policy_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newAgentPolicyResource()
}
