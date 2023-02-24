package aof

import (
	"29go_redis/config"
	"29go_redis/interface/database"
	"29go_redis/lib/logger"
	"29go_redis/lib/utils"
	"29go_redis/resp/connection"
	"29go_redis/resp/parser"
	"29go_redis/resp/reply"
	"io"
	"os"
	"strconv"
)

type CmdLine = [][]byte

const aofBufferSize = 1 << 16 // 65535

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

type AofHandler struct {
	database    database.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	currentDB   int // 记录上一条工作在哪一个db中
}

// NewAofHandler
func NewAofHandler(database database.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.database = database
	// 加载LoadAof
	handler.LoadAof()
	// 第三个参数是 创建文件时以什么方式（权限去创建）
	// 下面的文件打开是在redis的整个生命周期中，不存在关闭，所以这里没有加上defer
	aofile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofile
	// chan 操作， 一定需要一个缓冲区的
	handler.aofChan = make(chan *payload, aofBufferSize)
	go func() {
		handler.handlerAof()
	}()
	return handler, nil
}

// Add payload(set k v)  -> aofChan
// 由chan 落盘， 直接落盘会很慢, 这里采用channel， 复习一下channel 用法
func (handler *AofHandler) AddAof(dbIndex int, cmd CmdLine) {
	// 检查有没有打开aof 功能
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmd,
			dbIndex: dbIndex,
		}
	}
}

// 落盘操作函数
// handlerAof payload(set k v) <- aofChan (落盘)
func (handler *AofHandler) handlerAof() {
	// TODO : payload(set k v) <- aofChan (落盘)
	handler.currentDB = 0

	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {
			data := reply.MakeMultiBulkRelpy(utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))).ToBytes()
			// 往文件中写data
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Error(err)
				continue
			}
			handler.currentDB = p.dbIndex
		}

		// 正常的逻辑
		// 把用户输入的指令变成resp 格式写到文件中
		data := reply.MakeMultiBulkRelpy(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Error(err)
		}
	}
}

// LoadAof
// 把aof 从硬盘中重新执行一遍的方法
func (handler *AofHandler) LoadAof() {
	// 怎么做将文件数据load进去， 只读方式打开文件， 简化版的Openfile
	file, err := os.Open(handler.aofFilename)
	if err != nil {
		logger.Error(err)
		return
	}
	// 这种只需要打开一次
	defer file.Close()
	ch := parser.ParserStream(file)
	fackConn := &connection.Connection{}
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error(p.Err)
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			// 类型转化为成功
			logger.Error("need multi mulk")
			continue
		}
		rep := handler.database.Exec(fackConn, r.Args)
		if reply.IsErrReply(rep) {
			logger.Error("exec err", rep.ToBytes())
		}
	}
}
