package engine

import (
	"gedis/aof"
	"gedis/datastruct/sortedset"
	"gedis/gedis/proto"
	"strconv"
)

// ZADD key score member [score member...]
func cmdZAdd(db *DB, args [][]byte) proto.Reply {
	if len(args)%2 != 1 {
		return proto.NewArgNumErrReply("zadd")
	}
	key := string(args[0])
	size := (len(args) - 1) / 2
	pairs := make([]*sortedset.Pair, size)

	for i := 0; i < size; i++ {
		score, err := strconv.ParseFloat(string(args[i*2+1]), 64)
		if err != nil {
			return proto.NewGenericErrReply("value is not valid fload")
		}
		pair := &sortedset.Pair{
			Member: string(args[i*2+2]),
			Score:  score,
		}
		pairs[i] = pair
	}

	sortedSet, err, reply := db.getOrInitSortedSetObject(key)
	if err != nil {
		return reply
	}
	i := int64(0)
	for _, pair := range pairs {
		if sortedSet.Add(pair.Member, pair.Score) {
			i++
		}
	}
	db.writeAof(aof.ZAddCmd(args...))
	return proto.NewIntegerReply(i)
}

// zscore key member
func cmdZScore(db *DB, args [][]byte) proto.Reply {
	if len(args) != 2 {
		return proto.NewArgNumErrReply("zscore")
	}
	key := string(args[0])
	member := string(args[1])
	sortedSet, reply := db.getSortedSetObject(key)
	if reply != nil {
		return reply
	}
	if sortedSet == nil {
		return proto.NewNullBulkReply()
	}
	pair, exist := sortedSet.Get(member)
	if !exist {
		return proto.NewNullBulkReply()
	}
	return proto.NewBulkReply([]byte(strconv.FormatFloat(pair.Score, 'f', -1, 64)))
}

func init() {
	registerCommand("ZAdd", cmdZAdd, nil, -4)
	registerCommand("ZScore", cmdZScore, nil, -3)
}
