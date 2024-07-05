package main

import (
	"gedis/gedis"
	"gedis/server"
)

func main() {
	s := server.NewServer(":9000", gedis.NewGedisHandler())
	s.Start()
}
