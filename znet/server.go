package znet

import (
	"errors"
	"fmt"
	"github.com/jiangh156/zinx/utils"
	"github.com/jiangh156/zinx/ziface"
	"net"
	"time"
)

// IServer的接口实现，定义一个Server的服务类
type Server struct {
	//  服务器的名称
	Name string
	//  服务器绑定的IP版本
	IPVersion string
	//   服务器监听IP地址
	IP string
	//   服务器监听的端口号
	Port int
	//当前server由用户绑定的回调router，也就是Server注册的链接对应的处理业务
	Router ziface.IRouter
}

// NewServer 初始化Server模块的方法
func NewServer() ziface.IServer {
	//1 初始化全局配置文件
	utils.GlobalObject.Reload()
	return &Server{
		Name:      utils.GlobalObject.Name, //全局参数获取
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,    //全局参数获取
		Port:      utils.GlobalObject.TcpPort, //全局参数获取
		Router:    nil,
	}

}

// 定义当前客户端连接所绑定的handleAPI（目前写死，以后用户自定义）
func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	// 回显的业务
	fmt.Println("[Conn Handle] CallbackToClient...")
	if _, err := conn.Write(data[:cnt]); err != nil {
		fmt.Println("write back buf err, ", err)
		return errors.New("CallBackToClient error")
	}
	return nil
}

//============== 实现 ziface.IServer 里的全部接口方法 ========

// 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server Listenner at IP:%s, Port:%d is starting\n", s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d,  MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)

	// 开启一个go去做服务端listen业务
	go func() {
		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error:", err)
			return
		}
		//2 监听服务器的地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Printf("listen %s error: %s\n", s.IP, err)
			return
		}
		// 监听成功
		fmt.Printf("start Zinx server succ, %s is Listenning...\n", s.Name)

		//TODO server.go 应该有一个自动生成ID的方法
		var cid uint32
		cid = 0

		//3 启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept error:", err)
				continue
			}
			//3.2 TODO Server.Start()设置服务器最大连接控制，如果超过最大连接，那么则关闭次新的连接

			//3.3 处理新连接请求的业务方法，此时 handler 和 conn是绑定好的
			dealConn := NewConnection(conn, cid, s.Router)
			cid++

			//3.4 启动当前连接的业务处理
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server, name ", s.Name)

	//TODO Server.Stop() 将其他需要清理的连接信息或其他信息，一并停止并清理
}

func (s *Server) Serve() {
	// 启动server的服务功能
	s.Start()

	//TODO 启动服务器后的额外业务

	// 阻塞，保证主Go不会退出
	for {
		time.Sleep(10 * time.Second)
	}
}

// 路由功能：给当前服务注册一个路由业务方法，供客户端连接处理使用
func (s *Server) AddRouter(router ziface.IRouter) {
	s.Router = router
	fmt.Println("Add Router succ! ")
}
