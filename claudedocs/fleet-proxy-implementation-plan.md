# Fleet Proxy Implementation Plan

**Date**: 2026-01-14
**Feature**: Add Fleet Proxy resource and integrate with Output/Server Host resources
**Goal**: Implement proxy configuration support WITHOUT misimplementation

---

## Executive Summary

This plan outlines the implementation of Fleet Proxy support in the Terraform provider. The feature enables users to configure proxies for Fleet agent communication and reference them in Elasticsearch outputs and Fleet Server hosts.

**Key Complexity**: Space-aware resources with dependency validation (proxy must exist in space before being referenced).

---

## API Analysis

### Generated API Schema (from `kibana.gen.go`)

```go
// POST /api/fleet/proxies - Create Proxy
type PostFleetProxiesJSONBody struct {
    Certificate            *string  `json:"certificate,omitempty"`              // Optional: Client cert
    CertificateAuthorities *string  `json:"certificate_authorities,omitempty"`  // Optional: CA cert
    CertificateKey         *string  `json:"certificate_key,omitempty"`          // Optional: Client key (SENSITIVE)
    Id                     *string  `json:"id,omitempty"`                       // Optional: Custom ID
    IsPreconfigured        *bool    `json:"is_preconfigured,omitempty"`         // Optional: Preconfigured flag
    Name                   string   `json:"name"`                               // Required: Proxy name
    ProxyHeaders           *map[string]any `json:"proxy_headers,omitempty"`     // Optional: Custom headers
    Url                    string   `json:"url"`                                // Required: Proxy URL
}

// PUT /api/fleet/proxies/{itemId} - Update Proxy
type PutFleetProxiesItemidJSONBody struct {
    Certificate            *string  `json:"certificate,omitempty"`
    CertificateAuthorities *string  `json:"certificate_authorities,omitempty"`
    CertificateKey         *string  `json:"certificate_key,omitempty"`
    Name                   *string  `json:"name,omitempty"`                     // Optional in update
    ProxyHeaders           *map[string]any `json:"proxy_headers,omitempty"`
    Url                    *string  `json:"url,omitempty"`                      // Optional in update
}
```

### API Endpoints

- `GET /api/fleet/proxies` - List all proxies
- `POST /api/fleet/proxies` - Create proxy
- `GET /api/fleet/proxies/{itemId}` - Get specific proxy
- `PUT /api/fleet/proxies/{itemId}` - Update proxy
- `DELETE /api/fleet/proxies/{itemId}` - Delete proxy

### Space Awareness

Proxies are space-aware resources. API calls must use `spaceAwarePathRequestEditor(spaceID)` pattern:
- Empty or "default" spaceID → `/api/fleet/proxies`
- Non-default spaceID → `/s/{spaceID}/api/fleet/proxies`

---

## Resource Design

### 1. Proxy Resource Schema

**Resource Name**: `elasticstack_fleet_proxy`

**Terraform Schema**:
```hcl
resource "elasticstack_fleet_proxy" "example" {
  name                     = "my-proxy"                    # Required
  url                      = "https://proxy.example.com"   # Required
  proxy_id                 = "custom-id"                   # Optional (Computed if not set)
  certificate              = file("client.crt")            # Optional
  certificate_authorities  = file("ca.crt")                # Optional
  certificate_key          = file("client.key")            # Optional, Sensitive
  is_preconfigured         = false                         # Optional
  proxy_headers = {                                        # Optional
    "X-Custom-Header" = "value"
  }
  space_ids                = ["default", "space-1"]        # Optional, Computed
}
```

**Field Details**:

| Field | Type | Required | Computed | Sensitive | Notes |
|-------|------|----------|----------|-----------|-------|
| `id` | string | No | Yes | No | Terraform resource ID |
| `proxy_id` | string | No | Yes | No | Kibana proxy ID (optional, allows custom ID) |
| `name` | string | Yes | No | No | Proxy name |
| `url` | string | Yes | No | No | Proxy URL (e.g., `https://proxy:8080`) |
| `certificate` | string | No | No | No | Client SSL certificate (PEM format) |
| `certificate_authorities` | string | No | No | No | CA certificate (PEM format) |
| `certificate_key` | string | No | No | Yes | Client SSL key (PEM format) **SENSITIVE** |
| `is_preconfigured` | bool | No | No | No | Mark as preconfigured (read-only in UI) |
| `proxy_headers` | map[string]string | No | No | No | Custom HTTP headers |
| `space_ids` | set(string) | No | Yes | No | Kibana spaces where proxy is available |

