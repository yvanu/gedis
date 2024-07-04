package server

import (
	"gedis/gedis"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewServer(":6379", gedis.NewGedisHandler())
	s.Start()
}
