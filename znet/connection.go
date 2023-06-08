package znet

import (
	"errors"
	"fmt"
	"github.com/jiangh156/zinx/utils"
	"github.com/jiangh156/zinx/ziface"
	"io"
	"net"
)

// Connection 连接模块
type Connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer
	// 当前连接的socket TCP套接字
	Conn *net.TCPConn

	// 连接ID
	ConnID uint32

	// 当前的连接状态
	isClosed bool

	// 消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle

	// 告知当前程序连接已经退出/停止的 channel
	ExitBuffChan chan bool

	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte

	//有关冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte
}

// 初始化连接模块的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte), // msgChan初始化
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
	}
	//将新创建的Conn添加到连接管理中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

// 直接将Message数据发送给远程的TCP客户端(有缓冲)
func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	//将data封包发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msgId = ", msg)
		return errors.New("Pack error msg")
	}
	//写回客户端
	c.msgBuffChan <- msg
	return nil
}

// 启动连接，为当前连接准备准备开始工作
func (c *Connection) Start() {
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()

	//==================
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)
	//==================
}

// 直接将Message数据发送给远程的TCP客户端(无缓冲)
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
	c.msgChan <- msg //将之前直接写回conn.Writer的方法改为发送给Channel的 供Writer读取发送
	return nil
}

func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String() + "[conn Writer exit!]")
	defer c.Stop()

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			//针对有缓冲channel需要些的数据处理
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				//ok为false，说明msgBuffChan已经关闭
				fmt.Println("msgBuffChan is Closed")
				break
			}
		case <-c.ExitBuffChan:
			//conn已经关闭
			return
		}
	}

}

// 连接的读数据业务
func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Reader exit!]")
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
		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经开启工作机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

// 停止工作，结束当前连接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()...,ConnID = ", c.ConnID)

	// 如果当前连接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//==================
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)
	//==================

	// 关闭socket连接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该连接已经关闭
	c.ExitBuffChan <- true

	//将连接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c)

	// 关闭该连接全部管道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
}

// 获取当前连接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接的连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端的TCP状态TP Port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
