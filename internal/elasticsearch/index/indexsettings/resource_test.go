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

package indexsettings

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestNewIndexSettingsResource(t *testing.T) {
	assert.Implements(t, (*resource.Resource)(nil), NewIndexSettingsResource())
	assert.Implements(
		t,
		(*resource.ResourceWithValidateConfig)(nil),
		NewIndexSettingsResource(),
		"the resource must run the at-least-one-setting and overlap validation at plan time",
	)
}

func TestNewIndexSettingsResource_ImplementsImportStateAndConfigure(t *testing.T) {
	assert.Implements(
		t,
		(*resource.ResourceWithImportState)(nil),
		NewIndexSettingsResource(),
		"the resource must implement the composite-ID import with the import-hydration private-state flag",
	)
	assert.Implements(
		t,
		(*resource.ResourceWithConfigure)(nil),
		NewIndexSettingsResource(),
		"the Elasticsearch envelope must be configured with the scoped client",
	)
}
