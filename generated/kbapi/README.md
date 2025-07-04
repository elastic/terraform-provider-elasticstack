# Generated Kibana API client

This package includes an API client generated from the Kibana OpenAPI spec.
This client is very much a work in progress, there's opportunity for improvement both within this provider, but also within the OpenAPI spec created during the Kibana build process.

This readme serves to document the status quo, contributions improving this current process are very much welcome.

## Adding new endpoints

There's more detail on the full process below, this section aims to provide a tl;dr to what's required to cover new endpoints in the generated client.

1. `make transform`. This ensures you're starting with the latest API spec. If you encounter issues at this point, at least you know it's not related to any changes you've made :)
1. Add the required API paths to the allow list. See `transformFilterPaths` in `transform_schema.go`. The definition here is straightforward, the path must match exactly what's in the API spec.
1. If endpoints don't use any `oneOf` fields then you're likely done. Run `make transform generate` and start work on the provider resource. Sweeping statements like this tend to be wrong at times, if you encounter issues with the generated client then you'll have to investigate them and may need to define additional transforms.
1. If the endpoints use `oneOf` in the request/response bodies then you'll need to define additional transforms to extract those to reusable components. See `transformKibanaPaths` or `transformFleetPaths` for some examples.
1. `make transform generate` and go test the fresh client.

## Client generation

The actual final Go client generation is relatively straightforward. `make generate` executes [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) which creates `kibana.gen.go` from `oas-filtered.yaml`.

## OpenAPI spec transforms.

This is where the dragons lie.

`make transform` downloads the schema in kibana/main, and passes it through `transform_schema.go`. This transform step aims to patch over:
* Issues with the Kibana schema, any issues we encounter should also be raised in the Kibana repo.
* Issues with oapi-codegen, similarly we should ensure there's a corresponding issue in the oapi-codegen repo.

At a high level, `transform_schema.go`:
* Filters out API paths no in the allow list (see `transformFilterPaths`).
* Removes any `kbnXsrf` parameters. We add this heaer in the client by default. We should look to remove this transform to simplify the overall process.
* Removes the `elastic-api-version` header. We only have a single version at the moment, that said we should remove this transform and properly define this logic within the client.
* Transforms any versioned content types (e.g `application/json; Elastic-Api-Version=2023-10-31`) into plain `application/json`. Similarly to above, the TF provider should properly handle these content types moving forward.
* Ensures Kibana api paths have a `spaceId` parameter (see `transformKibanaPaths`). We should ensure Kibana is correctly including this parameter in any cases where it's required. Not defining this parameter should be seen as a bug in spec generation, not something for clients to work around.
* Extracts nested type definitions to a shared component (see `transformKibanaPaths`). We've encountered issues in oapi-codegen (https://github.com/oapi-codegen/oapi-codegen/issues/1900) where nested, polymorphic types (`oneOf`) create a syntactically invalid API client. To workaround this, `transform_schema` pulls those nested types into re-usable components, and then updated the schema to reference those new types.
    * Ultimately the spec is valid here, it would be ideal for oapi-codegen to handle this correctly. That said, this transformation is the most tedious, and complicated in this stack. If we continue to encounter this issue long term we should look to see if it's possible for the Kibana spec generation to create these re-usable types directly.
* Makes several endpoint specific request/response body updates. There are several bugs within some request/response body definitions. These have all been raised as issues in the Kibana repo, in the meantime the transformer is patching the schema to create a working client.

## Possible improvements

These should likely be issues in this repo...

1. Avoid the need for this transformation stuff entirely? It doesn't seem unreasonable to expect the spec provided by Kibana to 'work'. IMO that includes avoiding long standing issues with well used client generators if possible. Anything we can fix within the Kibana repo should be fixed there in preference to adding new transformation code.
1. Migrate the bespoke `transform_schema` to [redocly](https://redocly.com/docs/cli) or oapi-codegen [overlays](https://github.com/oapi-codegen/oapi-codegen?tab=readme-ov-file#modifying-the-input-openapi-specification-with-openapi-overlay). We're reinventing the wheel here. Specifically with the `$ref` issue, if this requires local transformation long term then we should blanket apply this transform to all request/response bodies. That's likely easier done with a well tested tool like redocly than it is with homegrown tooling.
1. Include all paths by default. The barrier here is the high level of transformation required to produce a working client. That's likely reduced substantially since `transform_schema` was introduced and it's likely worth investigation.
1. Remove pointless transforms. Why are we removing examples?
