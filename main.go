package main

import (
	"github.com/blinky-z/Blog/server"
	_ "github.com/lib/pq"
)

func main() {
	server.RunServer()
}
