package znet

import (
	"fmt"
	"github.com/jiangh156/zinx/ziface"
	"net"
)

// Connection 连接模块
type Connection struct {
	// 当前连接的socket TCP套接字
	Conn *net.TCPConn

	// 连接ID
	ConnId uint32

	// 当前的连接状态
	isClosed bool

	// 该连接的处理方法router
	Router ziface.IRouter

	// 告知当前程序连接已经退出/停止的 channel
	ExitBuffChan chan bool
}

// 初始化连接模块的方法
func NewConnection(conn *net.TCPConn, connID uint32, router ziface.IRouter) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnId:       connID,
		isClosed:     false,
		Router:       router,
		ExitBuffChan: make(chan bool, 1),
	}
	return c
}

// 连接的读数据业务
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Printf("Reader is exit,remote addr is %s, ConnID = %d\n", c.RemoteAddr(), c.ConnId)
	defer c.Stop()

	for {
		// 读取客户端的数据到buf中，最大512字节
		buf := make([]byte, 512)
		_, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("recv buf err", err)
			c.ExitBuffChan <- true
			continue
		}
		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			data: buf,
		}
		// 从路由Router中找到注册绑定的Conn所对应的Handle
		go func(request ziface.IRequest) {
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)
	}
}

// 启动连接，为当前连接准备准备开始工作
func (c *Connection) Start() {
	//开启处理该连接读取到客户端数据之后的请求业务
	go c.StartReader()

	for {
		select {
		case <-c.ExitBuffChan:
			//得到退出消息，不再阻塞
			return
		}
	}
}

// 停止工作，结束当前连接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()...,ConnID = ", c.ConnId)

	// 如果当前连接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//TODO Connection Stop() 如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用

	// 关闭socket连接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该连接已经关闭
	c.ExitBuffChan <- true

	// 关闭该连接全部管道
	close(c.ExitBuffChan)
}

// 获取当前连接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接的连接ID
func (c *Connection) GetConnId() uint32 {
	return c.ConnId
}

// 获取远程客户端的TCP状态TP Port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
