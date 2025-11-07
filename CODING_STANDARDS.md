# Coding Standards

This document outlines the coding standards and conventions used in the terraform-provider-elasticstack repository.

## General Principles

- Write idiomatic Go.
  - [Effective Go](https://go.dev/doc/effective_go)
  - [Code Review Comments](https://go.dev/wiki/CodeReviewComments)
  - The [Google Styleguide](https://google.github.io/styleguide/go/index#about) 

## Project Structure

- Use the Plugin Framework for all new resources (not SDKv2)
- Follow the code organization pattern of `internal/elasticsearch/security/system_user` for new Plugin Framework resources
- Avoid adding extra functionality to the existing `utils` package. Instead:
  - Code should live as close to the consumers.
  - Resource, area, application specific shared logic should live at that level. For example within `internal/kibana` for Kibana specific shared logic.
  - Provider wide shared logic should be packaged together by a logical concept. For example `internal/diagutil` contains shared code for managing Terraform Diagnostics, and translating between errors, SDKv2 diags, and Plugin Framework diags.

## Schema Definitions

- Use custom types to model attribute specific behaviour.
    - Use `jsontypes.NormalizedType{}` custom type for string attributes containing JSON blobs.
    - Use `customtypes.DurationType{}` for duration-based string attributes.
    - Use `customtypes.JSONWithDefaultsType{}` to allow users to specify only a subset of a JSON blob.
- Always include comprehensive descriptions for all resources, and attributes. 
- Long, multiline descriptions should be stored in an external markdown file, which is imported via Golang embedding. See `internal/elasticsearch/security/system_user/resource-description.md` for an example location.
- Use schema validation wherever possible. Only perform validation within create/read/update functions as a last resort. 
  - For example, any validation that relies on the actual Elastic Stack components (e.g Elasticsearch version)
    can only be performed during the create/read/update phase. 

## JSON Handling

- Use `jsontypes.NormalizedType{}` for JSON string attributes to ensure proper normalization and comparison.
- Use `customtypes.JSONWithDefaultsType{}` if API level defaults may be applied automatically. 

## Resource Implementation

- Follow the pattern: `resource.go`, `schema.go`, `models.go`, `create.go`, `read.go`, `update.go`, `delete.go`
- Use factory functions like `NewSystemUserResource()` to create resource instances
- Ensure appropriate interface assertions are included alongside the resource definition. 
  - For example, if a resource supports imports, include `var _ resource.ResourceWithImportState = &resource{}` or similar.
- Prefer using existing util functions over longer form, duplicated code:
  - `utils.IsKnown(val)` instead of `!val.IsNull() && !val.IsUnknown()`
  - `utils.ListTypeAs` instead of `val.ElementsAs` or similar for other collection types

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

## API Client Usage

- Use generated API clients from `generated/kbapi/` for new Kibana API interactions
- Avoid deprecated clients (`libs/go-kibana-rest`, `generated/alerting`, `generated/connectors`, `generated/slo`)