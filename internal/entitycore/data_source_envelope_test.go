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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// testModel embeds KibanaConnectionField for envelope tests.
type testModel struct {
	KibanaConnectionField
	ID types.String `tfsdk:"id"`
}

func getTestSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func testReadFunc(_ context.Context, _ *clients.KibanaScopedClient, model testModel) (testModel, diag.Diagnostics) {
	model.ID = types.StringValue("result")
	return model, nil
}

func kibanaConnectionBlockType() tftypes.Type {
	nestedObjType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key":      tftypes.String,
			"bearer_token": tftypes.String,
			"username":     tftypes.String,
			"password":     tftypes.String,
			"endpoints":    tftypes.List{ElementType: tftypes.String},
			"ca_certs":     tftypes.List{ElementType: tftypes.String},
			"insecure":     tftypes.Bool,
		},
	}
	return tftypes.List{ElementType: nestedObjType}
}

func TestNewKibanaDataSource_typeAssertions(t *testing.T) {
	t.Parallel()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)
	require.NotNil(t, ds)
	require.Implements(t, (*datasource.DataSource)(nil), ds)
	require.Implements(t, (*datasource.DataSourceWithConfigure)(nil), ds)
}

func TestNewElasticsearchDataSource_typeAssertions(t *testing.T) {
	t.Parallel()
	ds := NewElasticsearchDataSource[struct {
		ElasticsearchConnectionField
	}](ComponentElasticsearch, "test_entity", func(_ context.Context) dsschema.Schema {
		return dsschema.Schema{}
	}, func(_ context.Context, _ *clients.ElasticsearchScopedClient, model struct {
		ElasticsearchConnectionField
	}) (struct {
		ElasticsearchConnectionField
	}, diag.Diagnostics) {
		return model, nil
	})
	require.NotNil(t, ds)
	require.Implements(t, (*datasource.DataSource)(nil), ds)
	require.Implements(t, (*datasource.DataSourceWithConfigure)(nil), ds)
}

func TestNewKibanaDataSource_schemaInjection(t *testing.T) {
	t.Parallel()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "kibana_connection")
	require.Contains(t, resp.Schema.Attributes, "id")
}

func TestNewElasticsearchDataSource_schemaInjection(t *testing.T) {
	t.Parallel()
	ds := NewElasticsearchDataSource[struct {
		ElasticsearchConnectionField
	}](ComponentElasticsearch, "test_entity", func(_ context.Context) dsschema.Schema {
		return dsschema.Schema{}
	}, func(_ context.Context, _ *clients.ElasticsearchScopedClient, model struct {
		ElasticsearchConnectionField
	}) (struct {
		ElasticsearchConnectionField
	}, diag.Diagnostics) {
		return model, nil
	})

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "elasticsearch_connection")
}

func TestNewKibanaDataSource_schemaDefensiveClone(t *testing.T) {
	t.Parallel()
	originalSchema := getTestSchema(context.Background())
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", func(_ context.Context) dsschema.Schema {
		return originalSchema
	}, testReadFunc)

	// Call Schema twice; the second call should still inject a fresh block.
	var resp1 datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp1)
	require.False(t, resp1.Diagnostics.HasError())

	var resp2 datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp2)
	require.False(t, resp2.Diagnostics.HasError())

	// The original schema's Blocks map should remain nil because the factory
	// is called once per Schema invocation and the constructor clones the map.
	require.Nil(t, originalSchema.Blocks)
}

func TestNewKibanaDataSource_Configure(t *testing.T) {
	ctx := context.Background()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	t.Run("nil_provider_data", func(t *testing.T) {
		t.Parallel()
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
			ProviderData: nil,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("valid_factory", func(t *testing.T) {
		t.Parallel()
		f := nonNilTestFactory()
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
			ProviderData: f,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("invalid_provider_data", func(t *testing.T) {
		t.Parallel()
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
			ProviderData: "wrong-type",
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestNewKibanaDataSource_Metadata(t *testing.T) {
	t.Parallel()
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: testProviderTypeName,
	}, &resp)

	require.Equal(t, "elasticstack_kibana_test_entity", resp.TypeName)
}

func TestNewKibanaDataSource_Read_unconfiguredFactory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema, testReadFunc)

	// Configure with a zero-value factory (no default client).
	var cfgResp datasource.ConfigureResponse
	ds.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{
		ProviderData: new(clients.ProviderClientFactory),
	}, &cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError())

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, nil),
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})

	fullSchema := getTestSchema(context.Background())
	fullSchema.Blocks = map[string]dsschema.Block{
		"kibana_connection": providerschema.GetKbFWConnectionBlock(),
	}

	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    objValue,
			Schema: fullSchema,
		},
	}

	var resp datasource.ReadResponse
	ds.Read(ctx, req, &resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Provider not configured")
}

