package tcp

import (
	"29go_redis/lib/logger"
	"29go_redis/lib/sync/atomic"
	"29go_redis/lib/sync/wait"
	"bufio"
	"context"
	"io"
	"net"
	"sync"
	"time"
)

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait // 自己包装的wai（因为自己实现有超时功能）
}

func (e *EchoClient) Close() error {
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

// 业务引擎
type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

func MakeHandler() *EchoHandler {
	return &EchoHandler{} // 里面的不需要初始化
}

// 实现之前接口方法， 选择结构体， ctrl + i  --> 输入需要实现的方法
func (handler *EchoHandler) Handler(ctx context.Context, conn net.Conn) {
	if handler.closing.Get() {
		_ = conn.Close()
	}

	client := &EchoClient{
		Conn: conn, // wait 不初始化时， 系统会自动初始化为0
	}

	handler.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n') // 以\n 为标志位
		logger.Info("read client : ", msg)
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close...")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		// 告诉上层在左业务（回发），在未做完之前不要关掉我，可以等10s，在关掉
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		logger.Info("write client : ", msg)
		client.Waiting.Done()
	}

}

func (handler *EchoHandler) Close() error {
	logger.Info("handler  shutting down...")
	// 设置业务引擎状态为true，在收到新的连接，就不做服务了
	handler.closing.Set(true)

	// 关闭所有客户端, Map 干掉， 普通的map直接 遍历就行了， 下面是sync 的map
	handler.activeConn.Range(func(k, v interface{}) bool {
		// Range 中的匿名函数参数解释
		// k, v : 是对 值的操作
		// bool : 为true 是操作完这个元素后继续操作下一个元素，false： 是不继续操作下一个元素
		// 这个匿名函数就是对sync 包中map 每个进行关闭 close
		client := k.(*EchoClient) // 这里的k 是空接口，需要进行转换类型
		_ = client.Conn.Close()
		return true
	})
	return nil
}
