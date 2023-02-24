package handler

import (
	"29go_redis/cluster"
	"29go_redis/config"
	"29go_redis/database"
	databaseface "29go_redis/interface/database"
	"29go_redis/lib/logger"
	"29go_redis/lib/sync/atomic"
	"29go_redis/resp/connection"
	"29go_redis/resp/parser"
	"29go_redis/resp/reply"
	"context"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	db         databaseface.Database
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databaseface.Database
	//db = database.NewEchoDatabase()  // 上面是测试
	//db = database.NewStandaloneDatabase()
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		// 这个才是启动集群版
		db = cluster.MakeClusterDatebase()
	} else {
		// 单机版的实现
		db = database.NewStandaloneDatabase()
	}

	// TODO: 实现Database
	return &RespHandler{
		db: db,
	}

}

// 关闭一个client
func (r *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}

func (r *RespHandler) Handler(ctx context.Context, conn net.Conn) {
	//panic("implement me")
	if r.closing.Get() {
		_ = conn.Close()
	}
	client := connection.NewConn(conn)
	r.activeConn.Store(client, struct{}{})

	ch := parser.ParserStream(conn)

	for payload := range ch {
		// error
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				r.closeClient(client)
				logger.Info("connection close: " + client.RemoteAddr().String())
				return
			}
			// 协议错误
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes()) // 回写
			// 回写错误， 可能网络原因，回写不成功
			if err != nil {
				r.closeClient(client)
				logger.Info("connection close: " + client.RemoteAddr().String())
				return
			}
			// 回写成功
			continue
		}
		// exec
		if payload.Data == nil {
			continue
		}
		// 转成需要的格式
		reply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("===require multi bulk reply")
			continue
		}
		//logger.Info("reply : %v", reply.Args)
		if r == nil {
			panic("1")
		}
		if r.db == nil {
			panic("2")
		}

		result := r.db.Exec(client, reply.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

func (r *RespHandler) Close() error {
	//panic("implement me")
	logger.Info("handler shutting down")
	r.closing.Set(true)
	r.activeConn.Range(
		func(key, value interface{}) bool {
			client := key.(*connection.Connection)
			_ = client.Close()
			return true
		})
	r.db.Close()
	return nil
}
