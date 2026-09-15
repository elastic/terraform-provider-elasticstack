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

package panelkit

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Branch keys for the by_field/by_esql discriminated union shared by panels built from a
// by_field/by_esql attribute map (e.g. optionslist, rangeslider). Exported so panel packages and
// callers outside them (e.g. the dashboard resource's v0->v1 state upgrader) can reference a single
// source of truth instead of each declaring their own copy.
const (
	BranchByField = "by_field"
	BranchByEsql  = "by_esql"
)

// ExactlyOneOfBranchValidator enforces that exactly one of the by_field / by_esql branch
// attributes is configured inside a block built from a by_field/by_esql attribute map (e.g.
// optionslist.NestedAttributes, rangeslider.NestedAttributes). configName is the panel's config
// block name (e.g. "options_list_control_config") used in the diagnostic messages, and
// byFieldName/byEsqlName are the two mutually exclusive branch attribute names. Shared by each
// panel's regular schema and the pinned-panel control-bar schema.
func ExactlyOneOfBranchValidator(configName, byFieldName, byEsqlName string) validator.Object {
	return validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{byFieldName, byEsqlName},
		Summary:       fmt.Sprintf("Invalid %s", configName),
		MissingDetail: fmt.Sprintf("Exactly one of `%s` or `%s` must be configured inside `%s`.", byFieldName, byEsqlName, configName),
		TooManyDetail: fmt.Sprintf("Exactly one of `%s` or `%s` must be configured inside `%s`, not both.", byFieldName, byEsqlName, configName),
		Description:   fmt.Sprintf("Ensures exactly one of `%s` or `%s` is configured inside `%s`.", byFieldName, byEsqlName, configName),
	})
}
