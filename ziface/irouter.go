package ziface

// IRouter 路由抽象接口，这是给路由框架使用者连接自定处理业务方法
// 路由的IRequest 包含了该连接的连接信息和该连接的请求数据信息
type IRouter interface {
	PreHandle(request IRequest)  //处理conn业务之前的钩子方法
	Handle(request IRequest)     // 处理conn业务的方法
	PostHandle(request IRequest) // 处理conn业务之后的钩子方法
}
