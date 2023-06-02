package ziface

// IRequest 包装了客户端请求的连接信息、请求的数据
type IRequest interface {
	// GetConnection 得到当前连接
	GetConnection() IConnection
	// GetData 得到请求的消息数据
	GetData() []byte
}
