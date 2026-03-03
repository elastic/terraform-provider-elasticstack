# Reference: heuristics for schema ↔ test matching

## Schema extraction (Terraform Plugin SDK)

### Common shapes to follow

- Top-level:
  - `Schema: map[string]*schema.Schema{ "attr": { ... } }`
- Nested blocks:
  - `TypeList` / `TypeSet` with `Elem: &schema.Resource{ Schema: ... }`
  - `TypeMap` with `Elem: &schema.Schema{ ... }` (map of primitives) or `Elem: &schema.Resource{ ... }` (map of objects)
- Delegated schema:
  - `Schema: mySchema()` or `Schema: schemaForX()` → follow the helper until you reach the concrete map keys.

### What to record per attribute

- `Required` / `Optional` / `Computed`
- `ForceNew` (often implies special update behavior)
- `Type*` + `MinItems/MaxItems` (block cardinality)
- For collections: whether elements are primitives vs nested objects (`Elem`)

## Schema extraction (Terraform Plugin Framework)

### Where schema is defined

In the Framework, schemas typically live in methods like:
- `Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse)`
- `Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse)`

Look for assignments like:
- `resp.Schema = schema.Schema{ Attributes: ..., Blocks: ... }`

### What to extract

Capture **all attribute and block keys**, including nested object attributes.

1) **Top-level attributes** under:
- `schema.Schema{ Attributes: map[string]schema.Attribute{ ... } }`

2) **Blocks** under:
- `schema.Schema{ Blocks: map[string]schema.Block{ ... } }`
  - `schema.ListNestedBlock{ NestedObject: schema.NestedBlockObject{ Attributes: ..., Blocks: ... } }`
  - `schema.SetNestedBlock{ ... }`
  - `schema.SingleNestedBlock{ ... }`

3) **Nested object attributes** under:
- `schema.SingleNestedAttribute{ Attributes: map[string]schema.Attribute{ ... } }`
- `schema.ListNestedAttribute{ NestedObject: schema.NestedAttributeObject{ Attributes: ... } }`
- `schema.SetNestedAttribute{ ... }`
- `schema.MapNestedAttribute{ ... }`

Also record special collection element types that impact pathing/assertions:
- `schema.ListAttribute{ ElementType: ... }`, `schema.SetAttribute{ ... }`, `schema.MapAttribute{ ... }`

### Required/Optional/Computed mapping

Framework attributes encode this via boolean fields:
- `Required: true`
- `Optional: true`
- `Computed: true`

Record whichever apply; note that some attributes can be `Optional: true` and `Computed: true` (common for “Optional with default / server-populated” patterns).

### Plan modifiers / update semantics (best-effort)

Framework doesn’t use `ForceNew`; instead look for plan modifiers that imply replacement or state retention. When present, record them as hints for “update coverage” focus:
- `resource.RequiresReplace()` / `planmodifier.RequiresReplace()` (or similarly named helpers)
- `UseStateForUnknown()` (state retention; affects update/unknown behavior)

Because plan modifiers can be assembled via helper functions, follow references if the schema is delegated.

### Path normalization notes (Framework)

Acceptance tests still assert against state paths (strings), but Framework resources often represent:
- **Single nested** blocks/attributes as `block.attr` in state paths (no index)
- **List/Set nested** blocks/attributes as `block.0.attr` (or `block.#` / `block.*` patterns)
- **Maps** as `labels.%` / `labels.key`

Treat nested object attributes similarly to SDK blocks for matching purposes:
- `single_nested_attribute.child` behaves like `block.child`
- `list_nested_attribute.0.child` behaves like `block.0.child`

## Acceptance test extraction

### Where attribute usage appears

1) In HCL configs:
- `resource.TestStep{ Config: <string or function call that returns string> }`
- Config builder helpers like `testAcc<Resource>Config...(...)`

2) In assertions/checks:
- `resource.TestCheckResourceAttr(resourceName, "path", "value")`
- `resource.TestCheckResourceAttrSet(resourceName, "path")` (set-only)
- `resource.TestCheckNoResourceAttr(resourceName, "path")` (absence)
- `resource.TestMatchResourceAttr(resourceName, "path", regexp.MustCompile(...))`
- Sets/lists:
  - `resource.TestCheckTypeSetElemAttr(...)`
  - `resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block.*", map[string]string{...})`
  - length checks often appear as `"block.#"` / `"tags.%"`

### Update coverage detection

Treat an attribute as having update coverage only if:
- The test has **multiple steps** applying different configs for the same resource, AND
- The attribute’s value meaningfully differs between steps, AND
- A post-update check asserts the new value (prefer exact match over “set”).

## Matching attribute paths (normalization)

Acceptance tests use state paths like:
- Top-level: `name`
- List/set block: `block.0.attr` (index-based) or `block.#` (count)
- Map: `labels.%` (count) or `labels.key`

Normalize schema block children to a conceptual “any index” form for matching:
- Schema: `block` with nested `attr`
- Test: `block.0.attr` → match as `block[*].attr`

For sets where ordering is unstable, prefer set-element helpers (`TypeSetElem...`) rather than hard-coded indices.

## Coverage quality classification

### No coverage
Attribute path is absent from:
- all configs AND
- all checks

### Poor coverage (include explicit reasons)

Flag an attribute as poor coverage if any apply:
- **Configured but never asserted** (no state verification)
- **Set-only assertion** used where a deterministic value assertion is feasible
- **Single value only**: only one distinct value ever used/asserted
- **Optional never unset**: no test step omits it, and no absence/default behavior is asserted
- **Collections never empty**: no step tests empty/omitted collection + asserts empty state
- **No update coverage**: never changes across steps (or changes but is not asserted post-update)

## Report writing checklist

- Use the report template from `SKILL.md`
- Always list “no coverage” before “poor coverage”
- Prefer concrete suggestions:
  - “Add a second `resource.TestStep` that removes `<attr>` and assert with `TestCheckNoResourceAttr`”
  - “Replace `TestCheckResourceAttrSet` with `TestCheckResourceAttr` for `<attr>` using a deterministic expected value”
  - “Add an empty collection case and assert `.<collection>.#` or `.<map>.%` is `0`”
