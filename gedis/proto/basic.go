package proto

import "gedis/util"

type SimpleStringReply struct {
	S string
}

func (s *SimpleStringReply) Bytes() []byte {
	return []byte("+" + s.S + util.CRLF)
}

func NewSimpleStringReply(s string) Reply {
	return &SimpleStringReply{s}
}

var okReply = &SimpleStringReply{"ok"}

func NewOkReply() Reply {
	return okReply
}

var pongReply = &SimpleStringReply{"pong"}

func NewPongReply() Reply {
	return pongReply
}

var queueReply = &SimpleStringReply{"queue"}

func NewQueueReply() Reply {
	return queueReply
}

type EmptyReply struct {
}

func (e *EmptyReply) Bytes() []byte {
	return []byte("")
}

func NewEmptyReply() Reply {
	return &EmptyReply{}
}

func IsOkReply(reply Reply) bool {
	return string(reply.Bytes()) == "+OK"+util.CRLF
}
