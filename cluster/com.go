package cluster

import (
	"29go_redis/interface/resp"
	"29go_redis/lib/utils"
	"29go_redis/resp/client"
	"29go_redis/resp/reply"
	"context"
	"errors"
	"strconv"
)

func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New("connection not found111")
	}
	// 从对象池中借一个
	object, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	c, ok := object.(*client.Client) // 类型转化
	if !ok {
		return nil, errors.New("wrong type")
	}

	return c, err
}

// 从对象池中还对象，不还会造成对象池耗尽
func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection not found222")
	}
	return pool.ReturnObject(context.Background(), peerClient)
}

// 这里处理了两种情况，自己和选择兄弟节点
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self {
		// 如果是自己的情况
		return cluster.db.Exec(c, args)
	}
	// 操作集群中其他节点的情况
	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}

	defer func() {
		_ = cluster.returnPeerClient(peer, peerClient)
	}()

	// 转发其他集群中节点时，需要给出选择的几号db
	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex())))

	return peerClient.Send(args)
}

// 第三种情况：广播指令， 比如 flushdb
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	results := make(map[string]resp.Reply)

	// 调用所有节点就行了
	for _, node := range cluster.nodes {
		result := cluster.relay(node, c, args)
		results[node] = result
		//results = append(results, result)
	}
	return results
}
