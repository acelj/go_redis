package database

import (
	"29go_redis/datastruct/dict"
	"29go_redis/interface/database"
	"29go_redis/interface/resp"
	"29go_redis/lib/logger"
	"29go_redis/resp/reply"
	"strings"
)

type DB struct {
	index int
	data  dict.Dict // 未来可以直接换实现的数据结构
	// 添加一个落盘函数
	addAof func(CmdLine)
}

type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine = [][]byte

func makeDB() *DB {
	db := &DB{
		data:   dict.MakeSyncDict(),
		addAof: func(line CmdLine) {}, // 防止LoadAof 时，这里是未初始化的函数，出错
	}
	return db
}

func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	// 指令可能有 PING， SET， SETNX
	cmdName := strings.ToLower(string(cmdLine[0]))
	// 下面的map 只是只读的，不涉及并发问题
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command " + cmdName)
	}
	if !validateArity(cmd.arity, cmdLine) {
		logger.Error("arity =%v, len(cmdArgs) = %v  ", cmd.arity, len(cmdLine))
		return reply.MakeArgNumErrReply(cmdName)
	}

	fun := cmd.exector
	// SET k v  -- > k v
	return fun(db, cmdLine[1:])
}

// 校验用户发送指令的参数数量
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	//logger.Info("arity =%v, len(cmdArgs) = %v  ", arity, len(cmdArgs))
	//logger.Info(string(cmdArgs[0]))
	//logger.Info(string(cmdArgs[1]))
	if argNum >= 0 {
		return argNum == arity
	}
	return argNum >= -arity
}

// GET k
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (db *DB) Flush() {
	db.data.Clear()
}
