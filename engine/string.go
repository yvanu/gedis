package engine

import (
	"gedis/engine/entity"
	"gedis/gedis/proto"
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

	dataEntity := entity.DataEntity{Object: value}
	result := db.PutEntity(key, &dataEntity)
	if result > 0 {
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
