package tcp

import (
	"29go_redis/interface/tcp"
	"29go_redis/lib/logger"
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

func ListenAndServeWithSignal(cfg *Config,
	handler tcp.Handler) error {

	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	// 上面的sigChan 是自己定义的， 下面的是需要告诉系统哪些信号需要告诉我
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	// 将 sigChan 和 closeChan 关联起来
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen..")

	// closeChan ： 是为了感知程序人为关闭掉，这时其实资源并没有被释放，增加这个参数实际上能从系统上
	// 给程序发送一个关闭的信号
	ListenAndServe(listener, handler, closeChan)

	return nil
}
func ListenAndServe(listener net.Listener,
	handler tcp.Handler,
	closeChan <-chan struct{}) {
	go func() {
		// 没有数据进来这个chan， 就会卡在这里哈
		<-closeChan
		// 如果有信号过来了， 执行下面的
		logger.Info("shutting down...")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	// 每服务一个新的客户端 + 1, 服务完后 - 1，错误退出后在等待
	var waitDone sync.WaitGroup

	for true {
		conn, err := listener.Accept()
		if err != nil {

			break
		}
		logger.Info("accept link...")

		// 新增客户端 + 1
		waitDone.Add(1)
		go func() {
			defer func() {
				// 服务完后 - 1
				waitDone.Done()
			}()
			handler.Handler(ctx, conn)

		}()
	}
	// 错误退出后， 或者所有的服务完的客户端退出后，会卡在这里
	waitDone.Wait()
}
