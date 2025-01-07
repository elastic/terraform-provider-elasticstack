# Typical development workflow

Fork the repo, work on an issue

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