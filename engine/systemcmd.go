package engine

import (
	"gedis/gedis/proto"
)

func Ping(args [][]byte) proto.Reply {
	if len(args) == 0 {
		return proto.NewPongReply()
	} else if len(args) == 1 {
		return proto.NewBulkReply(args[1])
	}
	return proto.NewArgNumErrReply("ping")
}

func Auth(args [][]byte) proto.Reply {
	if len(args) != 1 {
		return proto.NewArgNumErrReply("auth")
	}
	password := string(args[0])
	_ = password
	// todo 先不验证密码，后续添加
	return proto.NewOkReply()
}
