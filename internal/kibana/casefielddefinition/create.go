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

package casefielddefinition

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var createFieldDefinition = entitycore.SimpleKibanaCreate[fieldDefinitionModel, kbapi.CasesFieldDefinitionWriteRequest, kbapi.CasesFieldDefinitionResponse](
	func(plan fieldDefinitionModel, _ context.Context) (kbapi.CasesFieldDefinitionWriteRequest, diag.Diagnostics) {
		return plan.toAPICreateModel(), nil
	},
	kibanaoapi.CreateCaseFieldDefinition,
	(*fieldDefinitionModel).populateWritten,
)

// populateWritten records only the server-assigned identifiers needed for the
// envelope's follow-up read; final state is derived from that read.
func (m *fieldDefinitionModel) populateWritten(_ context.Context, spaceID string, written *kbapi.CasesFieldDefinitionResponse) diag.Diagnostics {
	m.SpaceID = types.StringValue(spaceID)
	if written != nil {
		m.FieldDefinitionID = types.StringValue(written.FieldDefinitionId)
	}
	return nil
}
