Manages Kibana advanced settings for a single space, or the global advanced settings shared by all spaces. See the [Kibana advanced settings documentation](https://www.elastic.co/docs/reference/kibana/advanced-settings) for the available settings.

The resource only manages the settings declared in `settings`. Other advanced settings in the same scope are left untouched, so several resources can manage disjoint settings of the same space. Changes made outside Terraform (for example in the Kibana UI) to a managed setting are detected as drift. Destroying the resource resets the managed settings to their Kibana defaults.

Settings that are overridden in `kibana.yml` (`uiSettings.overrides`) cannot be changed through the API, and Kibana rejects updates to them. Managing advanced settings requires the `Advanced Settings` Kibana feature privilege with `all` access.

Kibana does not check that the target space exists, so create the space before its settings (for example by referencing an `elasticstack_kibana_space` resource in `space_id`).

~> **Note:** The `defaultIndex` setting is also managed by `elasticstack_kibana_default_data_view`. Manage it with only one of the two resources in a given space, otherwise each apply reverts the other resource's value.