func TestKibanaConnectionField_configDecode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"kibana_connection": connBlockType,
		},
	}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	})

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}

	config := tfsdk.Config{
		Raw:    objValue,
		Schema: schema,
	}

	var model struct {
		KibanaConnectionField
	}
	diags := config.Get(ctx, &model)
	require.False(t, diags.HasError())
	require.True(t, model.KibanaConnection.IsNull())
}

func TestKibanaConnectionField_stateRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	connBlockType := kibanaConnectionBlockType()
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"kibana_connection": connBlockType,
		},
	}

	schema := dsschema.Schema{
		Blocks: map[string]dsschema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}

	state := tfsdk.State{
		Raw:    tftypes.NewValue(objType, nil),
		Schema: schema,
	}

	model := struct {
		KibanaConnectionField
	}{
		KibanaConnectionField: KibanaConnectionField{
			KibanaConnection: providerschema.KibanaConnectionNullList(),
		},
	}

	diags := state.Set(ctx, &model)
	require.False(t, diags.HasError())

	var result struct {
		KibanaConnectionField
	}
	diags = state.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.True(t, result.KibanaConnection.IsNull())
}

// =============================================================================
// Version-requirement test infrastructure
// =============================================================================

// newMockKibanaStatusServer returns an httptest.Server that serves a minimal
// Kibana status JSON payload for GET /api/status. The caller must close the
// returned server.
func newMockKibanaStatusServer(versionStr, buildFlavor string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":%q,"build_flavor":%q}}`, versionStr, buildFlavor)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// newKibanaFactoryForURL builds a *clients.ProviderClientFactory whose default
// Kibana client points to kibanaURL. It sets
// TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT so that KIBANA_ENDPOINT
// env var cannot override the configured URL.
func newKibanaFactoryForURL(t *testing.T, kibanaURL string) *clients.ProviderClientFactory {
	t.Helper()
	t.Setenv(config.PreferConfiguredKibanaEndpointEnvVar, "true")

	ctx := context.Background()
	cfg := config.ProviderConfiguration{
		Kibana: []config.KibanaConnection{
			{
				Username:    types.StringValue("elastic"),
				Password:    types.StringValue("changeme"),
				APIKey:      types.StringValue(""),
				BearerToken: types.StringValue(""),
				Endpoints: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue(kibanaURL),
				}),
				CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
				Insecure: types.BoolValue(false),
			},
		},
	}

	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, cfg, "test-version")
	require.False(t, diags.HasError(), "factory construction must not fail: %v", diags)
	return factory
}

// newKibanaFactoryMinimal builds a *clients.ProviderClientFactory from an
// empty ProviderConfiguration. The resulting default client is non-nil so
// GetKibanaClient succeeds; however the scoped client has no configured
// endpoint so any HTTP method on it will fail. This is sufficient for read
// functions that do not make HTTP calls.
func newKibanaFactoryMinimal(t *testing.T) *clients.ProviderClientFactory {
	t.Helper()
	// Prevent environment variables from injecting unexpected endpoints.
	t.Setenv("KIBANA_ENDPOINT", "")
	t.Setenv("FLEET_ENDPOINT", "")

	ctx := context.Background()
	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, config.ProviderConfiguration{}, "test-version")
	require.False(t, diags.HasError(), "minimal factory construction must not fail: %v", diags)
	return factory
}

// configureDataSource calls Configure on ds with the given factory and asserts
// no errors.
func configureDataSource(t *testing.T, ds datasource.DataSource, factory *clients.ProviderClientFactory) {
	t.Helper()
	var cfgResp datasource.ConfigureResponse
	ds.(datasource.DataSourceWithConfigure).Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: factory,
	}, &cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError(), "Configure must not produce errors: %v", cfgResp.Diagnostics)
}

