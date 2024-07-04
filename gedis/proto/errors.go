package proto

import "gedis/util"

type SimpleErrReply struct {
	Status string
}

func (r *SimpleErrReply) Bytes() []byte {
	return []byte("-" + r.Status)
}

func NewSimpleErrReply(status string) Reply {
	return &SimpleErrReply{status}
}

type GenericErrReply struct {
	Status string
}

func (r *GenericErrReply) Bytes() []byte {
	return []byte("-ERR " + r.Status + util.CRLF)
}

func NewGenericErrReply(status string) Reply {
	return &GenericErrReply{status}
}

type UnknownErrReply struct {
}

func (r *UnknownErrReply) Bytes() []byte {
	return []byte("-ERR unknown" + util.CRLF)
}

func NewUnknownErrReply() Reply {
	return &UnknownErrReply{}
}

type ArgNumErrReply struct {
	Cmd string
}

func (r *ArgNumErrReply) Bytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command" + util.CRLF)
}

func NewArgNumErrReply(cmd string) Reply {
	return &ArgNumErrReply{cmd}
}

type WrongTypeErrReply struct {
}

func (r *WrongTypeErrReply) Bytes() []byte {
	return []byte("-WRONGTYPE Operation against a key holding the wrong kind of value" + util.CRLF)
}

func NewWrongTypeErrReply() Reply {
	return &WrongTypeErrReply{}
}

type SyntaxErrReply struct {
}

func (r *SyntaxErrReply) Bytes() []byte {
	return []byte("-ERR syntax error" + util.CRLF)
}

func NewSyntaxErrReply() Reply {
	return &SyntaxErrReply{}
}

func IsErrReply(reply Reply) bool {
	return reply.Bytes()[0] == '-'
}
