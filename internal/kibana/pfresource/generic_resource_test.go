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

package pfresource

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAPI struct {
	createCalls int
	getCalls    int
	updateCalls int
	deleteCalls int
	lastID      string
	lastSpace   string
	createID    string
	result      string
	present     bool
	diags       diag.Diagnostics
}

type fakeAssembly struct {
	api ResourceAPI[string, string, string]
}

type fakeModel struct {
	id               types.String
	spaceID          types.String
	kibanaConnection types.List
	versionReq       VersionRequirement
	createRequest    string
	updateRequest    string
	remoteValue      string
	populateCalls    int
}

func (a fakeAssembly) TypeNameSuffix() string                   { return "kibana_fake" }
func (a fakeAssembly) API() ResourceAPI[string, string, string] { return a.api }
func (a fakeAssembly) NewModel() *fakeModel                     { return &fakeModel{} }
func (a fakeAssembly) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughID(ctx, "id", req, resp)
}

func (a *fakeAPI) Create(_ context.Context, _ *kibanaoapi.Client, spaceID string, request string) (string, diag.Diagnostics) {
	a.createCalls++
	a.lastSpace = spaceID
	a.lastID = request
	if a.createID == "" {
		return request, a.diags
	}
	return a.createID, a.diags
}
func (a *fakeAPI) Get(_ context.Context, _ *kibanaoapi.Client, spaceID string, resourceID string) (string, bool, diag.Diagnostics) {
	a.getCalls++
	a.lastSpace = spaceID
	a.lastID = resourceID
	return a.result, a.present, a.diags
}
func (a *fakeAPI) Update(_ context.Context, _ *kibanaoapi.Client, spaceID string, resourceID string, request string) diag.Diagnostics {
	a.updateCalls++
	a.lastSpace = spaceID
	a.lastID = resourceID + ":" + request
	return a.diags
}
func (a *fakeAPI) Delete(_ context.Context, _ *kibanaoapi.Client, spaceID string, resourceID string) diag.Diagnostics {
	a.deleteCalls++
	a.lastSpace = spaceID
	a.lastID = resourceID
	return a.diags
}

func (m *fakeModel) GetKibanaConnection() types.List { return m.kibanaConnection }
func (m *fakeModel) GetID() types.String             { return m.id }
func (m *fakeModel) SetID(id types.String)           { m.id = id }
func (m *fakeModel) GetSpaceID() types.String        { return m.spaceID }
func (m *fakeModel) SetSpaceID(id types.String)      { m.spaceID = id }
func (m *fakeModel) VersionRequirement() VersionRequirement {
	return m.versionReq
}
func (m *fakeModel) ToCreateRequest(context.Context) (string, diag.Diagnostics) {
	return m.createRequest, nil
}
func (m *fakeModel) ToUpdateRequest(context.Context) (string, diag.Diagnostics) {
	return m.updateRequest, nil
}
func (m *fakeModel) PopulateFromRemote(_ context.Context, spaceID string, remote string) diag.Diagnostics {
	m.populateCalls++
	m.spaceID = types.StringValue(spaceID)
	m.remoteValue = remote
	return nil
}

func TestConfigure_PropagatesConversionDiagnostics(t *testing.T) {
	resp := &resource.ConfigureResponse{}

	result := Configure(context.Background(), "unexpected", resp)

	assert.Nil(t, result)
	require.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, "Unexpected Provider Data", resp.Diagnostics[0].Summary())
}

func TestMetadata_SetsTypeName(t *testing.T) {
	resp := &resource.MetadataResponse{}

	Metadata(resource.MetadataRequest{ProviderTypeName: "elasticstack"}, resp, "kibana_agentbuilder_tool")

	assert.Equal(t, "elasticstack_kibana_agentbuilder_tool", resp.TypeName)
}

func TestImportStatePassthroughID_RequiresInitializedState(t *testing.T) {
	ctx := context.Background()
	resp := &resource.ImportStateResponse{}

	require.Panics(t, func() {
		ImportStatePassthroughID(ctx, "id", resource.ImportStateRequest{ID: "tool-123"}, resp)
	})
}

func TestImportStateCompositeID_RequiresInitializedState(t *testing.T) {
	ctx := context.Background()
	resp := &resource.ImportStateResponse{}

	require.Panics(t, func() {
		ImportStateCompositeID(ctx, resource.ImportStateRequest{ID: "observability/tool-123"}, resp, "id", "space_id")
	})
}

func TestImportStateCompositeID_PropagatesParseDiagnostics(t *testing.T) {
	ctx := context.Background()
	resp := &resource.ImportStateResponse{}

	require.Panics(t, func() {
		ImportStateCompositeID(ctx, resource.ImportStateRequest{ID: "not-a-composite-id"}, resp, "id", "space_id")
	})
}

func TestComposeAndParseCompositeID(t *testing.T) {
	id := ComposeCompositeID("observability", "tool-123")
	assert.Equal(t, "observability/tool-123", id)

	parsed, diags := ParseCompositeID(id)
	require.False(t, diags.HasError())
	assert.Equal(t, "observability", parsed.ClusterID)
	assert.Equal(t, "tool-123", parsed.ResourceID)
}

func TestEffectiveSpaceID_DefaultsToDefault(t *testing.T) {
	assert.Equal(t, DefaultSpaceID, EffectiveSpaceID(types.StringNull()))
	assert.Equal(t, DefaultSpaceID, EffectiveSpaceID(types.StringUnknown()))
	assert.Equal(t, DefaultSpaceID, EffectiveSpaceID(types.StringValue("")))
	assert.Equal(t, "security", EffectiveSpaceID(types.StringValue("security")))
}

