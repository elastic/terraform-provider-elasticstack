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

package lenscommon

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBlockModel struct {
	Value string
}

func TestSnapshotAndResetBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil field yields nil prior and a fresh zero-value block", func(t *testing.T) {
		t.Parallel()
		var field *testBlockModel
		prior := SnapshotAndResetBlock(&field)
		require.Nil(t, prior)
		require.NotNil(t, field)
		require.Equal(t, testBlockModel{}, *field)
	})

	t.Run("set field is snapshotted then reset, independent copies", func(t *testing.T) {
		t.Parallel()
		field := &testBlockModel{Value: "old"}
		prior := SnapshotAndResetBlock(&field)
		require.NotNil(t, prior)
		assert.Equal(t, "old", prior.Value)
		require.NotNil(t, field)
		assert.Equal(t, testBlockModel{}, *field)

		// Mutating the reset block must not affect the snapshot.
		field.Value = "new"
		assert.Equal(t, "old", prior.Value)
	})
}

func TestPopulateFromNoESQLOrESQL(t *testing.T) {
	t.Parallel()

	t.Run("guard nil takes the NoESQL branch whenever decode succeeds", func(t *testing.T) {
		t.Parallel()
		config := &testBlockModel{}
		diags := PopulateFromNoESQLOrESQL(
			context.Background(), config, nil,
			func() (string, error) { return "noesql", nil },
			func() (int, error) { return 0, errors.New("should not be called") },
			nil,
			func(_ context.Context, m *testBlockModel, _ *testBlockModel, v string) diag.Diagnostics {
				m.Value = v
				return nil
			},
			func(_ context.Context, _ *testBlockModel, _ *testBlockModel, _ int) diag.Diagnostics {
				t.Fatal("ESQL branch should not run")
				return nil
			},
		)
		require.False(t, diags.HasError())
		assert.Equal(t, "noesql", config.Value)
	})

	t.Run("guard rejecting the NoESQL candidate falls back to ESQL", func(t *testing.T) {
		t.Parallel()
		config := &testBlockModel{}
		diags := PopulateFromNoESQLOrESQL(
			context.Background(), config, nil,
			func() (string, error) { return "actually_esql", nil },
			func() (int, error) { return 42, nil },
			func(v string) bool { return v != "actually_esql" },
			func(_ context.Context, _ *testBlockModel, _ *testBlockModel, _ string) diag.Diagnostics {
				t.Fatal("NoESQL branch should not run when guard rejects")
				return nil
			},
			func(_ context.Context, m *testBlockModel, _ *testBlockModel, _ int) diag.Diagnostics {
				m.Value = "esql"
				return nil
			},
		)
		require.False(t, diags.HasError())
		assert.Equal(t, "esql", config.Value)
	})

	t.Run("NoESQL decode error falls back to ESQL", func(t *testing.T) {
		t.Parallel()
		config := &testBlockModel{}
		diags := PopulateFromNoESQLOrESQL(
			context.Background(), config, nil,
			func() (string, error) { return "", errors.New("not this variant") },
			func() (int, error) { return 7, nil },
			func(_ string) bool { return true },
			func(_ context.Context, _ *testBlockModel, _ *testBlockModel, _ string) diag.Diagnostics {
				t.Fatal("NoESQL branch should not run on decode error")
				return nil
			},
			func(_ context.Context, m *testBlockModel, _ *testBlockModel, _ int) diag.Diagnostics {
				m.Value = "esql-fallback"
				return nil
			},
		)
		require.False(t, diags.HasError())
		assert.Equal(t, "esql-fallback", config.Value)
	})

	t.Run("ESQL decode error produces a diagnostic error", func(t *testing.T) {
		t.Parallel()
		config := &testBlockModel{}
		diags := PopulateFromNoESQLOrESQL(
			context.Background(), config, nil,
			func() (string, error) { return "", errors.New("not NoESQL") },
			func() (int, error) { return 0, errors.New("not ESQL either") },
			func(_ string) bool { return true },
			func(_ context.Context, _ *testBlockModel, _ *testBlockModel, _ string) diag.Diagnostics {
				t.Fatal("NoESQL branch should not run")
				return nil
			},
			func(_ context.Context, _ *testBlockModel, _ *testBlockModel, _ int) diag.Diagnostics {
				t.Fatal("ESQL branch should not run when its own decode fails")
				return nil
			},
		)
		require.True(t, diags.HasError())
	})

	t.Run("prior is threaded through to the winning FromAPI helper", func(t *testing.T) {
		t.Parallel()
		config := &testBlockModel{}
		prior := &testBlockModel{Value: "previous"}
		diags := PopulateFromNoESQLOrESQL(
			context.Background(), config, prior,
			func() (string, error) { return "noesql", nil },
			func() (int, error) { return 0, errors.New("unused") },
			nil,
			func(_ context.Context, m *testBlockModel, p *testBlockModel, v string) diag.Diagnostics {
				require.NotNil(t, p)
				m.Value = p.Value + "+" + v
				return nil
			},
			func(_ context.Context, _ *testBlockModel, _ *testBlockModel, _ int) diag.Diagnostics {
				t.Fatal("ESQL branch should not run")
				return nil
			},
		)
		require.False(t, diags.HasError())
		assert.Equal(t, "previous+noesql", config.Value)
	})
}
