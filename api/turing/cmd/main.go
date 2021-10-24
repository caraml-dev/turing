package main

import (
	"github.com/gojek/turing/api/turing/server"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	server.Run()
}
