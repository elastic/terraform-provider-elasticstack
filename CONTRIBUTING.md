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