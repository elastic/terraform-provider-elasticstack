Manages Kibana advanced settings (the settings shown under Stack Management > Advanced Settings) for one Kibana space.

Only the settings you declare in `string_settings` and `json_settings` are managed. Every other setting in the space is left untouched, whether it was changed in the Kibana UI, by another tool, or is still at its Kibana default.

## Setting values and their types

Kibana settings hold JSON values of any type, and the Kibana API does not report which type a setting expects, so the provider cannot infer it. Declare each setting in the map that matches the type Kibana stores:

- `string_settings` sends each value as a plain string. Use it for text settings, and for settings that store JSON as a string, such as `timepicker:timeDefaults`: pass the JSON text with `jsonencode(...)` and Kibana receives it unchanged as a string.
- `json_settings` sends each value as JSON, so a boolean, number, array or object keeps its type in Kibana. Pass the value as JSON text, for example `jsonencode(false)`, `jsonencode(0)` or `jsonencode(["@timestamp", "message"])`. A value that Kibana returns with different whitespace, object key order or number formatting than you configured does not show as drift. Reformatting a value in your own configuration can show as an in-place update that changes nothing in Kibana.

Terraform converts booleans and numbers to strings when assigning them to a map of strings, so a boolean or number setting such as `csv:quoteValues` belongs in `json_settings`, not in `string_settings`.

The provider does not keep a list of known settings. Whether a key and its value are accepted is decided by Kibana, so keys that Kibana does not register are passed through as they are.

## Changing and removing settings

Applying a change sends Kibana only the settings that are new or whose value changed. Removing a key from the configuration, so that it is declared in neither map (`string_settings` or `json_settings`), resets that setting to its Kibana default: the value you set is removed, and Kibana falls back to the default it ships with. Settings that were never declared are not affected. When moving a key from one map to the other, the setting is stored with the type of its new map and is not reset.

## Destroying the resource

Destroy resets only the managed keys: every setting declared in `string_settings` or `json_settings` returns to its Kibana default, and every other setting in the space is left unchanged. A resource that declares no settings changes nothing when it is destroyed. If the space was deleted outside Terraform, destroy succeeds without changing any setting.

## Spaces

Advanced settings belong to a Kibana space. `space_id` defaults to `default`, and changing it replaces the resource. A space other than `default` must already exist when the resource is created or updated, for example by creating it with `elasticstack_kibana_space`; otherwise the operation fails and nothing is written. If the space is deleted outside Terraform, the next refresh removes the resource from state. Settings can outlive a deleted space: Kibana may keep serving the settings previously written to it, so recreating a space with the same ID brings its earlier settings back.

## Permissions

Creating, updating and destroying the resource write settings, which requires the Kibana privilege `manage_advanced_settings`. Without it Kibana answers 403, for example `API [POST /api/kibana/settings] is unauthorized for user, this action is granted by the Kibana privileges [manage_advanced_settings]`. Reading settings does not require the privilege.

For a space other than `default`, the provider first looks the space up through the Kibana spaces API before it reads or writes settings, so the credentials must be able to read that space. If the lookup fails for any reason, for example because Kibana answers 403 when the credentials have no access to the space, the operation fails with that error instead of being treated as a missing space.

## Errors

When Kibana rejects a request or a request fails, the operation fails with an error that includes Kibana's message. Kibana rejects, for example, an invalid value for a registered setting, a setting that the Kibana server configuration pins to a fixed value (`Unable to update "<key>" because it is overridden`), and a request without the required privilege.

A failed operation records nothing: a failed create leaves the resource out of the Terraform state, and a failed update keeps the values recorded before the update. The next refresh reconciles the state with whatever Kibana holds.

When the failed request sent `string_settings` entries, the error also advises declaring settings with non-string values in `json_settings`. Only entries that were actually sent count: an update sends only new or changed keys, so a declared `string_settings` entry that did not change does not trigger the advice, and neither does a request that sent only `json_settings` entries or only reset removed keys.

The advice is added to every error that Kibana or the connection returns for such a request, including errors unrelated to types, such as a missing privilege. It is not added to the error for a missing settings route.

Kibana does not always reject a string for a typed setting: it accepted the string `"false"` for the boolean `csv:quoteValues` and the string `"5"` for the number `histogram:barTarget`, and stored both as strings. Declare typed settings in `json_settings` so that Kibana stores the type you intend.

If Kibana answers 404 for the settings route itself, for example on a deployment that does not serve it, the operation fails with an error naming the route (`GET /api/kibana/settings` for a refresh, `POST /api/kibana/settings` for a write). A refresh never reads that response as the space having no settings, so managed settings are not dropped from the state by mistake.

## Drift detection

On every refresh the provider reads each managed setting back from Kibana. A setting changed or reset outside Terraform appears in the next plan and is corrected by the next apply, and settings you do not declare are never compared.

A `string_settings` key whose value in Kibana is no longer a string is recorded as the JSON text of that value, for example `5`, so it shows as a difference from the configured string. Because of this, a string `"true"` and a boolean `true` are indistinguishable in drift detection: a `string_settings` key declared as `"true"` shows no difference if its value in Kibana becomes the boolean `true`.

Kibana returns integers beyond the JavaScript safe range (2^53) rounded, so such a number declared in `json_settings` may show as drift.

## Overlapping managers

Two `elasticstack_kibana_advanced_settings` resources that declare the same key in the same space overwrite each other, as does any other tool that writes that setting. For example, `elasticstack_kibana_default_data_view` writes the `defaultIndex` setting, so do not declare `defaultIndex` here for a space where that resource is used.

## Sensitive values

Setting values are stored in the Terraform state in plain text and are not marked sensitive, so they must not contain secrets.

## Kibana routes and supported versions

Kibana serves advanced settings through internal Kibana routes (`GET` and `POST /api/kibana/settings`, under `/s/<space_id>` for a space other than `default`) that have no public API contract, so Elastic could change or remove them in any release. Kibana 9.x serves them only to requests that carry the `x-elastic-internal-origin: Kibana` header, which the provider sends.

The resource's behavior was verified against current stateful Kibana (9.5.6). Older Kibana versions and Elastic Cloud Serverless are not verified: the route source exists in every Kibana release that was checked, but its behavior was not tested there.
