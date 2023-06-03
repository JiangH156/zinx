package znet

import (
	"errors"
	"fmt"
	"github.com/jiangh156/zinx/ziface"
	"io"
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

	// 消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle

	// 告知当前程序连接已经退出/停止的 channel
	ExitBuffChan chan bool
}

// 初始化连接模块的方法
func NewConnection(conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	return &Connection{
		Conn:         conn,
		ConnId:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
	}
}

//============== 实现 ziface.IConnection 里的全部接口方法 ========

// 直接将Message数据发送给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data封包发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg")
	}
	//写回客户端
	if _, err := c.GetTCPConnection().Write(msg); err != nil {
		fmt.Println("Write msg id ", msgId, " error ")
		c.ExitBuffChan <- true
		return errors.New("conn write error")
	}
	return nil
}

// 连接的读数据业务
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Printf("Reader is exit,remote addr is %s, ConnID = %d\n", c.RemoteAddr(), c.ConnId)
	defer c.Stop()

	for {
		//创建拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg Head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.ExitBuffChan <- true
			continue
		}
		//拆包，得到MsgId 和 dataLen --> Msg
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitBuffChan <- true
			continue
		}
		//根据dataLen读取data，放在msg.data中
		var data []byte //使用预声明方式，可以减少内容
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		//从绑定好的消息和对应的处理方法中执行对应的Handle方法
		go c.MsgHandler.DoMsgHandler(&req)
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
