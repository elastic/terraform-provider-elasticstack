## REMOVED Requirements

### Requirement: Embedded resource core constructs provider resource type names from typed namespace parts

**Reason:** The shared Plugin Framework substrate is broadened from "resource core" to "entity core" so it can also cover Plugin Framework data sources. The capability is renamed to `provider-framework-entity-core`, which restates this requirement against the renamed substrate type `entitycore.ResourceBase` and adds a parallel form for `entitycore.DataSourceBase`.

**Migration:** Read the equivalent requirement in `provider-framework-entity-core` ("Embedded entity core constructs Terraform type names from typed namespace parts"). The Terraform type-name format `<provider_type_name>_<component>_<resource_name>` is preserved exactly; only the substrate type name (`Core` → `ResourceBase`) and constructor (`New` → `NewResourceBase`) change.

### Requirement: Embedded resource core provides canonical provider client-factory wiring

**Reason:** Renamed to `provider-framework-entity-core` and broadened to cover both `ResourceBase` and `DataSourceBase` with identical Configure/diagnostics/`Client()` semantics.

**Migration:** Read the equivalent requirement in `provider-framework-entity-core` ("Embedded entity core provides canonical provider client-factory wiring"). The Configure diagnostics rule, the nil-`ProviderData` handling, and the `Client()` accessor contract are preserved.

### Requirement: Embedded resource core does not define import behavior

**Reason:** Renamed to `provider-framework-entity-core` and broadened. The new requirement ("Embedded entity core does not define entity-kind-specific behavior") covers both `ResourceBase` (no `ImportState`, no Schema/CRUD/UpgradeState/ValidateConfig/ConfigValidators/ModifyPlan) and `DataSourceBase` (no Schema/Read/ConfigValidators/ValidateConfig).

**Migration:** Read the equivalent requirement in `provider-framework-entity-core`. The "concrete resource owns import behavior" rule is preserved verbatim.

### Requirement: Compatible Plugin Framework resources use the shared resource core for bootstrap wiring

**Reason:** Renamed to `provider-framework-entity-core`. The compatibility rule for Plugin Framework resources is restated against `*entitycore.ResourceBase`, and a parallel rule is added for compatible Plugin Framework data sources embedding `*entitycore.DataSourceBase`.

**Migration:** Read the equivalent requirement in `provider-framework-entity-core` ("Compatible Plugin Framework resources use ResourceBase for bootstrap wiring"). Every existing migrated resource keeps its current `Component` choice and literal resource-name suffix; only the embedded type spelling changes from `*resourcecore.Core` to `*entitycore.ResourceBase` and the constructor from `resourcecore.New(...)` to `entitycore.NewResourceBase(...)`.
