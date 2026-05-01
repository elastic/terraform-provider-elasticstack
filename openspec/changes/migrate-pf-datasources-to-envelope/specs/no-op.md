# No Requirements Changes

This change is a pure internal refactoring. All existing data source schemas,
attributes, and read behaviors are preserved. No new capabilities are introduced
and no existing capability requirements are modified.

The migration from struct-based/`DataSourceBase` wiring to the generic envelope
is an implementation detail only; the Terraform user-visible contract remains
unchanged for every migrated data source.
