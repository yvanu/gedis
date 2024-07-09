package proto

import (
	"bytes"
	"gedis/util"
	"strconv"
)

type EmptyMultiBulkReply struct {
}

func (r *EmptyMultiBulkReply) Bytes() []byte {
	return []byte("*0" + util.CRLF)
}

func NewEmptyMultiBulkReply() Reply {
	return &EmptyMultiBulkReply{}
}

type MultiBulkReply struct {
	Command [][]byte
}

func (r *MultiBulkReply) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(len(r.Command)) + util.CRLF)
	for _, cmd := range r.Command {
		if cmd == nil {
			buf.WriteString("$-1" + util.CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(cmd)) + util.CRLF + string(cmd) + util.CRLF)
		}
	}
	return buf.Bytes()
}

func NewMultiBulkReply(command [][]byte) Reply {
	return &MultiBulkReply{command}
}

type BulkReply struct {
	Arg []byte
}

func (r *BulkReply) Bytes() []byte {
	if r.Arg == nil {
		return NewNullBulkReply().Bytes()
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + util.CRLF + string(r.Arg) + util.CRLF)
}

func NewBulkReply(arg []byte) Reply {
	return &BulkReply{arg}
}

type NullBulkReply struct {
}

func (r *NullBulkReply) Bytes() []byte {
	return []byte("$-1" + util.CRLF)
}

func NewNullBulkReply() Reply {
	return &NullBulkReply{}
}

type IntegerReply struct {
	Integer int64
}

func (r *IntegerReply) Bytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Integer, 10) + util.CRLF)
}

func NewIntegerReply(integer int64) Reply {
	return &IntegerReply{integer}
}

type MixReply struct {
	replies []Reply
}

func (r *MixReply) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(len(r.replies)) + util.CRLF)
	for _, reply := range r.replies {
		buf.Write(reply.Bytes())
	}
	return buf.Bytes()
}

func NewMixReply(replies ...Reply) Reply {
	return &MixReply{replies}
}