**Plan Modifiers**:
- `id` → `UseStateForUnknown()`
- `proxy_id` → `UseStateForUnknown()`, `RequiresReplace()`
- `space_ids` → `UseStateForUnknown()`

---

### 2. Update Output Resource

**File**: `internal/fleet/output/schema.go`

**Add Field**:
```go
"proxy_id": schema.StringAttribute{
    Description: "ID of the Fleet proxy to use for this output. The proxy must exist in the same space(s) as the output.",
    Optional:    true,
},
```

**API Integration**: Already exists in generated API
- Outputs can have `proxy_id` field in Elasticsearch/Logstash/Kafka configurations

---

### 3. Update Fleet Server Host Resource

**File**: `internal/fleet/server_host/schema.go`

**Add Field**:
```go
"proxy_id": schema.StringAttribute{
    Description: "ID of the Fleet proxy to use for Fleet Server connections. The proxy must exist in the same space(s) as the server host.",
    Optional:    true,
},
```

---

## Implementation Structure

### Directory Structure

```
internal/fleet/proxy/
├── resource.go           # Resource definition and registration
├── schema.go             # Terraform schema definition
├── models.go             # Go structs and API conversion
├── create.go             # Create operation
├── read.go               # Read operation
├── update.go             # Update operation
├── delete.go             # Delete operation
├── acc_test.go           # Acceptance tests
└── testdata/
    ├── TestAccResourceProxy/
    │   ├── create/
    │   │   └── main.tf
    │   └── update/
    │       └── main.tf
    └── TestAccResourceProxyWithSpaces/
        └── ...
```

### Client Layer Functions

**File**: `internal/clients/fleet/fleet.go`

**Add Functions**:
```go
// GetProxy retrieves a proxy by ID within a specific space
func GetProxy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.Proxy, diag.Diagnostics)

// ListProxies retrieves all proxies within a specific space
func ListProxies(ctx context.Context, client *Client, spaceID string) ([]kbapi.Proxy, diag.Diagnostics)

// CreateProxy creates a new proxy within a specific space
func CreateProxy(ctx context.Context, client *Client, spaceID string, req kbapi.PostFleetProxiesJSONBody) (*kbapi.Proxy, diag.Diagnostics)

// UpdateProxy updates an existing proxy within a specific space
func UpdateProxy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PutFleetProxiesItemidJSONBody) (*kbapi.Proxy, diag.Diagnostics)

// DeleteProxy deletes a proxy within a specific space
func DeleteProxy(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics
```

**Pattern**: Follow existing Output/AgentPolicy functions for space awareness.

---

## Space-Aware Validation

### Validation Requirements

1. **Proxy Creation**: Must specify which spaces proxy is available in
2. **Output/ServerHost → Proxy Reference**:
   - If output/serverhost has `space_ids`, proxy must exist in ALL those spaces
   - If output/serverhost doesn't specify spaces (default), proxy must be in default space

### Validation Strategy

**Option A: Runtime API Validation (RECOMMENDED)**
- Let Kibana API handle validation
- If proxy not in correct space, API returns error
- Provider surfaces clear error to user

**Option B: Provider-Side Validation**
- Fetch proxy details during plan
- Check space overlap
- Add custom validator

**Decision: Use Option A** because:
- Simpler implementation
- Always accurate (API is source of truth)
- Avoids race conditions
- Follows existing pattern in provider

### Error Handling

Clear error messages when validation fails:
```
Error: Proxy not available in required space

The proxy "my-proxy" (ID: proxy-123) is not available in space "production".
Make sure the proxy is configured with space_ids = ["production"] or the
output is configured to use a space where the proxy exists.
```

---

## Implementation Phases

### Phase 1: Proxy Resource (Core)