func TestResolveKibanaClient_WithNilFactoryReturnsDiagnostic(t *testing.T) {
	scoped, diags := ResolveKibanaClient(context.Background(), nil, types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{}}))

	assert.Nil(t, scoped)
	require.True(t, diags.HasError())
	assert.Equal(t, "Provider not configured", diags[0].Summary())
}

func TestResolveRuntime_PropagatesVersionFailure(t *testing.T) {
	model := &fakeModel{
		kibanaConnection: types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{}}),
		versionReq: VersionRequirement{
			MinimumVersion: version.Must(version.NewVersion("9.3.0")),
		},
	}

	runtime, diags := ResolveRuntime[string, string, string](context.Background(), &clients.ProviderClientFactory{}, model)

	assert.Nil(t, runtime)
	require.True(t, diags.HasError())
}

func TestEnforceVersion_WithNilMinimumVersionIsNoop(t *testing.T) {
	diags := EnforceVersion(context.Background(), nil, VersionRequirement{})
	assert.False(t, diags.HasError())
}

func TestEnforceVersion_WhenClientLookupFailsReturnsFrameworkDiagnostics(t *testing.T) {
	client := &clients.KibanaScopedClient{}
	req := VersionRequirement{MinimumVersion: version.Must(version.NewVersion("9.3.0"))}

	diags := EnforceVersion(context.Background(), client, req)

	require.True(t, diags.HasError())
}

func TestReadAfterWrite_UsesFollowUpRead(t *testing.T) {
	api := &fakeAPI{result: "created-id", present: true}
	result, diags := ReadAfterWrite(context.Background(), api, nil, "default", "created-id")

	require.False(t, diags.HasError())
	assert.Equal(t, 1, api.getCalls)
	assert.Equal(t, "default", api.lastSpace)
	assert.Equal(t, "created-id", api.lastID)
	assert.Equal(t, "created-id", result)
}

func TestReadAfterWrite_NotFoundReturnsDiagnostic(t *testing.T) {
	api := &fakeAPI{present: false}
	_, diags := ReadAfterWrite(context.Background(), api, nil, "default", "created-id")

	require.True(t, diags.HasError())
	assert.Equal(t, "Resource not found after write", diags[0].Summary())
}

func TestReadRemote_SuccessReturnsPresentTrue(t *testing.T) {
	api := &fakeAPI{result: "tool-123", present: true}
	result, present, diags := ReadRemote(context.Background(), api, nil, "default", "tool-123")

	require.False(t, diags.HasError())
	assert.True(t, present)
	assert.Equal(t, "tool-123", result)
}

func TestReadRemote_NotFoundReturnsPresentFalse(t *testing.T) {
	api := &fakeAPI{present: false}
	result, present, diags := ReadRemote(context.Background(), api, nil, "default", "missing")

	require.False(t, diags.HasError())
	assert.False(t, present)
	assert.Empty(t, result)
	assert.Equal(t, 1, api.getCalls)
}

func TestReadRemote_PropagatesDiagnostics(t *testing.T) {
	api := &fakeAPI{diags: diag.Diagnostics{diag.NewErrorDiagnostic("boom", "detail")}}
	result, present, diags := ReadRemote(context.Background(), api, nil, "default", "broken")

	require.True(t, diags.HasError())
	assert.False(t, present)
	assert.Empty(t, result)
	assert.Equal(t, "boom", diags[0].Summary())
}

func TestOrchestratorCreate_ReadsAuthoritativeState(t *testing.T) {
	api := &fakeAPI{createID: "remote-id", result: "remote-state", present: true}
	orchestrator := Orchestrator[string, string, string, *fakeModel]{
		Factory:  &clients.ProviderClientFactory{},
		Assembly: fakeAssembly{api: api},
	}
	model := &fakeModel{kibanaConnection: types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{}}), createRequest: "plan-id"}

	updated, diags := orchestrator.Create(context.Background(), model, "default")

	require.True(t, diags.HasError())
	require.NotNil(t, updated)
	assert.Equal(t, 0, api.createCalls)
}

func TestVersionRequirement_StructureCarriesConfiguredMessage(t *testing.T) {
	req := VersionRequirement{
		MinimumVersion: version.Must(version.NewVersion("9.3.0")),
		ErrorSummary:   "Unsupported server version",
		ErrorDetail:    "Agent Builder agents require Elastic Stack v9.3.0 or later.",
	}

	assert.Equal(t, "9.3.0", req.MinimumVersion.String())
	assert.Equal(t, "Unsupported server version", req.ErrorSummary)
	assert.Equal(t, "Agent Builder agents require Elastic Stack v9.3.0 or later.", req.ErrorDetail)
}

func TestAssembly_ExposesTypeNameSuffixAndAPI(t *testing.T) {
	api := &fakeAPI{}
	assembly := fakeAssembly{api: api}

	assert.Equal(t, "kibana_fake", assembly.TypeNameSuffix())
	assert.Same(t, api, assembly.API())
}

var _ ResourceAPI[string, string, string] = (*fakeAPI)(nil)
var _ Assembly[string, string, string, *fakeModel] = fakeAssembly{}
var _ ModelContract[string, string, string] = (*fakeModel)(nil)
