module github.com/gojek/turing/engines/experiment

go 1.14

require (
	github.com/gojek/turing/engines/experiment/v2 v2.0.0
	github.com/stretchr/testify v1.6.1
)

replace github.com/gojek/turing/engines/experiment/v2 => ../experiments/
