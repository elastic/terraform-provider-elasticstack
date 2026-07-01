# Tasks: Skip SLO acceptance tests on Kibana 8.10.4

## 1. Update Version Constraints

- [x] 1.1 Add `!=8.10.4` to `SLOKqlAccTestConstraints` in `internal/kibana/slo/constants.go`.
- [x] 1.2 Add `!=8.10.4` to `SLOKqlFleetAccTestConstraints` in `internal/kibana/slo/constants.go`.
- [x] 1.3 Add `!=8.10.4` to the inline `version.NewConstraint` string in `TestAccResourceSlo_36_char_slo_id`.
- [x] 1.4 Add `!=8.10.4` to the inline `version.NewConstraint` string in `TestAccResourceSloFromSDK`.

## 2. Validate Constraints

- [x] 2.1 Confirm the updated constraint strings match the requirements in `specs/kibana-slo/spec.md`.
- [x] 2.2 Run targeted SLO package validation to ensure the updated acceptance tests compile.
- [x] 2.3 Run project validation required by the implementation loop.
