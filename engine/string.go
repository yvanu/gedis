package engine

import (
	"gedis/aof"
	"gedis/engine/entity"
	"gedis/gedis/proto"
	"strconv"
	"strings"
	"time"
)

const (
	noLimitedTTL int64 = 0
)

func (d *DB) getStringObject(key string) ([]byte, proto.Reply, error) {
	entity, exist := d.GetEntity(key)
	if !exist {
		return nil, proto.NewNullBulkReply(), nil
	}
	bytes, ok := entity.Object.([]byte)
	if !ok {
		return nil, proto.NewWrongTypeErrReply(), nil
	}
	return bytes, proto.NewBulkReply(bytes), nil
}

func cmdGet(db *DB, args [][]byte) proto.Reply {
	key := string(args[0])
	bytes, reply, err := db.getStringObject(key)
	if err != nil {
		return reply
	}
	return proto.NewBulkReply(bytes)
}

func cmdMGet(db *DB, args [][]byte) proto.Reply {
	var res = make([][]byte, 0)
	for _, kBytes := range args {
		key := string(kBytes)
		bytes, _, err := db.getStringObject(key)
		if err != nil {
			res = append(res, []byte(""))
			continue
		}
		res = append(res, bytes)
	}
	return proto.NewMultiBulkReply(res)
}

func cmdSet(db *DB, args [][]byte) proto.Reply {
	key := string(args[0])
	value := args[1]

	var ttl = noLimitedTTL

	if len(args) > 2 {
		for i := 2; i < len(args); i++ {
			arg := strings.ToUpper(string(args[i]))
			// 支持过期设置
			if arg == "EX" {
				// 多个Ex
				if ttl != noLimitedTTL {
					return proto.NewSyntaxErrReply()
				}
				if i+1 > len(args) {
					return proto.NewSyntaxErrReply()
				}
				ttlArg, err := strconv.ParseInt(string(args[i+1]), 10, 64)
				if err != nil {
					return proto.NewSyntaxErrReply()
				}
				if ttlArg < 0 {
					return proto.NewGenericErrReply("过期时间不能是负数")
				}
				// 转换成毫秒
				ttl = ttlArg
				// 跳过下一个参数
				i++
			}
		}
	}

	dataEntity := entity.DataEntity{Object: value}
	result := db.PutEntity(key, &dataEntity)
	if result > 0 {
		if ttl != noLimitedTTL {
			expireTime := time.Now().Add(time.Duration(ttl) * time.Second)
			db.ExpireAt(key, expireTime)
			db.writeAof(aof.SetCmd(aof.Command{args[0], args[1]}...))
		} else {
			db.Persist(key)
			db.writeAof(aof.SetCmd(args...))
		}
		return proto.NewOkReply()
	}
	return proto.NewNullBulkReply()
}

func cmdMSet(db *DB, args [][]byte) proto.Reply {
	for i := 0; i < len(args); i = i + 2 {
		key := string(args[i])
		value := args[i+1]

		dataEntity := entity.DataEntity{Object: value}
		db.PutEntity(key, &dataEntity)
	}
	return proto.NewOkReply()
}

func init() {
	registerCommand("Get", cmdGet, nil, 2)
	registerCommand("MGet", cmdMGet, nil, -2)
	registerCommand("Set", cmdSet, nil, -3)
	registerCommand("MSet", cmdMSet, nil, -3)
}
