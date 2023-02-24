package database

import (
	"29go_redis/interface/database"
	"29go_redis/interface/resp"
	"29go_redis/lib/utils"
	"29go_redis/resp/reply"
)

/**
实现几个基本指令，后续的可以根据自定义的需求去实现
GET
SET
SETNX
GETSET
STRLEN
*/

//GET
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exist := db.GetEntity(key)
	if !exist {
		return reply.MakeNullBulkReply()
	}

	// 如果实现其他类型，这里需要判断类型转换有没有成功, 这里默认成功的
	bytes := entity.Data.([]byte)
	return reply.MakeBulkReply(bytes)
}

//SET k1 v
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	val := args[1]
	entity := &database.DataEntity{
		Data: val,
	}
	db.PutEntity(key, entity)
	db.addAof(utils.ToCmdLine3("set", args...))
	return reply.MakeOkReply()
}

//SETNX k1 v1( 检查k1 是否存在， 返回值是0或1)
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	val := args[1]
	//// 检查k1 是否存在
	//_, ok := db.GetEntity(key)
	//if ok {
	//	return reply.MakeIntReply(0) // 存在就什么都不干
	//}
	//entity := &database.DataEntity{
	//	Data: val,
	//}
	//db.PutEntity(key, entity)
	//return reply.MakeIntReply(1)

	entity := &database.DataEntity{
		Data: val,
	}
	result := db.PutIfAbsent(key, entity)
	db.addAof(utils.ToCmdLine3("setnx", args...))
	return reply.MakeIntReply(int64(result))
}

//GETSET k1 v1  (先获取k1 原来的值，再将它改为v1， 返回 k1原来的值)
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	val := args[1]
	entity, exist := db.GetEntity(key)
	db.PutEntity(key, &database.DataEntity{Data: val})
	db.addAof(utils.ToCmdLine3("getset", args...))
	if !exist {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeBulkReply(entity.Data.([]byte))
}

//STRLEN (查看k 对应val的长度)  k ->'value'
func execStrLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exist := db.GetEntity(key)
	if !exist {
		return reply.MakeNullBulkReply()
	}
	bytes := entity.Data.([]byte)
	return reply.MakeIntReply(int64(len(bytes)))
}

func init() {
	RegisterCommand("Get", execGet, 2) // get k1
	RegisterCommand("SET", execSet, 3) // set k v
	RegisterCommand("SETNX", execSetNX, 3)
	RegisterCommand("GETSET", execGetSet, 3)
	RegisterCommand("STRLEN", execStrLen, 2)
}
