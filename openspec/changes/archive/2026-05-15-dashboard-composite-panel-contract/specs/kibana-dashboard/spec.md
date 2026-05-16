## ADDED Requirements

### Requirement: Composite Lens panel handlers and `vis_config` naming (REQ-046)

For composite Lens-backed dashboard panel types, the `elasticstack_kibana_dashboard` resource SHALL use dedicated typed handler implementations while preserving the behaviors already defined for those panel types, except where this requirement explicitly changes the Terraform schema name.

For `type = "vis"` panels, the Terraform typed configuration block SHALL be named `vis_config`. The provider SHALL treat `vis_config` as the typed configuration entry point for `vis` panels and SHALL align routing and validation with the panel type discriminator `"vis"`.

For the Lens dashboard app panel type, the Terraform typed configuration block SHALL remain `lens_dashboard_app_config` and SHALL continue to support its defined by-value and by-reference behavior.

The handler architecture for these composite panel types SHALL consume the shared Lens converter registry for by-value Lens chart handling and shared by-reference conversion logic for by-reference handling.

Except for the `viz_config` to `vis_config` rename, this architectural change SHALL preserve the previously defined Terraform-facing behavior for `vis` and `lens_dashboard_app` panels, including read/write semantics, validation, null-preservation, typed by-value chart handling, and by-reference handling.

#### Scenario: `vis` panels use `vis_config` as the typed block name

- GIVEN a dashboard panel with `type = "vis"`
- WHEN Terraform validates or the provider processes the typed panel configuration
- THEN the typed configuration block name SHALL be `vis_config`
- AND routing and validation for that panel SHALL use the `"vis"` panel type contract

#### Scenario: Composite by-value Lens chart handling uses shared registry dispatch

- GIVEN a `vis` or `lens_dashboard_app` panel configured with a supported typed by-value Lens chart block
- WHEN the provider reads or writes that panel
- THEN the provider SHALL dispatch the by-value Lens chart conversion through the shared Lens converter registry

#### Scenario: Composite by-reference handling remains behaviorally unchanged

- GIVEN a `vis` or `lens_dashboard_app` panel configured in by-reference mode
- WHEN the provider plans, applies, reads, and refreshes that panel after the composite handler migration
- THEN the panel SHALL preserve the same by-reference behavior already defined for that panel type

#### Scenario: `viz_config` is no longer the accepted typed block name

- GIVEN a Terraform configuration that uses `viz_config` on a `type = "vis"` panel
- WHEN Terraform validates the configuration against the updated schema
- THEN the configuration SHALL be rejected because the typed block name is `vis_config`
