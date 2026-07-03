 Implemented top-level task 4 of the `selective-acceptance-tests` OpenSpec change.

Changed files:
- `.github/workflows/provider.yml`

Validation:
- `python3 -c 'import yaml; yaml.safe_load(open(".github/workflows/provider.yml")); print("YAML parses successfully")'` → `YAML parses successfully`

Commits created:
- `b213776bf ci(provider): add merge_group trigger and targeted acceptance test gating`

Open risks/questions:
- The `force-install-synthetics` step references `steps.get-api-key.outputs.apikey`; if it were to run while the API key step was skipped, that would fail. However, it is now gated on `has_packages == 'true'`, the same as `get-api-key`, so the API key step will always run whenever force-install-synthetics runs.

Recommended next step:
- Verify on an actual PR/merge-queue run that `scripts/targeted-testacc/...` exists and behaves as expected; the workflow change depends on that tool being present (tasks 1–3).