package iface

import "gedis/gedis/proto"

type Engine interface {
	Exec(conn Conn, command [][]byte) (result proto.Reply)
	Close()
}