**Files to Create**:
1. `internal/fleet/proxy/resource.go` - Resource registration
2. `internal/fleet/proxy/schema.go` - Terraform schema
3. `internal/fleet/proxy/models.go` - Model conversions
4. `internal/fleet/proxy/create.go` - Create operation
5. `internal/fleet/proxy/read.go` - Read operation
6. `internal/fleet/proxy/update.go` - Update operation
7. `internal/fleet/proxy/delete.go` - Delete operation

**Files to Modify**:
1. `internal/clients/fleet/fleet.go` - Add proxy client functions
2. `internal/provider/provider.go` - Register proxy resource

**Testing**:
1. Basic CRUD operations
2. Space-aware operations
3. Certificate handling (especially sensitive `certificate_key`)
4. Proxy headers handling

**Acceptance Criteria**:
- ✅ Can create proxy with minimum required fields (name, url)
- ✅ Can create proxy with all optional fields
- ✅ Can update proxy fields
- ✅ Can delete proxy
- ✅ Space IDs properly handled (default vs custom spaces)
- ✅ Certificate_key marked sensitive in state
- ✅ Proxy headers properly serialized/deserialized

---

### Phase 2: Output Integration

**Files to Modify**:
1. `internal/fleet/output/schema.go` - Add `proxy_id` field
2. `internal/fleet/output/models.go` - Handle proxy_id in API conversion
3. `internal/fleet/output/models_elasticsearch.go` - Elasticsearch-specific proxy handling
4. `internal/fleet/output/models_logstash.go` - Logstash-specific proxy handling
5. `internal/fleet/output/models_kafka.go` - Kafka-specific proxy handling

**Testing**:
1. Output with proxy_id reference
2. Output without proxy_id (backward compatibility)
3. Space validation (output + proxy in same space)
4. Update output to add/remove proxy

**Acceptance Criteria**:
- ✅ Can create output with proxy_id
- ✅ Can update output to add proxy_id
- ✅ Can update output to remove proxy_id
- ✅ Backward compatible (existing outputs without proxy_id still work)
- ✅ Clear error if proxy doesn't exist
- ✅ Clear error if proxy not in correct space

---

### Phase 3: Server Host Integration

**Files to Modify**:
1. `internal/fleet/server_host/schema.go` - Add `proxy_id` field
2. `internal/fleet/server_host/models.go` - Handle proxy_id in API conversion

**Testing**:
1. Server host with proxy_id reference
2. Server host without proxy_id (backward compatibility)
3. Space validation (server host + proxy in same space)

**Acceptance Criteria**:
- ✅ Can create server host with proxy_id
- ✅ Can update server host to add proxy_id
- ✅ Can update server host to remove proxy_id
- ✅ Backward compatible
- ✅ Clear error if proxy doesn't exist
- ✅ Clear error if proxy not in correct space

---

## Testing Strategy

### Unit Tests

**File**: `internal/fleet/proxy/models_test.go`

Test model conversions:
- Terraform model → API request
- API response → Terraform model
- Proxy headers serialization
- Space IDs handling

### Acceptance Tests

**File**: `internal/fleet/proxy/acc_test.go`

**Test Cases**:

1. **TestAccResourceProxy_basic**
   - Create proxy with minimum fields
   - Verify fields persisted correctly

2. **TestAccResourceProxy_complete**
   - Create proxy with all fields (including certificates)
   - Update proxy fields
   - Verify certificate_key is sensitive

3. **TestAccResourceProxy_spaces**
   - Create proxy in multiple spaces
   - Verify space_ids handling

4. **TestAccResourceProxy_headers**
   - Create proxy with custom headers
   - Update headers
   - Verify headers preserved

5. **TestAccResourceProxy_withOutput**
   - Create proxy
   - Create output referencing proxy
   - Verify output uses proxy

6. **TestAccResourceProxy_withServerHost**
   - Create proxy
   - Create server host referencing proxy
   - Verify server host uses proxy

7. **TestAccResourceProxy_spaceValidation**
   - Create proxy in space-1
   - Try to create output in space-2 referencing proxy
   - Expect error (proxy not in space-2)

### Integration Test Workflow

