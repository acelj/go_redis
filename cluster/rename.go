package cluster

import (
	"29go_redis/interface/resp"
	"29go_redis/resp/reply"
)

// 保留在同一个节点才能做rename， 在不同的节点不能做rename
// rename k1 k2
func Rename(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	if len(cmdArgs) != 3 {
		return reply.MakeErrReply("ERR Wrong number args")
	}
	src := string(cmdArgs[1]) // k1
	dest := string(cmdArgs[2])

	srcPeer := cluster.peerPicker.PickNode(src) // 节点 192.168.xx.xx：6379
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within on peer")
	}

	return cluster.relay(srcPeer, c, cmdArgs)
}
