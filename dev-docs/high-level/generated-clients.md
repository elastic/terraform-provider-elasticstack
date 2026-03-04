# Generated clients

This repo includes generated API clients.

## Kibana OpenAPI client (`generated/kbapi`)

- Location: `generated/kbapi`
- Canonical detailed doc: [`generated/kbapi/README.md`](../../generated/kbapi/README.md)
- Regenerate the Go client : `make -C generated/kbapi all`

When adding new Kibana endpoints, prefer using the `generated/kbapi` client (see “API Client Usage” in [`coding-standards.md`](./coding-standards.md)).

## Deprecated clients

These exist but *must* be avoided for new work:

- `libs/go-kibana-rest`
- `generated/slo`

See “Working with Generated API Clients” in [`contributing.md`](./contributing.md).