```hcl
# Test scenario: Complete proxy setup

resource "elasticstack_fleet_proxy" "corporate" {
  name = "corporate-proxy"
  url  = "https://proxy.corp.example.com:8080"
  proxy_headers = {
    "X-Auth-Token" = "secret"
  }
}

resource "elasticstack_fleet_output" "elasticsearch" {
  name     = "default-elasticsearch"
  type     = "elasticsearch"
  hosts    = ["https://es:9200"]
  proxy_id = elasticstack_fleet_proxy.corporate.proxy_id
}

resource "elasticstack_fleet_server_host" "default" {
  name     = "default-fleet-server"
  hosts    = ["https://fleet:8220"]
  proxy_id = elasticstack_fleet_proxy.corporate.proxy_id
}
```

**Validation**:
- ✅ All resources created successfully
- ✅ Dependencies handled correctly (proxy created before output/server host)
- ✅ Proxy correctly referenced in output and server host
- ✅ Can destroy in correct order (output/server host → proxy)

---

## Critical Implementation Details

### 1. Sensitive Data Handling

**`certificate_key` field MUST be marked sensitive**:
```go
"certificate_key": schema.StringAttribute{
    Description: "Client SSL certificate key (PEM format).",
    Optional:    true,
    Sensitive:   true,  // ← CRITICAL
},
```

**State File**:
- Terraform will encrypt sensitive values in state
- Never log certificate_key in provider logs

### 2. Proxy Headers Handling

**Challenge**: API accepts `map[string]any` but Terraform plugin framework prefers typed values.

**Solution**:
```go
"proxy_headers": schema.MapAttribute{
    Description: "Custom HTTP headers to send with proxy requests.",
    Optional:    true,
    ElementType: types.StringType,  // Restrict to string values for simplicity
},
```

**Alternative** (if complex values needed):
```go
// Use jsontypes.Normalized for flexible JSON
"proxy_headers": schema.StringAttribute{
    Description: "Custom HTTP headers as JSON.",
    Optional:    true,
    CustomType:  jsontypes.NormalizedType{},
},
```

**Decision: Use simple map[string]string** unless API requires complex values.

### 3. Space IDs Handling

**Pattern from Existing Resources**:
```go
"space_ids": schema.SetAttribute{
    Description: "The Kibana space IDs where this proxy is available.",
    ElementType: types.StringType,
    Optional:    true,
    Computed:    true,  // API may return spaces
    PlanModifiers: []planmodifier.Set{
        setplanmodifier.UseStateForUnknown(),
    },
},
```

**Important**: Use `Set` not `List` because order doesn't matter.

### 4. Custom ID Handling

**User can optionally specify proxy_id**:
```hcl
resource "elasticstack_fleet_proxy" "example" {
  proxy_id = "my-custom-proxy-id"  # Optional
  name     = "my-proxy"
  url      = "https://proxy:8080"
}
```

**API Behavior**:
- If `id` provided in POST request → uses that ID
- If `id` not provided → generates UUID

**Schema**:
```go
"proxy_id": schema.StringAttribute{
    Description: "Unique identifier of the proxy. If not specified, Kibana generates a UUID.",
    Computed:    true,
    Optional:    true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),  // Changing ID requires recreation
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

---

## Potential Pitfalls & How to Avoid Them

### ❌ Pitfall 1: Not Using Space-Aware API Calls

**Problem**: Calling `/api/fleet/proxies` directly instead of using `spaceAwarePathRequestEditor`.

**Solution**: Always use the helper:
```go
resp, err := client.API.GetFleetProxiesItemidWithResponse(
    ctx,
    proxyID,
    spaceAwarePathRequestEditor(spaceID),  // ← Always include this
)
```

### ❌ Pitfall 2: Forgetting Sensitive Marking

**Problem**: `certificate_key` not marked sensitive → leaked in logs/output.

**Solution**: Double-check schema has `Sensitive: true`.

### ❌ Pitfall 3: Incorrect Proxy Headers Serialization

**Problem**: API expects `map[string]interface{}` with union types, but we send strings only.

**Solution**:
- Start with `map[string]string` (simple)
- Test with Kibana API
- If API rejects, upgrade to flexible JSON type

### ❌ Pitfall 4: Not Testing Space Validation

**Problem**: Users can create mismatched space configurations that fail at runtime.

**Solution**:
- Write explicit acceptance test for space validation
- Document space requirements clearly
- Provide helpful error messages

### ❌ Pitfall 5: Incorrect Dependency Order

**Problem**: Terraform tries to delete proxy before output/server host.

**Solution**: Terraform handles this automatically via references:
```hcl
proxy_id = elasticstack_fleet_proxy.corporate.proxy_id
         # ↑ This creates implicit dependency
