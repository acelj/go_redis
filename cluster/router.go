package cluster

import "29go_redis/interface/resp"

func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["exists"] = defaultFunc // exists k1
	routerMap["type"] = defaultFunc
	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getset"] = defaultFunc
	routerMap["ping"] = ping
	routerMap["rename"] = Rename
	// renamenx 判断的逻辑跟这一层逻辑无关
	routerMap["renamenx"] = Rename
	routerMap["flushdb"] = flushdb
	routerMap["del"] = Del
	routerMap["select"] = execSelect

	return routerMap
}

// GET key  //  set k v
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	key := string(cmdArgs[1])
	peer := cluster.peerPicker.PickNode(key)
	return cluster.relay(peer, c, cmdArgs)
}
