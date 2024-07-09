package gedisconn

import (
	"gedis/iface"
	"net"
	"sync"
	"sync/atomic"
)

type Connection struct {
	c net.Conn

	dbIndex int

	mu   sync.Mutex
	subs map[string]struct{}

	closed atomic.Bool
}

func (c *Connection) UnSubscribe(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subs, key)
}

func (c *Connection) GetChannels() []string {

	var result = make([]string, 0)
	for channel, _ := range c.subs {
		result = append(result, channel)
	}
	return result
}

func (c *Connection) IsClosed() bool {
	return c.closed.Load()
}

func (c *Connection) Subscribe(chanName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.subs == nil {
		c.subs = make(map[string]struct{})
	}

	c.subs[chanName] = struct{}{}
}

func (c *Connection) SubCount() int {
	return len(c.subs)

}

func (c *Connection) Write(bytes []byte) (int, error) {
	if len(bytes) == 0 {
		return 0, nil
	}
	return c.c.Write(bytes)
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
