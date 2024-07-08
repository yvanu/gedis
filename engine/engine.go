package engine

import (
	"fmt"
	"gedis/aof"
	"gedis/gedis/proto"
	"gedis/iface"
	"gedis/tool/logger"
	"gedis/tool/timewheel"
	"runtime/debug"
	"strings"
	"sync/atomic"
)

const (
	maxDbNum = 13

	aofFileName = "./aof.db"
)

type Engine struct {
	dbSet []*atomic.Value

	delay *timewheel.Delay

	aof *aof.AOF
}

func NewEngine() *Engine {
	e := &Engine{}
	e.delay = timewheel.NewDelay()
	e.dbSet = make([]*atomic.Value, maxDbNum)
	for i := 0; i < maxDbNum; i++ {
		db := newDB(e.delay)
		db.SetIndex(i)
		dbset := new(atomic.Value)
		dbset.Store(db)
		e.dbSet[i] = dbset
	}
	// 默认开启aof
	aof_, err := aof.NewAOF(aofFileName, e, "always")
	if err != nil {
		panic(err)
	}
	e.aof = aof_
	e.aofBindAllDB()
	e.aof.LoadAof(0)

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

func (e *Engine) aofBindAllDB() {
	for _, dbset := range e.dbSet {
		db := dbset.Load().(*DB)
		db.writeAof = func(command [][]byte) {
			e.aof.SaveGedisCommand(db.index, command)
		}
	}
}
