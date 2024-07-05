package gedisconn

import (
	"gedis/iface"
	"net"
)

type Connection struct {
	c net.Conn

	dbIndex int
}

func (c *Connection) GetDbIndex() int {
	return c.dbIndex
}

func (c *Connection) SetDbIndex(dbIndex int) {
	c.dbIndex = dbIndex
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{c: conn}
}

var _ iface.Conn = new(Connection)