// buildReadRequestForSchema constructs a datasource.ReadRequest for the given
// schema (with kibana_connection already injected). The object value has a
// null kibana_connection and null string id.
func buildReadRequestForSchema(schema dsschema.Schema) datasource.ReadRequest {
	connBlockType := kibanaConnectionBlockType()
	attrTypes := map[string]tftypes.Type{
		"kibana_connection": connBlockType,
	}
	attrValues := map[string]tftypes.Value{
		"kibana_connection": tftypes.NewValue(connBlockType, nil),
	}
	// Add any string attributes as null strings.
	for name, attr := range schema.Attributes {
		_ = attr
		attrTypes[name] = tftypes.String
		attrValues[name] = tftypes.NewValue(tftypes.String, nil)
	}
	objType := tftypes.Object{AttributeTypes: attrTypes}
	objValue := tftypes.NewValue(objType, attrValues)
	return datasource.ReadRequest{
		Config: tfsdk.Config{Raw: objValue, Schema: schema},
	}
}

// =============================================================================
// Model types used in version-requirement tests
// =============================================================================

// modelNoVersionReqs is an alias for testModel to make test intent clearer.
// It does NOT implement WithVersionRequirements.
// (reuses the existing testModel type)

// modelWithVersionReqsDiagError always returns an error diagnostic from
// GetVersionRequirements. This exercises the "Version requirement diagnostics
// stop read" scenario entirely through the Read path.
type modelWithVersionReqsDiagError struct {
	KibanaConnectionField
	ID types.String `tfsdk:"id"`
}

func (*modelWithVersionReqsDiagError) GetVersionRequirements() ([]VersionRequirement, diag.Diagnostics) {
	return nil, diag.Diagnostics{
		diag.NewErrorDiagnostic("version requirements error", "injected GetVersionRequirements failure"),
	}
}

func getModelWithVersionReqsDiagErrorSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{Computed: true},
		},
		Blocks: map[string]dsschema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}
}

// supportedVersionModel has a minimum version that the mock server at 8.19.0
// will satisfy.
type supportedVersionModel struct {
	KibanaConnectionField
	ID types.String `tfsdk:"id"`
}

func (*supportedVersionModel) GetVersionRequirements() ([]VersionRequirement, diag.Diagnostics) {
	minVer := goversion.Must(goversion.NewVersion("8.0.0"))
	return []VersionRequirement{{MinVersion: *minVer, ErrorMessage: "needs 8.0.0"}}, nil
}

func getSupportedVersionModelSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{Computed: true},
		},
		Blocks: map[string]dsschema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}
}

// unsupportedVersionModel has a minimum version that the mock server at 7.17.0
// will NOT satisfy.
type unsupportedVersionModel struct {
	KibanaConnectionField
	ID types.String `tfsdk:"id"`
}

func (*unsupportedVersionModel) GetVersionRequirements() ([]VersionRequirement, diag.Diagnostics) {
	minVer := goversion.Must(goversion.NewVersion("8.0.0"))
	return []VersionRequirement{{MinVersion: *minVer, ErrorMessage: "requires Kibana 8.0.0 or later"}}, nil
}

// =============================================================================
// Subtask 2.1: model without version requirements
// =============================================================================

// TestNewKibanaDataSource_noVersionReqs_typeAssertionFalse confirms that the
// standard testModel (no version-requirements interface) does NOT satisfy
// WithVersionRequirements for either value or pointer forms,
// so the envelope correctly skips the version-check branch.
func TestNewKibanaDataSource_noVersionReqs_typeAssertionFalse(t *testing.T) {
	t.Parallel()
	var m testModel
	_, ok := any(m).(WithVersionRequirements)
	require.False(t, ok, "value testModel must not satisfy WithVersionRequirements")
	_, ok = any(&m).(WithVersionRequirements)
	require.False(t, ok, "*testModel must not satisfy WithVersionRequirements")
}

