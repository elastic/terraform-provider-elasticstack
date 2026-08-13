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

// Package fleetschema holds attribute builders shared between Fleet
// integration package installation resources (integration, customintegration)
// to keep their descriptions and semantics from drifting apart.
package fleetschema

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
)

func boolAttribute(description string, computed bool, def *bool) schema.BoolAttribute {
	attr := schema.BoolAttribute{
		Description: description,
		Optional:    true,
		Computed:    computed,
	}
	if def != nil {
		attr.Default = booldefault.StaticBool(*def)
	}
	return attr
}

// IgnoreMappingUpdateErrorsAttribute returns the canonical
// ignore_mapping_update_errors attribute for Fleet integration package
// installation resources.
func IgnoreMappingUpdateErrorsAttribute(computed bool, def *bool) schema.BoolAttribute {
	return boolAttribute("Set to true to ignore mapping update errors during package installation.", computed, def)
}

// SkipDataStreamRolloverAttribute returns the canonical
// skip_data_stream_rollover attribute for Fleet integration package
// installation resources.
func SkipDataStreamRolloverAttribute(computed bool, def *bool) schema.BoolAttribute {
	return boolAttribute("Set to true to skip data stream rollover during package installation.", computed, def)
}

// SkipDestroyAttribute returns the canonical skip_destroy attribute for Fleet
// integration package installation resources.
func SkipDestroyAttribute(computed bool, def *bool) schema.BoolAttribute {
	description := "Set to true if you do not wish the integration package to be uninstalled at destroy " +
		"time, and instead just remove the integration package from the Terraform state."
	return boolAttribute(description, computed, def)
}
