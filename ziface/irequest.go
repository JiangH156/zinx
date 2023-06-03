package ziface

// IRequest 包装了客户端请求的连接信息、请求的数据
type IRequest interface {
	GetConnection() IConnection // 获取请求连接信息
	GetData() []byte            // 获取请求信息的数据
	GetMsgId() uint32           //获取消息ID
}
