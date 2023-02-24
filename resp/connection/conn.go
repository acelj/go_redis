package connection

import (
	"29go_redis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

type Connection struct {
	conn         net.Conn
	waitingReply wait.Wait
	mu           sync.Mutex
	selectedDB   int
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn, // 其他就用默认初始化的值
	}
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	//panic("implement me")
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

func (c *Connection) Write(bytes []byte) error {
	//panic("implement me")

	if len(bytes) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()
	_, err := c.conn.Write(bytes)
	return err
}

func (c *Connection) GetDBIndex() int {
	//panic("implement me")
	return c.selectedDB
}

func (c *Connection) SelectDB(dbNum int) {
	//panic("implement me")
	c.selectedDB = dbNum
}
