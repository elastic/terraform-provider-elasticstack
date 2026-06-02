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
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// testActionModel is a minimal Elasticsearch action model: connection block,
// timeouts block (both via embedded fields), and a single user attribute.
type testActionModel struct {
	ElasticsearchConnectionField
	ActionTimeoutsField
	Value types.String `tfsdk:"value"`
}

// testKibanaActionModel mirrors testActionModel but for the Kibana variant.
type testKibanaActionModel struct {
	KibanaConnectionField
	ActionTimeoutsField
	Value types.String `tfsdk:"value"`
}

func testActionSchema(_ context.Context) actionschema.Schema {
	return actionschema.Schema{
		Attributes: map[string]actionschema.Attribute{
			"value": actionschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// configureESAction calls Configure on the typed action with the supplied
// factory and asserts no diagnostics.
func configureESAction(t *testing.T, a action.Action, factory *clients.ProviderClientFactory) {
	t.Helper()
	var resp action.ConfigureResponse
	a.(action.ActionWithConfigure).Configure(context.Background(), action.ConfigureRequest{ProviderData: factory}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "Configure must not error: %v", resp.Diagnostics)
}

// buildActionInvokeConfig constructs a tfsdk.Config with a `value` attribute,
// an `elasticsearch_connection` block (always null in tests), and a `timeouts`
// block configured to make `timeouts.invoke` null so the envelope falls back
// to the default invoke timeout. The `value` attribute is set to "hello" so
// callbacks can assert decoded config propagation.
func buildActionInvokeConfig(t *testing.T, schema actionschema.Schema) tfsdk.Config {
	t.Helper()
	const value = "hello"

	timeoutsObjType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"invoke": tftypes.String,
		},
	}

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"value":                    tftypes.String,
			"elasticsearch_connection": elasticsearchConnectionBlockType(),
			"timeouts":                 timeoutsObjType,
		},
	}

	return tfsdk.Config{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"value":                    tftypes.NewValue(tftypes.String, value),
			"elasticsearch_connection": tftypes.NewValue(elasticsearchConnectionBlockType(), nil),
			"timeouts": tftypes.NewValue(timeoutsObjType, map[string]tftypes.Value{
				"invoke": tftypes.NewValue(tftypes.String, nil),
			}),
		}),
		Schema: schema,
	}
}

// invokeSchema returns the schema as exposed via the envelope's Schema method,
// guaranteeing the timeouts and connection blocks match what tests build.
func invokeSchema(t *testing.T, a action.Action) actionschema.Schema {
	t.Helper()
	var resp action.SchemaResponse
	a.Schema(context.Background(), action.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "Schema must not error: %v", resp.Diagnostics)
	return resp.Schema
}

func TestNewElasticsearchAction_typeAssertions(t *testing.T) {
	t.Parallel()
	a := NewElasticsearchAction[testActionModel]("test_entity", ElasticsearchActionOptions[testActionModel]{
		Schema: testActionSchema,
		Invoke: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ ActionRequest[testActionModel]) diag.Diagnostics {
			return nil
		},
	})
	require.Implements(t, (*action.Action)(nil), a)
	require.Implements(t, (*action.ActionWithConfigure)(nil), a)
}

func TestNewKibanaAction_typeAssertions(t *testing.T) {
	t.Parallel()
	a := NewKibanaAction[testKibanaActionModel]("test_entity", KibanaActionOptions[testKibanaActionModel]{
		Schema: testActionSchema,
		Invoke: func(_ context.Context, _ *clients.KibanaScopedClient, _ ActionRequest[testKibanaActionModel]) diag.Diagnostics {
			return nil
		},
	})
	require.Implements(t, (*action.Action)(nil), a)
	require.Implements(t, (*action.ActionWithConfigure)(nil), a)
}

func TestNewElasticsearchAction_panicsOnNilSchema(t *testing.T) {
	t.Parallel()
	require.PanicsWithValue(t, "entitycore: ElasticsearchActionOptions.Schema must not be nil", func() {
		_ = NewElasticsearchAction[testActionModel]("x", ElasticsearchActionOptions[testActionModel]{
			Invoke: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ ActionRequest[testActionModel]) diag.Diagnostics {
				return nil
			},
		})
	})
}

