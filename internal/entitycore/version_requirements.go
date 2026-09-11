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

package entitycore

import (
	"context"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// VersionCheck reports whether a server version satisfies a requirement.
type VersionCheck func(*version.Version) bool

// VersionRequirement describes a server version that an entity model requires
// before the envelope invokes the concrete lifecycle callback.
type VersionRequirement struct {
	// MinVersion is the minimum server version required. It is ignored when
	// VersionCheck is set.
	MinVersion version.Version
	// VersionCheck supports requirements that cannot be expressed as a single
	// minimum version, such as features available in 8.19.x and 9.1.0+ but not
	// 9.0.x.
	VersionCheck VersionCheck
	// AttributePath scopes an unsupported-version diagnostic to the configured
	// Terraform attribute. A nil path produces a resource-level diagnostic.
	AttributePath *path.Path
	// ErrorMessage is the human-readable diagnostic detail.
	ErrorMessage string
}

// NewAttributeVersionRequirement creates an attribute-scoped minimum-version
// requirement.
func NewAttributeVersionRequirement(attr path.Path, minVersion version.Version, errorMessage string) VersionRequirement {
	return VersionRequirement{
		MinVersion:    minVersion,
		AttributePath: &attr,
		ErrorMessage:  errorMessage,
	}
}

// NewAttributeVersionCheckRequirement creates an attribute-scoped requirement
// evaluated by check.
func NewAttributeVersionCheckRequirement(attr path.Path, check VersionCheck, errorMessage string) VersionRequirement {
	return VersionRequirement{
		VersionCheck:  check,
		AttributePath: &attr,
		ErrorMessage:  errorMessage,
	}
}

// WithVersionRequirements is an optional interface that entity models may
// implement to declare server version requirements. When a decoded model
// satisfies this interface, Kibana and Elasticsearch envelopes evaluate the
// requirements after scoped client resolution and before invoking the concrete
// lifecycle callback.
type WithVersionRequirements interface {
	GetVersionRequirements(ctx context.Context) ([]VersionRequirement, diag.Diagnostics)
}

// MinVersionClient is implemented by scoped API clients used by entity envelopes
// for minimum server version checks.
type MinVersionClient interface {
	EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, diag.Diagnostics)
}

type versionCheckClient interface {
	EnforceVersionCheck(ctx context.Context, check func(*version.Version) bool) (bool, diag.Diagnostics)
}

// EnforceVersionRequirements checks whether model implements
// [WithVersionRequirements] and, if so, evaluates each requirement against the
// scoped client. It returns any diagnostics produced. Entity envelopes call
// this automatically; concrete resources whose Create/Update bypass the
// envelope can invoke it directly to honor the model's declared requirements.
func EnforceVersionRequirements(ctx context.Context, client MinVersionClient, model any) diag.Diagnostics {
	var diags diag.Diagnostics
	versionModel, ok := model.(WithVersionRequirements)
	if !ok {
		return diags
	}

	reqs, vDiags := versionModel.GetVersionRequirements(ctx)
	diags.Append(vDiags...)
	if diags.HasError() {
		return diags
	}

	for _, vReq := range reqs {
		var supported bool
		var vDiags diag.Diagnostics
		if vReq.VersionCheck != nil {
			checkClient, ok := client.(versionCheckClient)
			if !ok {
				diags.AddError(
					"Unable to enforce server version requirement",
					"The configured client does not support custom server version checks.",
				)
				return diags
			}
			supported, vDiags = checkClient.EnforceVersionCheck(ctx, vReq.VersionCheck)
		} else {
			supported, vDiags = client.EnforceMinVersion(ctx, &vReq.MinVersion)
		}
		diags.Append(vDiags...)
		if diags.HasError() {
			return diags
		}
		if !supported {
			if vReq.AttributePath != nil {
				diags.AddAttributeError(*vReq.AttributePath, "Unsupported server version", vReq.ErrorMessage)
			} else {
				diags.AddError("Unsupported server version", vReq.ErrorMessage)
			}
			return diags
		}
	}

	return diags
}
