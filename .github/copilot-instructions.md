This is a Go based repository hosting a Terrform provider for the elastic stack (elasticsearch and kibana) APIs.  This repo currently supports both [plugin framework](https://developer.hashicorp.com/terraform/plugin/framework/getting-started/code-walkthrough) and [sdkv2](https://developer.hashicorp.com/terraform/plugin/sdkv2) resources. Unless you're told otherwise, all new resources _must_ use the plugin framework. 

For further information, please see [README.md](../README.md) and the [CONTRIBUTING.md](../CONTRIBUTING.md) docs.

## Code Standards

### Required Before Each Commit
- Run `make fmt` before committing any changes to ensure proper code formatting, this will run gofmt on all Go files to maintain consistent style.
- Run `make lint` to ensure any linting errors have not surfaced with your changes

### Required Before Pull Request
- Run `make gen` to update the documentation and code based on your changes.

### Development Flow
- Develop feature or fix bug
- Write tests to validate behavior
- Run `make test` to run test suite

## Repository Structure

• **docs/** - Documentation files
  • **data-sources/** - Documentation for Terraform data sources (51 files)
  • **guides/** - User guides and tutorials
  • **resources/** - Documentation for Terraform resources (35 files)
• **examples/** - Example Terraform configurations
  • **cloud/** - Examples using the cloud to launch testing stacks
  • **data-sources/** - Data source usage examples (45+ examples)
  • **resources/** - Resource usage examples (30+ examples)
  • **provider/** - Provider configuration examples
• **generated/** - Auto-generated clients from the `generate-clients` make target
  • **alerting/** - Kibana alerting API client
  • **connectors/** - Kibana connectors API client
  • **kbapi/** - Kibana API client
  • **slo/** - SLO (Service Level Objective) API client
• **internal/** - Internal Go packages
  • **acctest/** - Acceptance test utilities
  • **clients/** - API client implementations
  • **elasticsearch/** - Elasticsearch-specific logic
  • **fleet/** - Fleet management functionality
  • **kibana/** - Kibana-specific logic
  • **models/** - Data models and structures
  • **schema/** - Connection schema definitions for plugin framework
  • **utils/** - Utility functions
  • **versionutils/** - Version handling utilities
• **libs/** - External libraries
  • **go-kibana-rest/** - Kibana REST API client library
• **provider/** - Core Terraform provider implementation
• **scripts/** - Utility scripts for development and CI
• **templates/** - Template files for documentation generation
  • **data-sources/** - Data source documentation templates
  • **resources/** - Resource documentation templates
  • **guides/** - Guide documentation templates
• **xpprovider/** - Additional provider functionality needed for Crossplane

## Key Guidelines
* Follow Go best practices and idiomatic patterns
* Maintain existing code structure and organization
* Write unit tests for new functionality. Use table-driven unit tests when possible.
* Avoid adding any extra functionality into the `utils` package, instead preferencing adding to a more specific package or creating one to match the purpose
* Think through your planning first using the codebase as your guide before creating new resources and data sources

