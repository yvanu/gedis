package engine

import (
	"fmt"
	"gedis/gedis/proto"
	"gedis/iface"
	"gedis/tool/logger"
	"runtime/debug"
	"strings"
	"sync/atomic"
)

const (
	maxDbNum = 13
)

type Engine struct {
	dbSet []*atomic.Value
}

func NewEngine() *Engine {
	e := &Engine{}
	e.dbSet = make([]*atomic.Value, maxDbNum)
	for i := 0; i < maxDbNum; i++ {
		db := newDB()
		db.SetIndex(i)
		dbset := new(atomic.Value)
		dbset.Store(db)
		e.dbSet[i] = dbset
	}
	return e
}

func (e *Engine) Close() {

}

func (e *Engine) Exec(conn iface.Conn, command [][]byte) (result proto.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("执行命令失败: %v\n%s", err, string(debug.Stack())))
			result = proto.NewGenericErrReply(err.(string))
		}
	}()
	commandName := strings.ToLower(string(command[0]))
	if commandName == "command" {
		return proto.NewMultiBulkReply([][]byte{[]byte("ok")})
	}
	if commandName == "ping" {
		return Ping(command[1:])
	}
	if commandName == "auth" {
		return Auth(command[1:])
	}

	switch commandName {
	case "select":
		return execSelect(conn, command[1:])
	}

	dbIndex := conn.GetDbIndex()
	logger.Debugf("db index: %v\n", dbIndex)
	db := e.selectDb(dbIndex)
	return db.Exec(conn, command)
}

func (e *Engine) selectDb(index int) *DB {
	return e.dbSet[index].Load().(*DB)
}
