# Typical development workflow

Fork the repo, work on an issue

## Updating the generated Kibana client.

If your work involves the Kibana API, the endpoints may or may not be included in the generated client.
Check [generated/kbapi](./generated/kbapi/) for more details.

## Acceptance tests

```bash
make docker-testacc
```

Run a single test with terraform debug enabled:
```bash
env TF_LOG=DEBUG make docker-testacc TESTARGS='-run ^TestAccResourceDataStreamLifecycle$$'
```

A way to forward debug logs to a file:
```bash
env TF_ACC_LOG_PATH=/tmp/tf.log TF_ACC_LOG=DEBUG TF_LOG=DEBUG make docker-testacc
```


## Update documentation

Update documentation templates in `./templates` directory and re-generate docs via:
```bash
make docs-generate
```

## Update `./CHANGELOG.md`

List of previous commits is a good example of what should be included in the changelog.


## Pull request

Format the code before pushing:
```bash
make fmt
```

Check if the linting:
```bash
make lint
```

Create a PR and check acceptance test matrix is green.

## Run provider with local terraform

TBD

## Releasing

Releasing is implemented in CI pipeline.

To release a new provider version:

* Create PR which
- updates Makefile with the new provider VERSION (e.g. `VERSION ?= 0.11.13`);
- updates CHANGELOG.md with the list of changes being released.
[Example](https://github.com/elastic/terraform-provider-elasticstack/commit/be866ebc918184e843dc1dd2f6e2e1b963da386d).

* Once the PR is merged, the release CI pipeline can be started by pushing a new release tag to the `main` branch.
