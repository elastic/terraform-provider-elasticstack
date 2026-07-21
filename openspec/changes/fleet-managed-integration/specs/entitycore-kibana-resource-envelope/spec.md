## ADDED Requirements

### Requirement: Write callbacks may opt out of read-after-write via `KibanaWriteResult.SkipReadAfterWrite`

The system SHALL define a boolean field `SkipReadAfterWrite` on `KibanaWriteResult[T]` returned by Kibana envelope Create and Update callbacks. The field SHALL default to `false` when unset. When `SkipReadAfterWrite` is `false`, the envelope SHALL retain the existing write path: invoke the read callback after a successful write callback and commit the read result to state (and invoke PostRead when configured). When `SkipReadAfterWrite` is `true`, the envelope SHALL persist `written.Model` directly to state without invoking the read callback and without invoking PostRead (no read occurred on the write path).

In both paths the envelope SHALL continue to apply `preserveModelTimeouts` using the plan model's `timeouts` value and SHALL set the `timeouts` attribute on state after `Set`, matching existing timeout persistence behavior.

Concrete resources that skip read-after-write SHOULD merge known server-computed values from prior state into the returned model when the plan leaves those attributes Unknown, so direct state write does not persist Unknown computed fields.

#### Scenario: SkipReadAfterWrite false — read-after-write and PostRead run as today

- **WHEN** a Create or Update write callback returns `SkipReadAfterWrite: false` (or the zero value)
- **THEN** the envelope SHALL invoke the read callback after the write callback succeeds
- **AND** when `PostRead` is configured, the envelope SHALL invoke PostRead with the read callback result before committing state

#### Scenario: SkipReadAfterWrite true — no read callback and no PostRead

- **WHEN** an Update write callback returns `SkipReadAfterWrite: true` and a model to persist
- **THEN** the envelope SHALL NOT invoke the read callback
- **AND** the envelope SHALL NOT invoke PostRead
- **AND** the envelope SHALL commit `written.Model` to state

#### Scenario: SkipReadAfterWrite true — plan timeouts still persisted

- **WHEN** an Update write callback returns `SkipReadAfterWrite: true` and a model whose embedded `timeouts` field is a zero value
- **THEN** the envelope SHALL write the plan model's `timeouts` value into state after `Set`
- **AND** the operation SHALL succeed without a `timeouts` value-conversion diagnostic