func TestNewElasticsearchAction_panicsOnNilInvoke(t *testing.T) {
	t.Parallel()
	require.PanicsWithValue(t, "entitycore: ElasticsearchActionOptions.Invoke must not be nil", func() {
		_ = NewElasticsearchAction[testActionModel]("x", ElasticsearchActionOptions[testActionModel]{
			Schema: testActionSchema,
		})
	})
}

func TestNewKibanaAction_panicsOnNilSchema(t *testing.T) {
	t.Parallel()
	require.PanicsWithValue(t, "entitycore: KibanaActionOptions.Schema must not be nil", func() {
		_ = NewKibanaAction[testKibanaActionModel]("x", KibanaActionOptions[testKibanaActionModel]{
			Invoke: func(_ context.Context, _ *clients.KibanaScopedClient, _ ActionRequest[testKibanaActionModel]) diag.Diagnostics {
				return nil
			},
		})
	})
}

func TestNewKibanaAction_panicsOnNilInvoke(t *testing.T) {
	t.Parallel()
	require.PanicsWithValue(t, "entitycore: KibanaActionOptions.Invoke must not be nil", func() {
		_ = NewKibanaAction[testKibanaActionModel]("x", KibanaActionOptions[testKibanaActionModel]{
			Schema: testActionSchema,
		})
	})
}

func TestElasticsearchAction_SchemaInjectsBlocks(t *testing.T) {
	t.Parallel()
	a := NewElasticsearchAction[testActionModel]("test_entity", ElasticsearchActionOptions[testActionModel]{
		Schema: testActionSchema,
		Invoke: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ ActionRequest[testActionModel]) diag.Diagnostics {
			return nil
		},
	})

	schema := invokeSchema(t, a)

	require.Contains(t, schema.Blocks, "elasticsearch_connection", "envelope must inject elasticsearch_connection block")
	require.Contains(t, schema.Blocks, "timeouts", "envelope must inject timeouts block")
	require.Contains(t, schema.Attributes, "value", "concrete attributes must be preserved")
}

func TestKibanaAction_SchemaInjectsBlocks(t *testing.T) {
	t.Parallel()
	a := NewKibanaAction[testKibanaActionModel]("test_entity", KibanaActionOptions[testKibanaActionModel]{
		Schema: testActionSchema,
		Invoke: func(_ context.Context, _ *clients.KibanaScopedClient, _ ActionRequest[testKibanaActionModel]) diag.Diagnostics {
			return nil
		},
	})

	schema := invokeSchema(t, a)

	require.Contains(t, schema.Blocks, "kibana_connection", "envelope must inject kibana_connection block")
	require.Contains(t, schema.Blocks, "timeouts", "envelope must inject timeouts block")
	require.Contains(t, schema.Attributes, "value", "concrete attributes must be preserved")
}

