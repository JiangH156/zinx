package utils

import (
	"encoding/json"
	"github.com/jiangh156/zinx/ziface"
	"os"
)

/*
存储一切有关Zinx框架的全局参数，供其他模块使用
一些参数也可以通过用户根据zinx.json配置
*/
type GlobalObj struct {
	/*
		Server
	*/
	TcpServer ziface.IServer // Zinx的全局Server对象
	Host      string         //服务器主机IP
	TcpPort   int            //服务器主机监听端口号
	Name      string         //服务器名称

	/*
		Zinx
	*/
	Version          string //Zinx版本号
	MaxPacketSize    uint32 // 数据包最大值
	MaxConn          int    //服务器主机允许的最大连接个数
	WorkerPoolSize   uint32 //业务工作Worker池的数量
	MaxWorkerTaskLen uint32 //业务工作Worker对应负责的任务队列最大任务存储的数量
	MaxMsgChanLen    uint32 //业务工作中goroutine之间消息通信长度
	/*
		config file path
	*/
	ConfFilePath string //配置文件路径
}

// 读取用户的配置文件
func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}
	//将json数据解析到struct中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

// 定义一个全局对象
var GlobalObject *GlobalObj

// init 默认加载
func init() {
	// 初始化GlobalObject变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "V0.10",
		TcpPort:          7777,
		Host:             "0.0.0.0",
		MaxConn:          12000,
		MaxPacketSize:    4096,
		ConfFilePath:     "conf/zinx.json",
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
	}
	//从配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}
