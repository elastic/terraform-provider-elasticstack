# Coding Standards

This document outlines the coding standards and conventions used in the terraform-provider-elasticstack repository.

## General Principles

- Write idiomatic Go.
  - [Effective Go](https://go.dev/doc/effective_go)
  - [Code Review Comments](https://go.dev/wiki/CodeReviewComments)
  - The [Google Styleguide](https://google.github.io/styleguide/go/index#about) 

## Project Structure

- Use the Plugin Framework for all new resources (not SDKv2)
- Follow the code organization pattern of [the `system_user` resource](./internal/elasticsearch/security/system_user) for new Plugin Framework resources
  - [`testdata/`](./internal/elasticsearch/security/system_user/testdata) - This directory contains Terraform definitions used within the resource acceptance tests. In most cases, this will contain a subdirectory for each test, which then contain subdirectories for individual named test steps. 
  - [`acc_test.go`](./internal/elasticsearch/security/system_user/acc_test.go) - Contains acceptance tests for the resource
  - [`create.go`](./internal/elasticsearch/security/system_user/create.go) - Contains the resources `Create` method and any required logic. Depending on the underlying API, the create and update handlers may share a single code path. 
  - [`delete.go`](./internal/elasticsearch/security/system_user/delete.go) - Contains the resources `Delete` method.
  - [`models.go`](./internal/elasticsearch/security/system_user/models.go) - Contains Golang models used by the resource. At a minimum this will contain a model for reading plan/config/state from the Terraform plugin framework. Any non-trivial models should also define receivers for translating between Terraform models and API client models. 
  - [`read.go`](./internal/elasticsearch/security/system_user/read.go) - Contains the resources `Read` method. This should also define an internal `read` function that can be re-used by the create/update paths to populate the final Terraform state after performing the create/update operation. 
  - [`resource.go`](./internal/elasticsearch/security/system_user/resource.go) - Contains:
    - A factory function for creating the resource (e.g `NewSystemUserResource`)
    - `Metadata`, `Configure`, and optionally `ImportState` functions. 
    - Type assertions ensuring the resource fully implement the relevant Plugin Framework interfaces (e.g `var _ resource.ResourceWithConfigure = &systemUserResource{}`)
  - [`schema.go`](./internal/elasticsearch/security/system_user/schema.go) - Contains the `Schema` function fully defining the resources schema
  - [`update.go`](./internal/elasticsearch/security/system_user/update.go) - Contains the `Update` method. Depending on the underlying API this may share significant logic with the `Create` method. 
  - Some resources may define other files, for example:
    - [`models_*.go`](./internal/kibana/security_detection_rule/) - Complex APIs may result in significant model related logic. Split these files as appropriate if they become large. 
    - Custom [plan modifiers](./internal/elasticsearch/security/api_key/set_unknown_if_access_has_changes.go), [validators](./internal/elasticsearch/security/api_key/validators.go) and [types](./internal/elasticsearch/security/api_key/role_descriptor_defaults.go) - Resource specific plan modifiers and custom types should be contained within the resource package. 
    - [`state_upgrade.go`](./internal/elasticsearch/security/api_key/state_upgrade.go) - Resources requiring state upgrades should place the `UpgradeState` method within this file.
- Avoid adding extra functionality to the existing `utils` package. Instead:
  - Code should live as close to the consumers.
  - Resource, area, application specific shared logic should live at that level. For example within `internal/kibana` for Kibana specific shared logic.
  - Provider wide shared logic should be packaged together by a logical concept. For example [diagutil](./internal/diagutil) contains shared code for managing Terraform Diagnostics, and translating between errors, SDKv2 diags, and Plugin Framework diags.
- Prefer using existing util functions over longer form, duplicated code:
  - `utils.IsKnown(val)` instead of `!val.IsNull() && !val.IsUnknown()`
  - `utils.ListTypeAs` instead of `val.ElementsAs` or similar for other collection types
  - `typeutils.StringishValue` instead of casting to a string eg `types.StringValue(string(apiResp.Id))`. Use `typeutils.StringishPointerValue` for pointers
- The final state for a resource should be derived from a read request following a mutative request (eg create or update). We should not use the response from a mutative request to build the final resource state.

## Schema Definitions
- Use custom types to model attribute specific behaviour.
    - Use [`jsontypes.NormalizedType{}`](https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/main/jsontypes/normalized_type.go) custom type for string attributes containing JSON blobs.
    - Use [`customtypes.DurationType{}`](./internal/utils/customtypes/duration_type.go) for duration-based string attributes.
    - Use [`customtypes.JSONWithDefaultsType{}`](./internal/utils/customtypes/json_with_defaults_type.go) to allow users to specify only a subset of a JSON blob.
- Always include comprehensive descriptions for all resources, and attributes. 
- Long, multiline descriptions should be stored in an external markdown file, which is imported via Golang embedding. For [example](./internal/elasticsearch/security/system_user/resource-description.md).
- Use schema validation wherever possible. Only perform validation within create/read/update functions as a last resort. 
  - For example, any validation that relies on the actual Elastic Stack components (e.g Elasticsearch version)
    can only be performed during the create/read/update phase. 
- Kibana and Fleet resources will be backed by the Kibana API. The schema definition should closely follow the defined API request/response models defined in the [OpenAPI specification](./generated/kbapi/oas-filtered.yaml).
  - Further details may be found in the [API documentation](https://www.elastic.co/docs/api/doc/kibana/v9/)
- Elasticsearch resources will be backed by the [go-elasticsearch](https://github.com/elastic/go-elasticsearch) client. 
  - Further details may be found in the [API documentation](https://www.elastic.co/docs/api/doc/elasticsearch/)
- Use `EnforceMinVersion` to ensure the backing Elastic Stack applications support the defined fields. 
  - The provider supports a wide range of Stack versions, and so newer features will not be available in all versions. 
  - See [`assertKafkaSupport`](./internal/fleet/output/models.go) for an example of how to handle the use of unsupported attributes.


## JSON Handling

- Use [`jsontypes.NormalizedType{}`](https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/main/jsontypes/normalized_type.go) for JSON string attributes to ensure proper normalization and comparison.
- Use [`customtypes.JSONWithDefaultsType{}`](./internal/utils/customtypes/json_with_defaults_type.go) if API level defaults may be applied automatically. 

## Testing

- Use table-driven unit tests when possible with `t.Run()` for test cases
- Use testify library (`assert`, `require`) for test assertions
- Ensure that *every* resource attribute is covered by at least one acceptance test case whenever possible.
  - Features that *require* external services are likely the only excuse to not include acceptance test coverage. 
- Organize acceptance tests in `acc_test.go` files
- Test Terraform code should be vanilla, valid Terraform
  - Store test Terraform modules in `testdata/<test_name>/<step_description>` directories. 
  - Define any required variables within the module
  - Reference the test code via `ConfigDirectory: acctest.NamedTestCaseDirectory("<step description>")`
  - Define any required variables via `ConfigVariables`
- Resources should include tests for the following
  - Creating a resource
  - Updating a resource
  - Deleting a resource
  - Importing a resource
  - Creating a resoure in another space (if applicable)

## API Client Usage

- Use generated API clients from [`generated/kbapi/`](./generated/kbapi/) for new Kibana API interactions
- Avoid deprecated clients (`libs/go-kibana-rest`, `generated/alerting`, `generated/connectors`, `generated/slo`)