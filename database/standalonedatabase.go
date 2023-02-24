package database

import (
	"29go_redis/aof"
	"29go_redis/config"
	"29go_redis/interface/resp"
	"29go_redis/lib/logger"
	"29go_redis/resp/reply"
	"strconv"
	"strings"
)

type StandaloneDatabase struct {
	dbSet      []*DB // 一组DB指针
	aofHandler *aof.AofHandler
}

// 就是新建一个Database结构体， 然后把默认16个db 初始化
func NewStandaloneDatabase() *StandaloneDatabase {
	// 通过下面的代码可以读取配置文件中的配置
	//config.Properties.Databases

	database := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		db := makeDB()
		db.index = i // i 作为计数器
		database.dbSet[i] = db
	}
	// 初始化aof功能
	// 查看配置是否打开
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
		for _, db := range database.dbSet {
			sdb := db
			sdb.addAof = func(line CmdLine) {
				// 调用了下面中的方法
				// 这里出现闭包的问题
				// db = dbSet[0], 这里内部变量就逃逸到了堆上
				// db = dbSet[15]
				database.aofHandler.AddAof(sdb.index, line)
			}
		}
	}

	return database
}

// set k v
// get k
// select 2
func (database *StandaloneDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	// 执行用户指令
	// 其实这里的就是把用户指令转到分db去执行，
	// 这里需要做一个recover， 因为这里有可能出现崩溃
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	// 这里只涉及一个 select 命令， 剩下都只用交给分db
	cmdName := strings.ToLower(string(args[0]))
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, database, args[1:])
	}

	// 一般的命令
	dbIndex := client.GetDBIndex()
	db := database.dbSet[dbIndex]
	return db.Exec(client, args)
}

func (database *StandaloneDatabase) Close() {

}

func (database *StandaloneDatabase) AfterClientClose(c resp.Connection) {

}

// 用户切换db执行的逻辑， 用户需要选择某个db
// select 2
func execSelect(c resp.Connection, database *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		// select a
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(database.dbSet) {
		// 从0开始编号的额
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
