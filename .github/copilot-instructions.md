You will be writing or reviewing code for the Terraform provider for Elastic Stack (Elasticsearch, Kibana, Fleet, APM, and Logstash). This is a Go-based repository hosting the provider source. 

When writing code, you must adhere to the coding standards and conventions outlined in the [CODING_STANDARDS.md](../CODING_STANDARDS.md) document in this repository.

When reviewing code, ensure that all changes comply with the coding standards and conventions specified in the [CODING_STANDARDS.md](../CODING_STANDARDS.md) document. Pay special attention to project structure, schema definitions, JSON handling, resource implementation, and testing practices.

Take your time and think through every step - remember to check solutions rigorously and watch out for boundary cases, especially with the changes being made. 

When writing code, your solution must be perfect. If not, continue working on it. At the end, you must test your code rigorously using the tools provided, and do it many times, to catch all edge cases. If it is not robust, iterate more and make it perfect. Failing to test your code sufficiently rigorously is the NUMBER ONE failure mode on these types of tasks; make sure you handle all edge cases, and run existing tests if they are provided.

Please see [README.md](../README.md) and the [CONTRIBUTING.md](../CONTRIBUTING.md) docs before getting started.

# Development Workflow

## High-Level Problem Solving Strategy

1. Understand the problem deeply. Carefully read the issue and think critically about what is required.
2. Investigate the codebase. Explore relevant files, search for key functions, and gather context.
3. Develop a clear, step-by-step plan. Break down the fix into manageable, incremental steps.
4. Implement the fix incrementally. Make small, testable code changes.
5. Debug as needed. Use debugging techniques to isolate and resolve issues.
6. Test frequently. Run tests after each change to verify correctness.
7. Iterate until the root cause is fixed and all tests pass.
8. Reflect and validate comprehensively. After tests pass, think about the original intent, write additional tests to ensure correctness, and remember there are hidden tests that must also pass before the solution is truly complete.

Refer to the detailed sections below for more information on each step.

## 1. Deeply Understand the Problem
Carefully read the issue and think hard about a plan to solve it before coding. Your thinking should be thorough and so it's fine if it's very long. You can think step by step before and after each action you decide to take. 

## 2. Codebase Investigation
- Explore relevant files and directories.
- Search for key functions, classes, or variables related to the issue.
- Read and understand relevant code snippets.
- Identify the root cause of the problem.
- Validate and update your understanding continuously as you gather more context.

## 3. Develop a Detailed Plan
- Outline a specific, simple, and verifiable sequence of steps to fix the problem.
- Break down the fix into small, incremental changes.

## 4. Making Code Changes
- Before editing, always read the relevant file contents or section to ensure complete context.
- If a patch is not applied correctly, attempt to reapply it.
- Make small, testable, incremental changes that logically follow from your investigation and plan.

## 5. Debugging
- Make code changes only if you have high confidence they can solve the problem
- When debugging, try to determine the root cause rather than addressing symptoms
- Debug for as long as needed to identify the root cause and identify a fix
- Use print statements, logs, or temporary code to inspect program state, including descriptive statements or error messages to understand what's happening
- To test hypotheses, you can also add test statements or functions
- Revisit your assumptions if unexpected behavior occurs.
- You MUST iterate and keep going until the problem is solved.

## 6. Testing
- Run tests frequently using `make test` and `make testacc`
- After each change, verify correctness by running relevant tests.
- If tests fail, analyze failures and revise your patch.
- Write additional tests if needed to capture important behaviors or edge cases.
- NEVER accept acceptance tests that have been skipped due to environment issues; always ensure the environment is correctly set up and all tests run successfully.

### 6.1 Acceptance Testing Requirements
When running acceptance tests, ensure the following:

- **Environment Variables** - The following environment variables are required for acceptance tests:
  - `ELASTICSEARCH_ENDPOINTS` (default: http://localhost:9200)
  - `ELASTICSEARCH_USERNAME` (default: elastic) 
  - `ELASTICSEARCH_PASSWORD` (default: password)
  - `KIBANA_ENDPOINT` (default: http://localhost:5601)
  - `TF_ACC` (must be set to "1" to enable acceptance tests)
- **Run targeted tests using `go test`** - Ensure the required environment variables are explicitly defined when running targeted tests. Example:
  ```bash
  ELASTICSEARCH_ENDPOINTS=http://localhost:9200 ELASTICSEARCH_USERNAME=elastic ELASTICSEARCH_PASSWORD=password KIBANA_ENDPOINT=http://localhost:5601 TF_ACC=1 go test -v -run TestAccResourceName ./path/to/testfile.go
  ```

## 7. Final Verification
- Confirm the root cause is fixed.
- Review your solution for logic correctness and robustness.
- Iterate until you are extremely confident the fix is complete and all tests pass.
- Run the acceptance tests for any changed resources. Ensure acceptance tests pass without any environment-related skips. Use `make testacc` to verify this, explicitly defining the required environment variables.
- Run `make lint` to ensure any linting errors have not surfaced with your changes. This task may automatically correct any linting errors, and regenerate documentation. Include any changes in your commit. 

## 8. Final Reflection and Additional Testing
- Reflect carefully on the original intent of the user and the problem statement.
- Think about potential edge cases or scenarios that may not be covered by existing tests.
- Write additional tests that would need to pass to fully validate the correctness of your solution.
- Run these new tests and ensure they all pass.
- Be aware that there are additional hidden tests that must also pass for the solution to be successful.
- Do not assume the task is complete just because the visible tests pass; continue refining until you are confident the fix is robust and comprehensive.

## 9. Before Submitting Pull Requests
- Run `make docs-generate` to update the documentation, and ensure the results of this command make it into your pull request.

## Repository Structure

• **docs/** - Documentation files
  • **data-sources/** - Documentation for Terraform data sources
  • **guides/** - User guides and tutorials
  • **resources/** - Documentation for Terraform resources
• **examples/** - Example Terraform configurations
  • **cloud/** - Examples using the cloud to launch testing stacks
  • **data-sources/** - Data source usage examples
  • **resources/** - Resource usage examples
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
* When creating a new Plugin Framework based resource, follow the code organisation of `internal/elasticsearch/security/system_user` 
* Avoid adding any extra functionality into the `utils` package, instead preferencing adding to a more specific package or creating one to match the purpose
* Think through your planning first using the codebase as your guide before creating new resources and data sources

