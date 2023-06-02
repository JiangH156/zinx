package ziface

import "net"

// IConnection 定义连接模块的抽象
type IConnection interface {
	// Start 启动连接，为当前连接准备准备开始工作
	Start()

	// Stop 停止工作，结束当前连接的工作
	Stop()

	// GetTCPConnection 获取当前连接的绑定socket conn
	GetTCPConnection() *net.TCPConn

	// GetConnId 获取当前连接的连接ID
	GetConnId() uint32

	// RemoteAddr 获取远程客户端的TCP状态TP Port
	RemoteAddr() net.Addr
}

// HandleFunc 定义一个处理连接业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error
