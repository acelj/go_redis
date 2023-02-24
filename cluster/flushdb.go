package cluster

import (
	"29go_redis/interface/resp"
	"29go_redis/resp/reply"
)

// 这个命令是广播
func flushdb(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	replies := cluster.broadcast(c, cmdArgs)

	// 我们这里规范map中有一个错误就是执行不成功
	var errReply reply.ErrorReply
	for _, r := range replies {
		if reply.IsErrReply(r) {
			// 遇到一个错误就不用遍历了
			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return reply.MakeOkReply()
	}

	return reply.MakeErrReply("error : " + errReply.Error())
}
