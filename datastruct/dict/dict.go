package dict

type Consumer func(key string, val interface{}) bool

// 写一个接口的，这样后续想迭代就直接重写就行了，如果直接实现的话，后续迭代都的改动
type Dict interface {
	Get(key string) (val interface{}, exists bool)
	Len() int
	// 把k, v 存进去, 返回存进去几个
	Put(key string, val interface{}) (result int)
	// redis 中的 setnx 命令
	PutIfAbsent(key string, val interface{}) (result int)
	PutIfExists(key string, val interface{}) (result int)
	Remove(key string) (result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
}
