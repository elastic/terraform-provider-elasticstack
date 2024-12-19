# Unit tests

To run the unit tests, you can use the following command:
```bash

docker compose up -d
make test
```

One could enable debug output for tests by setting environment variable `DEBUG=true`, that helps to see http requests and responses. 

Here is an example of environment variables to run a single test, e.g. from IDE or CLI:
```bash
DEBUG=true TEST="-run TestKBAPITestSuite/TestKibanaSpaces ./..." make test
```