module github.com/caraml-dev/turing/engines/experiment/examples/plugins/hardcoded

go 1.18

require (
	github.com/caraml-dev/turing/engines/experiment v1.0.0
	github.com/gojek/mlp v1.5.3
	github.com/hashicorp/go-hclog v0.16.0
	github.com/stretchr/testify v1.8.0
)

replace (
	github.com/caraml-dev/turing/engines/experiment => ../../../
	github.com/caraml-dev/turing/engines/router => ../../../../router
)
