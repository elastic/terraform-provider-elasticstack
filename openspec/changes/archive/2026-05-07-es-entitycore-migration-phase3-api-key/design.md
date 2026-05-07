# Design: Migrate security_api_key to entitycore envelope

## Overview

The API key resource is a special case: it uses private provider state to cache the Elasticsearch cluster version, then gates behavior (e.g., whether the API supports updating role descriptors) based on that cached version. This makes Create/Update unsuitable for simple envelope callbacks.

## Strategy: Override Create/Read/Update

1. Pass `PlaceholderElasticsearchWriteCallbacks` to `NewElasticsearchResource`
2. Implement `Create`, `Read`, and `Update` methods on the concrete `Resource` type
3. The envelope handles `Delete`, `Schema`, `Configure`, and `Metadata`
4. The concrete type keeps `UpgradeState`

## Private Data Flow

The current flow:
1. Create/Update/Read resolve the client
2. Call `r.saveClusterVersion` to cache version in private state
3. On subsequent reads, call `r.clusterVersionOfLastRead` to retrieve cached version
4. Use cached version to decide behavior (e.g., whether to set unknown fields)

This flow stays on the concrete type. The override methods call `r.Client()` (available via embedded `ResourceBase`) to get the factory.

## Read Callback

`readAPIKey` is extracted as a package-level callback so API interaction is shared, but the concrete resource still overrides `Read`. The override duplicates the envelope prelude, delegates the Elasticsearch API interaction to `readAPIKey`, persists or removes state, and calls `saveClusterVersion` after successful reads so private cluster-version caching remains intact.

## Delete Callback

Delete is standard: parse composite ID, resolve client, call `DeleteAPIKey`. Fits envelope.

## Model

Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to `tfModel`.
