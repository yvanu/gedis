package engine

import (
	"gedis/aof"
	"gedis/gedis/proto"
	"strconv"
	"time"
)

func execExpire(db *DB, args [][]byte) proto.Reply {
	if len(args) < 2 || len(args) > 3 {
		return proto.NewArgNumErrReply("expire")
	}
	key := string(args[0])
	seconds, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil || seconds < 0 {
		return proto.NewGenericErrReply("invalid expire time")
	}
	_, exist := db.GetEntity(key)
	if !exist {
		return proto.NewIntegerReply(0)
	}

	// todo 暂时不处理nx等情况
	db.ExpireAt(key, time.Now().Add(time.Duration(seconds)*time.Second))
	db.writeAof(aof.ExpireAtCmd(key, seconds))
	return proto.NewIntegerReply(1)
}

func init() {
	registerCommand("Expire", execExpire, nil, -3)
}
