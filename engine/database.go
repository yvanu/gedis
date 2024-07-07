package engine

import (
	"gedis/datastruct/dict"
	"gedis/engine/entity"
	"gedis/gedis/proto"
	"gedis/iface"
	"gedis/tool/logger"
	"gedis/tool/timewheel"
	"strings"
	"time"
)

const (
	dataDictSize = 1 << 16
)

type DB struct {
	index int

	dataDict *dict.ConcurrentDict
	ttlDict  *dict.ConcurrentDict

	delay *timewheel.Delay
}

func newDB(delay *timewheel.Delay) *DB {
	db := &DB{
		dataDict: dict.NewConcurrentDict(dataDictSize),
		ttlDict:  dict.NewConcurrentDict(dataDictSize),
		delay:    delay,
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
	// 检测key是否过期
	if d.IsExpire(key) {
		return nil, false
	}
	dataEntity, ok := val.(*entity.DataEntity)
	return dataEntity, ok
}

func (d *DB) PutEntity(key string, e *entity.DataEntity) int {
	d.dataDict.Put(key, e)
	return 1
}

func (d *DB) IsExpire(key string) bool {
	val, exist := d.ttlDict.Get(key)
	if !exist {
		return false
	}
	expireTime := val.(time.Time)
	isExpire := time.Now().After(expireTime)
	if isExpire {
		d.Remove(key)
	}
	return isExpire
}

func (d *DB) Remove(key string) {
	d.ttlDict.Delete(key)
	d.dataDict.Delete(key)
}

func (d *DB) ExpireAt(key string, expireTime time.Time) {
	d.ttlDict.Put(key, expireTime)
	d.addDelayAt(key, expireTime)
}

func genExpireKey(key string) string {
	return "expire" + key
}

func (d *DB) addDelayAt(key string, expireTime time.Time) {
	if d.delay != nil {
		d.delay.AddAt(expireTime, genExpireKey(key), func() {
			logger.Debug("expire: ", key)

			d.IsExpire(key)
		})
	}
}

func validateArgsNum(num int, args [][]byte) bool {
	argNum := len(args)
	if num >= 0 {
		return argNum == num
	}
	return argNum >= -num
}
