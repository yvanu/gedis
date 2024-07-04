package gedis

import (
	"context"
	"fmt"
	"net"
)

type TestHandler struct {
}

func (t *TestHandler) Handle(ctx context.Context, conn net.Conn) {
	fmt.Println("test handler")
}

func (t *TestHandler) Close() error {
	fmt.Println("test handler close")
	return nil
}

var _ Handler = &TestHandler{}
