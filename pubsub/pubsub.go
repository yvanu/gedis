package pubsub

import (
	"gedis/datastruct/dict"
	"gedis/datastruct/list"
	"gedis/gedis/proto"
	"gedis/iface"
	"gedis/tool/locker"
	"gedis/tool/logger"
	"gedis/util"
	"strconv"
)

type Pubsub struct {
	dataDict *dict.ConcurrentDict
	locker   *locker.Locker
}

func NewPubsub() *Pubsub {
	return &Pubsub{
		dataDict: dict.NewConcurrentDict(16),
		locker:   locker.NewLocker(16),
	}
}

func (p *Pubsub) Publish(c iface.Conn, args [][]byte) proto.Reply {
	if len(args) != 2 {
		return proto.NewArgNumErrReply("publish")
	}

	chanName := string(args[0])
	p.locker.Locks(chanName)
	defer p.locker.Unlocks(chanName)

	raw, exist := p.dataDict.Get(chanName)
	if exist {
		var success int64
		// 用于记录失败的客户端
		var failedClient = make(map[any]struct{})

		l := raw.(*list.LinkedList)
		l.ForEach(func(idx int, val any) bool {
			conn := val.(iface.Conn)
			if conn.IsClosed() {
				failedClient[conn] = struct{}{}
				return true
			}
			if val == c {
				return true
			}
			conn.Write(publishMsg(chanName, string(args[1])))
			success++
			return true
		})
		if len(failedClient) != 0 {
			removed := l.DelByVal(func(val any) bool {
				_, ok := failedClient[val]
				return ok
			})
			logger.Debugf("del %d closed client", removed)
		}
		return proto.NewIntegerReply(success)
	}
	return proto.NewIntegerReply(0)
}

func publishMsg(channel, msg string) []byte {
	return []byte("*3" + util.CRLF +
		"$" + strconv.Itoa(len("message")) + util.CRLF + "message" + util.CRLF +
		"$" + strconv.Itoa(len(channel)) + util.CRLF + channel + util.CRLF +
		"$" + strconv.Itoa(len(msg)) + util.CRLF + msg + util.CRLF)
}

func (p *Pubsub) Subscribe(c iface.Conn, args [][]byte) proto.Reply {
	if len(args) < 1 {
		return proto.NewArgNumErrReply("subscribe")
	}
	keys := make([]string, 0, len(args))
	for _, key := range args {
		keys = append(keys, string(key))
	}

	p.locker.Locks(keys...)
	defer p.locker.Unlocks(keys...)

	for _, key := range keys {
		// 记录当前客户端订阅的通道
		c.Subscribe(key)
		var l *list.LinkedList
		raw, exist := p.dataDict.Get(key)
		if !exist {
			l = list.NewLinkedList()
			p.dataDict.Put(key, l)
		} else {
			l = raw.(*list.LinkedList)
		}

		//  看看l中是否包含了该客户端
		if !l.Contain(func(val any) bool {
			return c == val
		}) {
			logger.Debug("subscribe channel [" + key + "] success")
			l.Add(c)
		}

		_, err := c.Write(channelMsg("subscribe", key, c.SubCount()))
		if err != nil {
			logger.Warn("Error writing channel message: ", err)
		}
	}

	return proto.NewEmptyReply()
}

func channelMsg(action, chanName string, count int) []byte {
	return []byte("*3" + util.CRLF +
		"$" + strconv.Itoa(len(action)) + util.CRLF + action + util.CRLF +
		"$" + strconv.Itoa(len(chanName)) + util.CRLF + chanName + util.CRLF +
		":" + strconv.Itoa(count) + util.CRLF)
}

func (p *Pubsub) Unsubscribe(c iface.Conn, args [][]byte) proto.Reply {
	var channelList = make([]string, 0)
	if len(args) < 1 {
		// 取消全部
		channelList = c.GetChannels()
	}
	keys := make([]string, 0, len(args))
	for _, key := range args {
		keys = append(keys, string(key))
	}
	channelList = append(channelList, keys...)
	if len(channelList) == 0 {
		return proto.NewSimpleStringReply("no channel now")
	}

	p.locker.Locks(channelList...)
	defer p.locker.Unlocks(channelList...)

	for _, key := range channelList {
		// 记录当前客户端订阅的通道
		c.UnSubscribe(key)
		var l *list.LinkedList
		raw, exist := p.dataDict.Get(key)
		if exist {
			l = raw.(*list.LinkedList)
		} else {
			continue
		}

		//  看看l中是否包含了该客户端
		logger.Debug("unsubscribe channel [" + key + "] success")
		l.DelByVal(func(val any) bool {
			return c == val
		})
		if l.Len() == 0 {
			p.dataDict.Delete(key)
		}

		_, err := c.Write(channelMsg("unsubscribe", key, c.SubCount()))
		if err != nil {
			logger.Warn("Error writing channel message: ", err)
		}
	}

	return proto.NewEmptyReply()
}
