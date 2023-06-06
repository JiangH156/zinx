package znet

import "github.com/jiangh156/zinx/ziface"

type Request struct {
	conn ziface.IConnection // 已经和客户端建立的连接
	msg  ziface.IMessage    // 客户端请求的数据
}

//============== 实现 ziface.IRequest 里的全部接口方法 ========

// 获取消息ID
func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgId()
}

// 获取请求连接信息
func (r *Request) GetConnection() ziface.IConnection {
	return r.conn
}

// 获取请求信息的数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}