// TestNewKibanaDataSource_Read_noVersionReqs_readFuncInvoked proves that when a
// model does NOT implement WithVersionRequirements the envelope
// calls readFunc and persists state normally.
//
// Scenario: Model without version requirements reads normally.
//
// NOTE: Uses t.Setenv via newKibanaFactoryMinimal; must NOT call t.Parallel.
func TestNewKibanaDataSource_Read_noVersionReqs_readFuncInvoked(t *testing.T) {
	ctx := context.Background()

	readFuncCalled := false
	ds := NewKibanaDataSource[testModel](ComponentKibana, "test_entity", getTestSchema,
		func(_ context.Context, _ *clients.KibanaScopedClient, model testModel) (testModel, diag.Diagnostics) {
			readFuncCalled = true
			model.ID = types.StringValue("no-version-reqs-result")
			return model, nil
		},
	)

	// Use a minimal factory: GetKibanaClient succeeds (defaultClient non-nil)
	// but no HTTP calls are made because readFunc ignores the client.
	factory := newKibanaFactoryMinimal(t)
	configureDataSource(t, ds, factory)

	schemaWithConn := getTestSchema(context.Background())
	schemaWithConn.Blocks = map[string]dsschema.Block{
		"kibana_connection": providerschema.GetKbFWConnectionBlock(),
	}
	req := buildReadRequestForSchema(schemaWithConn)

	var resp datasource.ReadResponse
	resp.State = tfsdk.State{Schema: schemaWithConn}
	ds.Read(ctx, req, &resp)

	// Guard: if the factory's defaultClient is nil, GetKibanaClient returns a
	// "Provider not configured" error before readFunc is ever reached, causing a
	// false pass when asserting readFuncCalled == false. Detect that here.
	for _, d := range resp.Diagnostics {
		require.NotEqual(t, "Provider not configured", d.Summary(),
			"factory must be configured — test would give a false pass if factory has nil client")
	}
	require.True(t, readFuncCalled, "readFunc must be called when model has no version requirements")
	require.False(t, resp.Diagnostics.HasError(), "Read must not produce errors: %v", resp.Diagnostics)

	var result testModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "no-version-reqs-result", result.ID.ValueString(),
		"state must reflect the value set by readFunc")
}

// =============================================================================
// Subtask 2.2: model WITH version requirements
// =============================================================================

// TestWithVersionRequirements_dataSourcePointerAssertionTrue confirms that
// a model implementing GetVersionRequirements on its pointer receiver satisfies
// the interface after the any(&model) cast used inside the envelope.
func TestWithVersionRequirements_dataSourcePointerAssertionTrue(t *testing.T) {
	t.Parallel()
	var m modelWithVersionReqsDiagError
	// Value form must NOT satisfy the interface (method on pointer receiver).
	_, ok := any(m).(WithVersionRequirements)
	require.False(t, ok, "value modelWithVersionReqsDiagError must not satisfy the interface")
	// Pointer form MUST satisfy it — this matches any(&model) in the envelope.
	_, ok = any(&m).(WithVersionRequirements)
	require.True(t, ok, "*modelWithVersionReqsDiagError must satisfy WithVersionRequirements")
}

// TestKibanaDataSource_Read_versionReqDiagsStopRead exercises the full Read
// path when GetVersionRequirements returns error diagnostics. The envelope must
// short-circuit before calling readFunc.
//
// Scenario: Version requirement diagnostics stop read.
//
// NOTE: Uses t.Setenv via newKibanaFactoryMinimal; must NOT call t.Parallel.
func TestKibanaDataSource_Read_versionReqDiagsStopRead(t *testing.T) {
	ctx := context.Background()

	readFuncCalled := false
	ds := NewKibanaDataSource[modelWithVersionReqsDiagError](ComponentKibana, "diag_err_entity",
		func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{
				Attributes: map[string]dsschema.Attribute{
					"id": dsschema.StringAttribute{Computed: true},
				},
			}
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, model modelWithVersionReqsDiagError) (modelWithVersionReqsDiagError, diag.Diagnostics) {
			readFuncCalled = true
			return model, nil
		},
	)

	// Factory must succeed so GetKibanaClient does not short-circuit first.
	factory := newKibanaFactoryMinimal(t)
	configureDataSource(t, ds, factory)

	schema := getModelWithVersionReqsDiagErrorSchema(context.Background())
	req := buildReadRequestForSchema(schema)

	var resp datasource.ReadResponse
	ds.Read(ctx, req, &resp)

	require.False(t, readFuncCalled, "readFunc must NOT be called when GetVersionRequirements returns error diags")
	require.True(t, resp.Diagnostics.HasError(), "Read must propagate error from GetVersionRequirements")

	summaries := make([]string, 0)
	for _, e := range resp.Diagnostics.Errors() {
		summaries = append(summaries, e.Summary())
	}
	require.Contains(t, summaries, "version requirements error",
		"diagnostic from GetVersionRequirements must be appended; got: %v", summaries)
}

