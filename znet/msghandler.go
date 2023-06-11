package znet

import (
	"fmt"
	"github.com/jiangh156/zinx/utils"
	"github.com/jiangh156/zinx/ziface"
	"strconv"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter // 存放每个MsgId所对应的处理方法的map属性
	WorkerPoolSize uint32                    //业务工作Worker池的数量
	TaskQueue      []chan ziface.IRequest    //Worker负责取任务的消息队列
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue: make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

//============== 实现 ziface.IMsgHandler 里的全部接口方法 ========

// 启动一个worker，阻塞等待任务
func (m *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started")
	//不断等待队列任务
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			m.DoMsgHandler(request)
		}
	}
}

// 启动worker工作池
func (m *MsgHandle) StartWorkPool() {
	// 遍历需要启动worker的数量，依次开启
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		// 一个worker启动
		// 给当前worker对应的任务队列开辟空间
		m.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//启动当前worker，阻塞地等待对应的任务队列是否由消息传递进来
		go m.StartOneWorker(i, m.TaskQueue[i])
	}
}

// 将消息交给TaskQueue，由worker进行处理
func (m *MsgHandle) SendMsgToQueue(request ziface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法制
	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % m.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(), " request msgID=", request.GetMsgID(),
		"to workerID=", workerID)
	//将请求信息发送给任务队列
	m.TaskQueue[workerID] <- request
}

// 马上以非阻塞方式处理消息
func (m *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := m.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is NOT FOUND! Need Register!")
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
