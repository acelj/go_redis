package tcp

import (
	"context"
	"net"
)

// 这个接口可以忽略tcp 这一层处理业务逻辑， 将业务逻辑的交给具体的业务去实现
type Handler interface {
	Handler(ctx context.Context, conn net.Conn)
	Close() error
}
