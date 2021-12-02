module github.com/gojek/turing/engines/plugins/nop/v1

go 1.14

require (
	github.com/gojek/turing/engines/experiment/v2 v2.0.0
	github.com/hashicorp/go-plugin v1.4.3
)

replace github.com/gojek/turing/engines/experiment/v2 => ./../../
