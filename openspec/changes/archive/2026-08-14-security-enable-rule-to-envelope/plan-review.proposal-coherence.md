# Proposal Coherence Review

**Change:** `security-enable-rule-to-envelope`
**Lane:** `proposal-coherence`
**Verdict:** approve

## WHY

Clear and scoped. The motivation (wrapper-level `Create`/`Update` methods shadow the envelope's promoted methods, bypassing client resolution, version enforcement, and read-after-write) is precise. The scope is correctly bounded to `security_enable_rule` only.

## WHAT

Specific and complete. Five distinct changes are enumerated with exact file paths. No user-visible behavior change is claimed and that claim is plausible (no schema or API semantics change). No breaking changes require marking.

## Capabilities vs Spec Deltas

The delta spec (`specs/kibana-kibana-resource-envelope/spec.md`) modifies the "Kibana resources fully implement envelope CRUD callbacks" requirement by:

1. Adding `security_enable_rule` to the `PlaceholderKibanaWriteCallback` SHALL NOT list. ✓ Matches capability bullet 1 and 3.
2. Updating the scenario count from "three migrated resources" to "four migrated resources". ✓ Implicitly extends the wrapper-override prohibition.

---

### SUGGESTION — proposal.md:21

**Bullet:** *"extend the 'wrapper struct SHALL NOT override Create or Update' constraint to include `security_enable_rule`"*

The "wrapper struct SHALL NOT override" sentence in the main spec (`kibana-kibana-resource-envelope/spec.md:79`) is already phrased as a general rule for all `KibanaResource` embeds — not a named-resource list. The delta spec handles this implicitly by updating the scenario from "three" to "four migrated resources"; it does not add `security_enable_rule` to a separate constraint sentence because none exists. The proposal bullet implies the constraint was previously scoped to the three named resources, which overstates the change. No substantive harm; precision only.

**Suggested fix:** Collapse bullets 2 and 3 into a single bullet: *"extend the `PlaceholderKibanaWriteCallback` SHALL NOT be used list and the lifecycle-dispatch scenario to include `security_enable_rule`"*.

---

## Redundant EnforceVersionRequirements (delete.go)

The removal of the redundant `EnforceVersionRequirements` call from `deleteSecurityEnableRule` is not reflected as a capability delta. This is correct — the existing spec requirement "Kibana resources with server-version constraints implement WithVersionRequirements" (`kibana-kibana-resource-envelope/spec.md:43`) already prohibits inline version enforcement in CRUD code paths. No new spec delta is needed.

## Summary

No CRITICAL or WARNING findings. One SUGGESTION (precision of language in capability bullet). The proposal is coherent and the delta spec accurately captures the required spec changes.
