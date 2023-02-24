package cluster

import (
	"29go_redis/config"
	database2 "29go_redis/database"
	"29go_redis/interface/database"
	"29go_redis/interface/resp"
	"29go_redis/lib/consistenthash"
	"29go_redis/lib/logger"
	"29go_redis/resp/reply"
	"context"
	pool "github.com/jolestar/go-commons-pool/v2"
	"strings"
)

/**
引入github 外部开源组件
1. 在go.mod 文件中输入 require github.com/jolestar/go-commons-pool/v2 v2.1.2
2. 在项目目录下的终端输入 go get （貌似不起作用）

另一种直接方法
1. 在项目目录下的终端输入 go get github.com/jolestar/go-commons-pool/v2
等待下载就可以了
*/

/**
集群中的命令分三种模式
1. ping， 直接用单机版的redis即可
2. get、set 命令， 需要用一致性哈希找到具体执行单机中去
3. flushdb  清空数据库， 此时是集群中的数据库都需要清空
*/

// 作为转发端
type ClusterDatabase struct {
	self           string // 记录自己的名称的地址
	nodes          []string
	peerPicker     *consistenthash.NodeMap     // 引入一致性hash
	peerConnection map[string]*pool.ObjectPool // 因为每个节点对其他节点都需要有连接，这里就需要一个map管理
	db             database.Database           // 用这个接口就行了
}

func MakeClusterDatebase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:           config.Properties.Self,
		db:             database2.NewStandaloneDatabase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}

	// len(config.Properties.Peers) + 1 兄弟节点 + 1（自己）
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}

	nodes = append(nodes, config.Properties.Self)

	cluster.peerPicker.AddNode(nodes...)

	// 连接池初始化  peerConnection
	ctx := context.Background()

	// 默认新建 8个连接， 连接池中
	for _, peer := range config.Properties.Peers {
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})

	}

	cluster.nodes = nodes
	return cluster
}

type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply

var router = makeRouter()

func (cluster *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	//panic("implement me")
	// 替换了单机版的执行
	defer func() {
		// 这里需要保证集群的正常运行
		if err := recover(); err != nil {
			logger.Error(err)
			result = reply.UnknowErrReply{}
		}
	}()

	cmdName := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		// 不支持这个命令
		reply.MakeErrReply("not supported cmd")
	}
	result = cmdFunc(cluster, client, args)

	return
}

func (c *ClusterDatabase) Close() {
	//panic("implement me")
	// 关掉集群层
	c.db.Close()
}

func (cluster *ClusterDatabase) AfterClientClose(conn resp.Connection) {
	//panic("implement me")
	cluster.db.AfterClientClose(conn)
}
