package znet

import "github.com/jiangh156/zinx/ziface"

// BaseRouter 实现router时，先嵌入这个基类，然后根据需要对这个基类的方法进行重写
type BaseRouter struct{}

//============== 实现 ziface.IRouter 里的全部接口方法 ========

func (b *BaseRouter) PreHandle(request ziface.IRequest) {
}

func (b *BaseRouter) Handle(request ziface.IRequest) {
}

func (b *BaseRouter) PostHandle(request ziface.IRequest) {
}
