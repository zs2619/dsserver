package kissnet

type RpcConnectorMgr struct {
	RpcConnMap map[string]*RpcConnector
}

func (this *RpcConnectorMgr) Init() error {
	rc := &RpcConnector{}
	rc.start()
	this.addRpcConnector(rc)
	return nil
}

func (this *RpcConnectorMgr) addRpcConnector(c *RpcConnector) error {
	this.RpcConnMap[c.serviceName] = c
	return nil
}
