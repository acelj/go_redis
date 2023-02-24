package cluster

import "29go_redis/interface/resp"

// 本地执行
func ping(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	return cluster.db.Exec(c, cmdArgs)
}