```

No additional work needed.

### ❌ Pitfall 6: Not Handling API Response Correctly

**Problem**: Assuming API returns exactly what we sent.

**Solution**:
- Always populate model from API response in `populateFromAPI()`
- Don't copy from request back to state
- Use framework's plan modifiers for computed fields

### ❌ Pitfall 7: Breaking Backward Compatibility

**Problem**: Adding proxy_id to Output breaks existing configurations.

**Solution**:
- Make proxy_id optional
- Don't change existing field behavior
- Test with existing Output configurations

---

## Documentation Requirements

### 1. Resource Documentation

**File**: `docs/resources/fleet_proxy.md` (auto-generated)

**Include**:
- Description of what Fleet Proxy does
- All schema attributes with descriptions
- Example usage (basic and advanced)
- Space awareness explanation
- Certificate usage examples

### 2. Updated Output Documentation

**File**: `docs/resources/fleet_output.md`

**Add**:
- `proxy_id` attribute description
- Example with proxy reference
- Note about space requirements

### 3. Updated Server Host Documentation

**File**: `docs/resources/fleet_server_host.md`

**Add**:
- `proxy_id` attribute description
- Example with proxy reference
- Note about space requirements

---

## Implementation Checklist

### Pre-Implementation

- [x] API analysis complete
- [x] Schema design complete
- [ ] User review and approval of this plan

### Phase 1: Proxy Resource

- [ ] Create directory structure
- [ ] Implement schema.go
- [ ] Implement models.go with API conversion
- [ ] Implement create.go
- [ ] Implement read.go
- [ ] Implement update.go
- [ ] Implement delete.go
- [ ] Add client functions to fleet.go
- [ ] Register resource in provider.go
- [ ] Write unit tests
- [ ] Write acceptance tests
- [ ] Verify sensitive field handling
- [ ] Verify space awareness

### Phase 2: Output Integration

- [ ] Update output schema
- [ ] Update output models (all types)
- [ ] Write integration tests
- [ ] Verify backward compatibility
- [ ] Test space validation

### Phase 3: Server Host Integration

- [ ] Update server host schema
- [ ] Update server host models
- [ ] Write integration tests
- [ ] Verify backward compatibility
- [ ] Test space validation

### Final

- [ ] Run full test suite
- [ ] Generate documentation
- [ ] Code review
- [ ] User validation

---

## Success Criteria

This implementation is successful when:

1. ✅ **Proxy Resource Works**
   - Can create/read/update/delete proxies
   - Space awareness works correctly
   - Sensitive fields properly handled
   - Proxy headers work correctly

2. ✅ **Output Integration Works**
   - Can reference proxy from output
   - Space validation works
   - Backward compatible (existing outputs still work)

3. ✅ **Server Host Integration Works**
   - Can reference proxy from server host
   - Space validation works
   - Backward compatible

4. ✅ **Tests Pass**
   - All unit tests pass
   - All acceptance tests pass
   - No regressions in existing tests

5. ✅ **Documentation Complete**
   - All resources documented
   - Examples provided
   - Space awareness explained

6. ✅ **No Embarrassment**
   - Code reviewed thoroughly
   - No obvious bugs or issues
   - Follows provider patterns consistently
   - User validates functionality

---

## Questions for Review

Before implementation, please confirm:

1. **Schema Design**: Is the proxy resource schema appropriate? Any missing fields?

2. **Space Validation**: Is relying on API validation (Option A) acceptable, or do you want provider-side validation?

3. **Proxy Headers**: Simple `map[string]string` or flexible JSON? (Need to test with API)

4. **Testing**: Is the test coverage plan sufficient?

5. **Implementation Order**: Proxy → Output → Server Host makes sense?

---

**Ready for Implementation**: ⏳ Awaiting user approval

**Estimated Effort**: ~8-12 hours (systematic, careful implementation)

**Risk Level**: LOW (if we follow this plan carefully)
