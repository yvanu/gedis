package engine

import (
	"gedis/datastruct/dict"
	"gedis/engine/entity"
	"gedis/gedis/proto"
	"gedis/iface"
	"strings"
)

const (
	dataDictSize = 1 << 16
)

type DB struct {
	index int

	dataDict *dict.ConcurrentDict
}

func newDB() *DB {
	db := &DB{
		dataDict: dict.NewConcurrentDict(dataDictSize),
	}
	return db
}

func (d *DB) SetIndex(index int) {
	d.index = index
}

func (d *DB) Exec(c iface.Conn, command [][]byte) proto.Reply {
	//cmdName := strings.ToLower(string(command[0]))
	return d.execNormalCommand(command)
}

func (d *DB) execNormalCommand(command [][]byte) proto.Reply {
	cmdName := strings.ToLower(string(command[0]))
	cmd, ok := commandCenter[cmdName]
	if !ok {
		return proto.NewGenericErrReply("未知的命令：" + cmdName)
	}
	if !validateArgsNum(cmd.argsNum, command) {
		return proto.NewArgNumErrReply(cmdName)
	}

	return cmd.execFunc(d, command[1:])
}

func (d *DB) GetEntity(key string) (*entity.DataEntity, bool) {
	val, exist := d.dataDict.Get(key)
	if !exist {
		return nil, false
	}
	dataEntity, ok := val.(*entity.DataEntity)
	return dataEntity, ok
}

func (d *DB) PutEntity(key string, e *entity.DataEntity) int {
	d.dataDict.Put(key, e)
	return 1
}

func validateArgsNum(num int, args [][]byte) bool {
	argNum := len(args)
	if num >= 0 {
		return argNum == num
	}
	return argNum >= -num
}
