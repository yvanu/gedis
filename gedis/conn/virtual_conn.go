package gedisconn

// 用于回放时

type VirtualConn struct {
	Connection
}

func (v *VirtualConn) GetDbIndex() int {
	return v.dbIndex
}

func (v *VirtualConn) SetDbIndex(dbIndex int) {
	v.dbIndex = dbIndex
}

func NewVirtualConnection() *VirtualConn {
	return &VirtualConn{}
}
