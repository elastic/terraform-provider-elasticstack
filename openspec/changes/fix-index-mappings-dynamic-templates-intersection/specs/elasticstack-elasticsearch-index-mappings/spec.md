## MODIFIED Requirements

### Requirement: Read — user-declared subset only (REQ-004)

On read, the resource SHALL retrieve the index metadata via the existing `GetIndex` helper (which returns the `Mappings` payload from the index state) and reconstruct the stored `mappings` by retaining **only the top-level keys that are present in the previously stored state**. Top-level keys that exist in the API response but are absent from the stored state SHALL be silently discarded.

For the `properties` top-level key, the filtering SHALL be **recursive**: only field names that appear in the previously stored `properties` tree SHALL be retained at every nesting level. Dynamically-added fields within `properties` that are absent from the stored state SHALL be silently discarded.

For the `dynamic_templates` top-level key, the filtering SHALL be **name-keyed**, mirroring `properties`: only dynamic template names that appear in the previously stored `dynamic_templates` array SHALL be retained, using the API's definition for each retained name. Template entries contributed by an index template (or any other out-of-band source) that are absent from the stored state SHALL be silently discarded. The retained entries SHALL preserve the **API response's relative order of declared names**, dropping undeclared extras, because dynamic template order is semantically significant to Elasticsearch (first match wins). A live reorder of declared templates is thereby written into state. A declared template name that is absent from the API response SHALL be dropped from state rather than retained with a stale value, whether that name is missing from an otherwise-present `dynamic_templates` array or the API response omits the `dynamic_templates` key entirely (for example, because every declared template was removed out-of-band). When no declared names remain, the resource SHALL persist `dynamic_templates` as an empty array rather than omitting the key, so Framework semantic equality does not re-pin the previously stored array (an omitted key is indistinguishable from never having declared `dynamic_templates`). This allows drift to surface on the next plan rather than following the general rule of retaining the stored value for a top-level key absent from the API response. If either the stored state's or the API response's `dynamic_templates` value cannot be parsed into named templates (a duplicate template name within the array, or a template entry whose value is not an object — reachable only through manual state edits, since Elasticsearch itself does not produce this shape), the resource SHALL fall back to storing the API's `dynamic_templates` value unfiltered for that key, unchanged from its behavior for other unrecognized top-level keys.

Because `Read` already drops undeclared extras, plan/apply semantic equality for this resource's `dynamic_templates` SHALL require **matching name sets** (not extras-tolerant subset matching used by the index and template resources, which store the full API mappings) in addition to equivalent definitions and relative order. A declared name absent from the API, or a config name absent from state, SHALL surface as drift. An empty declared `dynamic_templates` array SHALL NOT be treated as a subset of a nonempty array.

If the previously stored `mappings` is empty (e.g. immediately after `terraform import` via `ImportStatePassthroughID`), the resource SHALL store the full API response as the initial mask. This allows users to narrow the declaration in subsequent configuration changes.

The resource SHALL use `index.MappingsType{ExactDynamicTemplateNames: true}` so that equivalent JSON representations (different key ordering, different whitespace) do not produce a spurious diff, while `dynamic_templates` name sets are compared exactly.

#### Scenario: Dynamic extras do not cause drift

- GIVEN a resource that declares only `properties` with two explicit fields (`title` and `body`)
- AND Elasticsearch adds a dynamic field `tags` to the index (e.g. via auto-mapping a document)
- WHEN `terraform plan` runs after the dynamic field is added
- THEN the plan SHALL show no diff for the `mappings` attribute
- AND the stored state SHALL continue to contain only `title` and `body` under `properties`

#### Scenario: Not found on read removes from state

- GIVEN the target index is deleted outside Terraform
- WHEN `terraform refresh` or `terraform plan` runs
- THEN the resource SHALL be removed from state
- AND Terraform SHALL propose recreating it on the next apply

#### Scenario: Index-template-injected dynamic templates do not persist into state

- GIVEN a resource that declares one `dynamic_templates` entry named `text_ja_example`
- AND an index template contributes an additional entry named `template_default` to the index's mappings
- WHEN `terraform plan` or `terraform refresh` runs after the index template's entry appears in the API response
- THEN the stored `dynamic_templates` state SHALL contain only the `text_ja_example` entry
- AND `template_default` SHALL NOT appear anywhere in the stored `mappings` state
- AND the plan SHALL show no diff for the `mappings` attribute

#### Scenario: Declared dynamic template removed out-of-band is dropped from state

- GIVEN a resource that declares two `dynamic_templates` entries, `alpha` and `beta`
- AND `beta` is removed from the index's mappings out-of-band (outside Terraform)
- WHEN `terraform plan` or `terraform refresh` runs
- THEN the stored `dynamic_templates` state SHALL contain only `alpha`
- AND the next `terraform plan` SHALL show a diff proposing to restore `beta` (drift surfaced, not silently masked)

#### Scenario: Unparseable dynamic_templates shape falls back to passthrough

- GIVEN a resource whose stored state `dynamic_templates` array contains two entries sharing the same template name (only reachable via a manually edited state file)
- WHEN `terraform plan` or `terraform refresh` runs
- THEN the resource SHALL store the API's `dynamic_templates` value for that key unfiltered
- AND SHALL NOT error or drop the `dynamic_templates` key entirely

#### Scenario: All declared dynamic templates removed out-of-band leaves no declared names in state

- GIVEN a resource that declares one `dynamic_templates` entry named `alpha`
- AND `alpha` is removed from the index's mappings out-of-band, leaving the API response with no `dynamic_templates` key at all
- WHEN `terraform plan` or `terraform refresh` runs
- THEN the stored `dynamic_templates` state SHALL be an empty array
- AND SHALL NOT retain the previously stored `alpha` entry
- AND the next `terraform plan` SHALL show a diff proposing to restore `alpha` (drift surfaced, not silently pinned)

#### Scenario: Reordering declared dynamic templates surfaces as drift

- GIVEN a resource that declares two `dynamic_templates` entries in the order `alpha`, then `beta`
- AND Elasticsearch's index mappings report the same two templates with equivalent definitions but in the order `beta`, then `alpha` (a live reorder, not an index-template-injected extra)
- WHEN `terraform plan` runs
- THEN the stored `dynamic_templates` state SHALL reflect the API order `beta`, then `alpha`
- AND the plan SHALL show a diff for the `mappings` attribute reflecting the changed order
- AND the diff SHALL NOT be suppressed by semantic equality
