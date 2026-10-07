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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// importHydrationPrivateKey is the private-state key set by ImportState and
// cleared by the first read, signalling the one-shot hydration of the known
// dynamic typed attributes after terraform import. Hydration is keyed on this
// flag and never on an empty tracked-settings set, which is also reachable
// via a legitimate configuration after an external reset of the last tracked
// setting.
const importHydrationPrivateKey = "index_settings_import_hydration"

// setImportHydrationFlag records the import read request in private state so
// the first read hydrates the known dynamic typed attributes.
func setImportHydrationFlag(ctx context.Context, private entitycore.PrivateStateStorage) diag.Diagnostics {
	return private.SetKey(ctx, importHydrationPrivateKey, []byte(`true`))
}

// importHydrationRequested reports whether the import-hydration flag is set in
// private state.
func importHydrationRequested(ctx context.Context, private entitycore.PrivateStateStorage) bool {
	if private == nil {
		return false
	}

	value, diags := private.GetKey(ctx, importHydrationPrivateKey)
	if diags.HasError() {
		return false
	}

	return len(value) > 0
}

// clearImportHydrationFlag removes the import-hydration flag from private
// state so later reads do not hydrate again. Setting a key to a zero-length
// value removes it.
func clearImportHydrationFlag(ctx context.Context, private entitycore.PrivateStateStorage) diag.Diagnostics {
	if private == nil {
		return nil
	}

	return private.SetKey(ctx, importHydrationPrivateKey, nil)
}
