package iface

type Conn interface {
	SetDbIndex(dbIndex int)
	GetDbIndex() int
}
