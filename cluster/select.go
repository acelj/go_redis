package cluster

import "29go_redis/interface/resp"

func execSelect(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	return cluster.db.Exec(c, cmdArgs)
}