// TestKibanaDataSource_Read_supportedServer_invokesReadFunc tests the
// "Supported server invokes read function" scenario end-to-end with an
// httptest Kibana server reporting a version that satisfies the minimum
// requirement.
//
// NOTE: Uses t.Setenv via newKibanaFactoryForURL; must NOT call t.Parallel.
func TestKibanaDataSource_Read_supportedServer_invokesReadFunc(t *testing.T) {
	ctx := context.Background()

	srv := newMockKibanaStatusServer("8.19.0", "default")
	defer srv.Close()

	readFuncCalled := false
	ds := NewKibanaDataSource[supportedVersionModel](ComponentKibana, "supported_entity",
		func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{
				Attributes: map[string]dsschema.Attribute{
					"id": dsschema.StringAttribute{Computed: true},
				},
			}
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, model supportedVersionModel) (supportedVersionModel, diag.Diagnostics) {
			readFuncCalled = true
			model.ID = types.StringValue("supported-result")
			return model, nil
		},
	)

	factory := newKibanaFactoryForURL(t, srv.URL)
	configureDataSource(t, ds, factory)

	schema := getSupportedVersionModelSchema(context.Background())
	req := buildReadRequestForSchema(schema)

	var resp datasource.ReadResponse
	resp.State = tfsdk.State{Schema: schema}
	ds.Read(ctx, req, &resp)

	require.False(t, resp.Diagnostics.HasError(),
		"Read must succeed when server satisfies minimum version: %v", resp.Diagnostics)
	require.True(t, readFuncCalled, "readFunc must be invoked when server satisfies minimum version")

	var result supportedVersionModel
	diags := resp.State.Get(ctx, &result)
	require.False(t, diags.HasError())
	require.Equal(t, "supported-result", result.ID.ValueString())
}

// TestKibanaDataSource_Read_unsupportedServer_stopsBeforeReadFunc tests the
// "Unsupported server stops before read function" scenario end-to-end with an
// httptest Kibana server reporting a version below the minimum requirement.
//
// NOTE: Uses t.Setenv via newKibanaFactoryForURL; must NOT call t.Parallel.
func TestKibanaDataSource_Read_unsupportedServer_stopsBeforeReadFunc(t *testing.T) {
	ctx := context.Background()

	// Server reports 7.17.0, which is below the required 8.0.0.
	srv := newMockKibanaStatusServer("7.17.0", "default")
	defer srv.Close()

	readFuncCalled := false
	ds := NewKibanaDataSource[unsupportedVersionModel](ComponentKibana, "unsupported_entity",
		func(_ context.Context) dsschema.Schema {
			return dsschema.Schema{
				Attributes: map[string]dsschema.Attribute{
					"id": dsschema.StringAttribute{Computed: true},
				},
			}
		},
		func(_ context.Context, _ *clients.KibanaScopedClient, model unsupportedVersionModel) (unsupportedVersionModel, diag.Diagnostics) {
			readFuncCalled = true
			return model, nil
		},
	)

	factory := newKibanaFactoryForURL(t, srv.URL)
	configureDataSource(t, ds, factory)

	schema := getSupportedVersionModelSchema(context.Background())
	req := buildReadRequestForSchema(schema)

	var resp datasource.ReadResponse
	ds.Read(ctx, req, &resp)

	require.False(t, readFuncCalled,
		"readFunc must NOT be called when server is below minimum version")
	require.True(t, resp.Diagnostics.HasError(),
		"Read must produce an error diagnostic for unsupported server")

	var foundUnsupported bool
	for _, e := range resp.Diagnostics.Errors() {
		if e.Summary() == "Unsupported server version" {
			foundUnsupported = true
			require.Contains(t, e.Detail(), "requires Kibana 8.0.0 or later",
				"Unsupported server version detail must contain the model error message")
		}
	}
	require.True(t, foundUnsupported,
		"must have an 'Unsupported server version' diagnostic; got: %v", resp.Diagnostics.Errors())
}
