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

// Package casefielddefinition implements the elasticstack_kibana_case_field_definition resource.
package casefielddefinition

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newResource()
	_ resource.ResourceWithConfigure   = newResource()
	_ resource.ResourceWithImportState = newResource()
)

// Resource manages a single field in the Kibana Cases field library.
type Resource struct {
	*entitycore.KibanaResource[fieldDefinitionModel]
	*entitycore.KibanaSpaceImporter
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[fieldDefinitionModel](
			entitycore.ComponentKibana,
			"case_field_definition",
			entitycore.KibanaResourceOptions[fieldDefinitionModel]{
				Schema: getResourceSchema,
				Read:   entitycore.SimpleKibanaRead[fieldDefinitionModel, kbapi.CasesFieldDefinitionResponse](kibanaoapi.GetCaseFieldDefinition, (*fieldDefinitionModel).populateFromAPI),
				Delete: entitycore.SimpleKibanaDelete[fieldDefinitionModel](kibanaoapi.DeleteCaseFieldDefinition),
				Create: createFieldDefinition,
				Update: updateFieldDefinition,
			},
		),
		KibanaSpaceImporter: entitycore.NewKibanaSpaceImporter(
			path.Root("id"), path.Root("space_id"), path.Root("field_definition_id"),
		).RequireSpaceID(
			"Invalid import ID",
			"The import ID must be in the format \"<space_id>/<field_definition_id>\".",
		),
	}
}

// NewResource returns the resource factory used for provider registration.
func NewResource() resource.Resource {
	return newResource()
}
