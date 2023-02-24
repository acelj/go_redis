package database

import "29go_redis/interface/resp"

type CmdLine = [][]byte

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(c resp.Connection)
}

// 这个可以实现任意数据类型，这阶段只实现string
type DataEntity struct {
	Data interface{}
}