func TestActionBase_MetadataUsesProviderTypeName(t *testing.T) {
	t.Parallel()
	base := NewActionBase(ComponentElasticsearch, "snapshot_create")
	var resp action.MetadataResponse
	base.Metadata(context.Background(), action.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
	require.Equal(t, "elasticstack_elasticsearch_snapshot_create", resp.TypeName)
}

func TestActionBase_ConfigureNilProviderDataNoError(t *testing.T) {
	t.Parallel()
	base := NewActionBase(ComponentElasticsearch, "x")
	var resp action.ConfigureResponse
	base.Configure(context.Background(), action.ConfigureRequest{ProviderData: nil}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	require.Nil(t, base.Client())
}

func TestActionBase_ConfigureValidFactoryStored(t *testing.T) {
	t.Parallel()
	base := NewActionBase(ComponentElasticsearch, "x")
	factory := nonNilTestFactory()
	var resp action.ConfigureResponse
	base.Configure(context.Background(), action.ConfigureRequest{ProviderData: factory}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	require.Same(t, factory, base.Client())
}

func TestElasticsearchAction_InvokeFactoryNilSurfacesError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	called := false
	a := NewElasticsearchAction[testActionModel]("test_entity", ElasticsearchActionOptions[testActionModel]{
		Schema: testActionSchema,
		Invoke: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ ActionRequest[testActionModel]) diag.Diagnostics {
			called = true
			return nil
		},
	})

	schema := invokeSchema(t, a)
	cfg := buildActionInvokeConfig(t, schema)

	var resp action.InvokeResponse
	a.Invoke(ctx, action.InvokeRequest{Config: cfg}, &resp)

	require.True(t, resp.Diagnostics.HasError(), "expected error when factory is nil")
	require.False(t, called, "callback must not be called when factory is nil")
}

func TestElasticsearchAction_InvokeCallbackReceivesDecodedConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	var seen testActionModel
	a := NewElasticsearchAction[testActionModel]("test_entity", ElasticsearchActionOptions[testActionModel]{
		Schema: testActionSchema,
		Invoke: func(_ context.Context, _ *clients.ElasticsearchScopedClient, req ActionRequest[testActionModel]) diag.Diagnostics {
			seen = req.Config
			return nil
		},
		DefaultInvokeTimeout: 5 * time.Minute,
	})
	configureESAction(t, a, factory)

	schema := invokeSchema(t, a)
	cfg := buildActionInvokeConfig(t, schema)

	var resp action.InvokeResponse
	a.Invoke(ctx, action.InvokeRequest{Config: cfg}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "Invoke must succeed: %v", resp.Diagnostics)
	require.Equal(t, "hello", seen.Value.ValueString(), "callback should receive decoded config value")
}

func TestElasticsearchAction_InvokePropagatesCallbackDiagnostics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	a := NewElasticsearchAction[testActionModel]("test_entity", ElasticsearchActionOptions[testActionModel]{
		Schema: testActionSchema,
		Invoke: func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ ActionRequest[testActionModel]) diag.Diagnostics {
			var d diag.Diagnostics
			d.AddError("callback failed", "the callback returned an error diagnostic")
			return d
		},
	})
	configureESAction(t, a, factory)

	schema := invokeSchema(t, a)
	cfg := buildActionInvokeConfig(t, schema)

	var resp action.InvokeResponse
	a.Invoke(ctx, action.InvokeRequest{Config: cfg}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "callback failed", resp.Diagnostics[0].Summary())
}

func TestElasticsearchAction_InvokeAppliesTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	var receivedDeadline time.Time
	a := NewElasticsearchAction[testActionModel]("test_entity", ElasticsearchActionOptions[testActionModel]{
		Schema: testActionSchema,
		Invoke: func(ctx context.Context, _ *clients.ElasticsearchScopedClient, _ ActionRequest[testActionModel]) diag.Diagnostics {
			deadline, ok := ctx.Deadline()
			require.True(t, ok, "envelope must apply context deadline")
			receivedDeadline = deadline
			return nil
		},
		DefaultInvokeTimeout: 7 * time.Second,
	})
	configureESAction(t, a, factory)

	schema := invokeSchema(t, a)
	cfg := buildActionInvokeConfig(t, schema)

	before := time.Now()
	var resp action.InvokeResponse
	a.Invoke(ctx, action.InvokeRequest{Config: cfg}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	expected := before.Add(7 * time.Second)
	require.WithinDuration(t, expected, receivedDeadline, 2*time.Second,
		"deadline should match the configured default timeout")
}

func TestElasticsearchAction_InvokeFallsBackToPackageDefaultTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestConfiguredFactory(ctx, t)

	var receivedDeadline time.Time
	a := NewElasticsearchAction[testActionModel]("test_entity", ElasticsearchActionOptions[testActionModel]{
		Schema: testActionSchema,
		Invoke: func(ctx context.Context, _ *clients.ElasticsearchScopedClient, _ ActionRequest[testActionModel]) diag.Diagnostics {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			receivedDeadline = deadline
			return nil
		},
	})
	configureESAction(t, a, factory)

	schema := invokeSchema(t, a)
	cfg := buildActionInvokeConfig(t, schema)

	before := time.Now()
	var resp action.InvokeResponse
	a.Invoke(ctx, action.InvokeRequest{Config: cfg}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	expected := before.Add(DefaultActionInvokeTimeout)
	require.WithinDuration(t, expected, receivedDeadline, 5*time.Second,
		"deadline should fall back to DefaultActionInvokeTimeout")
}
