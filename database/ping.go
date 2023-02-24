package database

import (
	"29go_redis/interface/resp"
	"29go_redis/resp/reply"
)

func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// 初始化， 在启动的时候就会调用
// 这个就实际上 注册了（在其他语言中 这里就等于注册功能）
func init() {
	RegisterCommand("ping", Ping, 1)
}
