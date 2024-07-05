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

func cmdSet(db *DB, args [][]byte) proto.Reply {
	//if len(args) != 2 {
	//	return proto.NewArgNumErrReply("set")
	//}
	key := string(args[0])
	value := args[1]

	dataEntity := entity.DataEntity{Object: value}
	result := db.PutEntity(key, &dataEntity)
	if result > 0 {
		return proto.NewOkReply()
	}
	return proto.NewNullBulkReply()
}

func init() {
	registerCommand("Get", cmdGet, nil, 2)
	registerCommand("Set", cmdSet, nil, -3)
}
