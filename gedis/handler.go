package gedis

import (
	"context"
	"gedis/gedis/parser"
	"gedis/gedis/proto"
	"gedis/iface"
	"gedis/tool/logger"
	"io"
	"net"
	"strings"
	"sync"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

type GedisHandler struct {
	activeConn sync.Map
	engine     iface.Engine
}

func NewGedisHandler() *GedisHandler {
	return &GedisHandler{activeConn: sync.Map{}}
}

func (g *GedisHandler) Handle(ctx context.Context, conn net.Conn) {
	g.activeConn.Store(conn, struct{}{})
	outChan := parser.ParseStream(conn)
	for payload := range outChan {
		if payload.Err != nil {
			if payload.Err == io.EOF || payload.Err == io.ErrUnexpectedEOF || strings.Contains(payload.Err.Error(), "use of closed network connection") {
				g.activeConn.Delete(conn)
				conn.Close()
				logger.Warn("客户端关闭链接: " + conn.RemoteAddr().String())
				break
			}

			errReply := proto.NewGenericErrReply(payload.Err.Error())
			_, err := conn.Write(errReply.Bytes())
			if err != nil {
				g.activeConn.Delete(conn)
				conn.Close()
				logger.Error("Error writing error reply: ", err)
				break
			}
		} else {
			if payload.Reply == nil {
				logger.Error("空payload")
				continue
			}
			reply, ok := payload.Reply.(*proto.MultiBulkReply)
			if !ok {
				logger.Error("is not MultiBulkReply")
				continue
			}
			logger.Debugf("%q", string(reply.Bytes()))

			// redis-cli连接的时候会发送一个command请求，需要返回支持的命令
			conn.Write(proto.NewMultiBulkReply([][]byte{[]byte("get")}).Bytes())
			result := g.engine.Exec(conn, reply.Command)
			if result != nil {
				conn.Write(result.Bytes())
			} else {
				conn.Write(proto.NewGenericErrReply("command exec fail").Bytes())
			}
		}
	}
}

func (g *GedisHandler) Close() error {
	return nil
}
