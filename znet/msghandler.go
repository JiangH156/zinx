package znet

import (
	"fmt"
	"github.com/jiangh156/zinx/ziface"
	"strconv"
)

type MsgHandle struct {
	Apis map[uint32]ziface.IRouter // 存放每个MsgId所对应的处理方法的map属性
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis: make(map[uint32]ziface.IRouter),
	}
}

//============== 实现 ziface.IMsgHandler 里的全部接口方法 ========

// 马上以非阻塞方式处理消息
func (m *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := m.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgId())
		return
	}
	//执行对应的处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (m *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := m.Apis[msgId]; ok {
		panic("repeated api, msgId = " + strconv.Itoa(int(msgId)))
	}
	//2 添加msg与api的绑定关系
	m.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}
