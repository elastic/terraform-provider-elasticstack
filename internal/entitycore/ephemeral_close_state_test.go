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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type plainGoCloseState struct {
	Name     string
	Count    int64
	Enabled  bool
	Optional *bool
	Tags     []string
	Headers  map[string]string
	Nested   nestedPlainGoCloseState
	embeddedPlainGoCloseState
}

type nestedPlainGoCloseState struct {
	Field string
}

type embeddedPlainGoCloseState struct {
	EmbeddedField string
}

func TestMustBePlainGoCloseState_acceptsPlainGoTypes(t *testing.T) {
	t.Parallel()
	require.NotPanics(t, func() {
		mustBePlainGoCloseState[plainGoCloseState]()
	})
}

func TestMustBePlainGoCloseState_rejectsTfsdkField(t *testing.T) {
	t.Parallel()

	type badState struct {
		KeyID types.String
	}

	require.Panics(t, func() {
		mustBePlainGoCloseState[badState]()
	})
}

func TestMustBePlainGoCloseState_rejectsEmbeddedTfsdkField(t *testing.T) {
	t.Parallel()

	type inner struct {
		Field types.Bool
	}
	type badState struct {
		Inner inner
	}

	require.Panics(t, func() {
		mustBePlainGoCloseState[badState]()
	})
}

func TestMustBePlainGoCloseState_rejectsSliceElementTfsdk(t *testing.T) {
	t.Parallel()

	type badState struct {
		Items []types.List
	}

	require.Panics(t, func() {
		mustBePlainGoCloseState[badState]()
	})
}

func TestMustBePlainGoCloseState_rejectsMapValueTfsdk(t *testing.T) {
	t.Parallel()

	type badState struct {
		Values map[string]types.Object
	}

	require.Panics(t, func() {
		mustBePlainGoCloseState[badState]()
	})
}

func TestMustBePlainGoCloseState_rejectsPointerTfsdk(t *testing.T) {
	t.Parallel()

	type badState struct {
		Field *types.String
	}

	require.Panics(t, func() {
		mustBePlainGoCloseState[badState]()
	})
}

func TestMustBePlainGoCloseState_rejectsJsontypesNormalized(t *testing.T) {
	t.Parallel()

	type badState struct {
		Metadata jsontypes.Normalized
	}

	require.Panics(t, func() {
		mustBePlainGoCloseState[badState]()
	})
}

func TestEncodeDecodeUserCloseState_roundTrip(t *testing.T) {
	t.Parallel()

	falseVal := false
	original := plainGoCloseState{
		Name:     "test",
		Count:    42,
		Enabled:  true,
		Optional: &falseVal,
		Tags:     []string{"a", "b"},
		Headers:  map[string]string{"X-Foo": "bar"},
		Nested:   nestedPlainGoCloseState{Field: "nested"},
		embeddedPlainGoCloseState: embeddedPlainGoCloseState{
			EmbeddedField: "embedded",
		},
	}

	data, encodeDiags := encodeUserCloseState(original)
	require.False(t, encodeDiags.HasError())

	decoded, decodeDiags := decodeUserCloseState[plainGoCloseState](data)
	require.False(t, decodeDiags.HasError())
	require.Equal(t, original.Name, decoded.Name)
	require.Equal(t, original.Count, decoded.Count)
	require.Equal(t, original.Enabled, decoded.Enabled)
	require.Equal(t, *original.Optional, *decoded.Optional)
	require.Equal(t, original.Tags, decoded.Tags)
	require.Equal(t, original.Headers, decoded.Headers)
	require.Equal(t, original.Nested.Field, decoded.Nested.Field)
	require.Equal(t, original.EmbeddedField, decoded.EmbeddedField)
}
