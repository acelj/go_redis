package database

import (
	"29go_redis/interface/resp"
	"29go_redis/lib/utils"
	"29go_redis/lib/wildcard"
	"29go_redis/resp/reply"
)

/**
DEL
EXISTS
KETS
FLUSHDB
TYPE
RENAME
RENAMENX
KEYS
*/

// DEL k1 k2 k3
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))

	for i, v := range keys {
		keys[i] = string(v)
	}
	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLine3("del", args...))
	}
	return reply.MakeIntReply(int64(deleted))
}

// EXISTS k1 k2 k3
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

//FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLine3("flushdb", args...))
	return reply.MakeOkReply()
}

// TYPE k1 查看k1的类型
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		// 不存在
		return reply.MakeStatusReply("none") // TCP : none\r\n
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
		// TODO
		// 下面如何实现其他的数据类型可以在这里进行定义
	}

	return &reply.UnknowErrReply{}
}

//RENAME k1 k2  [k1:v] -- > [k2:v]
func execRename(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])
	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine3("rename", args...))
	return reply.MakeOkReply()
}

//RENAMENX  (检查一下k1 改成 k2 , k2 会不会已经存在，如果存在，就是覆盖掉之前的map了)
func execRenamenx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])
	// 检查k2 是否存在
	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0) // 存在就什么都不干
	}
	// 下面的逻辑都是一样了
	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine3("renamenx", args...))
	return reply.MakeIntReply(1) // 这个函数成功是返回1的， 跟上面的不同
}

//KEYS *
func execKeys(db *DB, args [][]byte) resp.Reply {
	// 因为到这里的都是已经切掉前面的关键字的， 这里直接取出后面的通配符就行了
	pattern := wildcard.CompilePattern(string(args[0]))

	// 二维字节的切片存放用来存放所有符合通配符的key，然后返回回去
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			// 如果符合，就追加到map中
			result = append(result, []byte(key))
		}
		return true
	})

	return reply.MakeMultiBulkRelpy(result)
}

func init() {
	// 变长就是 -， 最少2个参数
	RegisterCommand("DEL", execDel, -2)
	RegisterCommand("EXISTS", execExists, -2)

	// 实际上只有一个参数， 这里填写-1 只是忽略后面的参数
	RegisterCommand("FLUSHDB", execFlushDB, -1)  // FLUSH a b c
	RegisterCommand("TYPE", execType, 2)         // TYPE k1
	RegisterCommand("RENAME", execRename, 3)     // RENAME k1 k2
	RegisterCommand("RENAMENX", execRenamenx, 3) // RENAMENX k1 k2
	RegisterCommand("KEYS", execKeys, 2)         // KEYS *
}
