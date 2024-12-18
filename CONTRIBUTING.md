## Acceptance tests

```bash
make docker-testacc
```

Run a single test with terraform debug enabled:
```bash
make docker-testacc TF_LOG=DEBUG TESTARGS='-run ^TestAccResourceDataStreamLifecycle$$'
```