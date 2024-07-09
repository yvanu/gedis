package iface

type Conn interface {
	SetDbIndex(dbIndex int)
	GetDbIndex() int

	// 发布订阅相关
	Subscribe(chanName string)
	SubCount() int
	Write([]byte) (int, error)
	IsClosed() bool
	GetChannels() []string
	UnSubscribe(key string)
}
