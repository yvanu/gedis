package engine

import (
	"gedis/gedis/proto"
	"gedis/iface"
	"strconv"
)

func execSelect(c iface.Conn, args [][]byte) proto.Reply {
	if len(args) != 1 {
		return proto.NewArgNumErrReply("select")
	}
	dbIndex, err := strconv.ParseInt(string(args[0]), 10, 64)
	if err != nil {
		return proto.NewGenericErrReply("db index error")
	}
	if dbIndex < 0 || dbIndex > maxDbNum {
		return proto.NewGenericErrReply("db index out of range")
	}
	c.SetDbIndex(int(dbIndex))
	return proto.NewOkReply()
}
