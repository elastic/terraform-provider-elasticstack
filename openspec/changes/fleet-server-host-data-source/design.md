## Context

The `elasticstack_fleet_agent_policy` data source (#5369, `internal/fleet/agentpolicyds`) is the model: a selector data source on the `entitycore.NewKibanaDataSource` envelope with `Optional+Computed` selectors and an error that lists the ids of an ambiguous match. Fleet server hosts differ from agent policies in three ways that shape this design: one selector is a bool, the list API cannot filter, and hosts are not scoped to a space.

## Decisions

### 1. `default` is both a selector and an output

`default` is an `Optional+Computed` bool without a schema default. As a selector only `true` is accepted (`boolvalidator.Equals(true)`): "a host that is not the default" does not identify one host. After any successful lookup the read sets `default` from the API's `is_default`, so a lookup by `host_id` or `name` reports whether the matched host is the default one. An absent `is_default` maps to `false`.

### 2. One selector validator, one diagnostic

`ExactlyOneOf` is attached to `host_id` only and names all three selectors. The validator also runs when its own attribute is null, so one validator covers every combination and Terraform reports a single diagnostic: `No attribute specified when one (and only one) of [host_id,name,default] is required`. Attaching it to each of the three selectors produces three diagnostics, and each leaves the attribute it is attached to out of its list.

`host_id` and `name` also reject an empty string at plan time.

### 3. `name` and `default` lookups filter the complete list

`GET /api/fleet/fleet_server_hosts` takes no parameters. Kibana returns every host in one response (`perPage` 10000), so there is nothing to page or filter on the server. The provider compares `name` exactly, and `is_default` for the default lookup. The list item and the get-by-id item are the same type (`kbapi.ServerHost`), so the match is used directly, without a second request.

If `total` is greater than the number of returned items, the read fails. A name that looks unique in a truncated list could be ambiguous.

Kibana does not enforce unique Fleet server host names, so the ambiguous-name error is reachable on a real stack, and an acceptance test covers it. More than one default host is not expected from Kibana; the same error path reports it if it happens.

### 4. `space_id` does not filter

Fleet server hosts are shared by all spaces: the resource is unscoped (`IsUnscopedSpace`), and on a 9.5.5 stack the list API returns the same hosts through the default space, through another space and through a space that does not exist. `space_id` is kept because the issue asks for it and the other Fleet data sources have it. It only selects the space in the request path, which matters when the credentials are limited to one space. The attribute description says so instead of reusing the shared "identifier for the space" text.

### 5. `ssl` mirrors the resource without the key

`ssl` has `certificate`, `certificate_authorities` and `client_auth`. `key` is not in the schema, so it cannot reach state. `ssl` is null when the API returns no `ssl` object or none of the three exposed fields; `certificate_authorities` is null, not an empty list, when the API returns none.

### 6. No version gate

The data source only reads. On stacks older than the resource's `proxy_id` (8.7.1) and `ssl` (9.1.0) minimums the API does not return those fields and the attributes are null.

## Risks

- A lookup by `default = true` fails on a stack without a default Fleet server host. Deleting the default host with the resource clears the flag first, so this state is reachable. The error says that no default host was found.
