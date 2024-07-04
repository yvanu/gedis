package server

import (
	"gedis/gedis"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewServer(":9000", gedis.NewGedisHandler())
	s.Start()
	//_ = s
}
