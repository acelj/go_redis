package database

import (
	"29go_redis/interface/resp"
	"29go_redis/resp/reply"
)

// 做测试用的database 客户端发送什么指令，这里返回什么指令
type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	//panic("implement me")
	//logger.Info("Exec: %v", args)
	return reply.MakeMultiBulkRelpy(args)
}

func (e EchoDatabase) Close() {
	//panic("implement me")
}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	//panic("implement me")
}
