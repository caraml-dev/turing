# Turing SDK

Python SDK for interacting with Turing, machine learning models testing 
and experiments configuration component Gojek MLP.

## Install
Install Turing SDK from PyPI:
```shell
pip install turing-sdk
```

## Getting Started

Check out [samples](./samples) for examples on how to use Turing SDK.

* Quickstart – [samples/quickstart](./samples/quickstart)

## Development

#### Prerequisites

* Python >=3.7,<3.10
* openapi-generator >= 5.1.0 (`brew install openapi-generator`)

### Make commands

* Setup development environment
```shell
make setup
```

* (Re-)generate openapi client
```shell
make gen-client
```

* Run unit tests
```shell
make test
```
