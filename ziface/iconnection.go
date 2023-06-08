package ziface

import "net"

// 定义连接模块的抽象
type IConnection interface {
	//  启动连接，为当前连接准备准备开始工作
	Start()

	//  停止工作，结束当前连接的工作
	Stop()

	//  获取当前连接的绑定socket conn
	GetTCPConnection() *net.TCPConn

	//  获取当前连接的连接ID
	GetConnID() uint32

	//  获取远程客户端的TCP状态TP Port
	RemoteAddr() net.Addr

	//直接将Message数据发送给远程的TCP客户端(无缓冲)
	SendMsg(msgId uint32, data []byte) error

	//直接将Message数据发送给远程的TCP客户端(有缓冲)
	SendBuffMsg(msgId uint32, data []byte) error

	//设置连接属性
	SetProperty(key string, value any)
	//获取连接属性
	GetProperty(key string) (any, error)
	//移除连接属性
	RemoveProperty(key string)
}

type HandleFunc func(*net.TCPConn, []byte, int) error
