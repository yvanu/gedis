package engine

import (
	"gedis/gedis/proto"
	"strings"
)

type ExecFunc func(db *DB, args [][]byte) proto.Reply

type KeyFunc func(args [][]byte) ([]string, []string)

var commandCenter = make(map[string]*command)

type command struct {
	name     string
	execFunc ExecFunc
	keyFunc  KeyFunc
	argsNum  int // number of arguments 负数表示要大于后面的值
}

func registerCommand(name string, execFunc ExecFunc, keyFunc KeyFunc, argsNum int) {
	name = strings.ToLower(name)
	cmd := command{name, execFunc, keyFunc, argsNum}
	commandCenter[name] = &cmd
}
