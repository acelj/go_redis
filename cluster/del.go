package cluster

import (
	"29go_redis/interface/resp"
	"29go_redis/resp/reply"
)

// del k1 k2 k3 k4 k5
func Del(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	// 直接广播命令，但需要吧所以的回复汇总
	replies := cluster.broadcast(c, cmdArgs)

	var errReply reply.ErrorReply
	var deleted int64 = 0
	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
		intReply, ok := r.(*reply.IntReply)
		if !ok {
			errReply = reply.MakeErrReply("error")
		}
		deleted += intReply.Code
	}
	if errReply == nil {
		// 返回就OK了
		return reply.MakeIntReply(deleted)
	}

	return reply.MakeErrReply("error : " + errReply.Error())
}
