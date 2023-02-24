package reply

// 这里写的一些固定的正常回复

// PONG 回复
type PongReply struct {
}

var pongbytes = []byte("+PONG\r\n")

func (p PongReply) ToBytes() []byte {
	return pongbytes
}

// 很多编程风格保留这个make 方法
func MakePongReply() *PongReply {
	return &PongReply{}
}

// 固定回复OK
type OkReply struct{}

var okReply = []byte("+OK\r\n")

func (r *OkReply) ToBytes() []byte {
	//panic("implement me")
	return okReply
}

// 小优化： 先持有， 当作常量，这样在每次make 就不用重新生成了
var theOkReply = new(OkReply)

func MakeOkReply() *OkReply {
	return theOkReply
}

type NullBulkReply struct {
}

// 这个是null， 和“” 表示的有点区别
var nullBulkBytes = []byte("$-1\r\n")

func (n NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

// 这个是回复空数组
var emptyMultiBulkBytes = []byte("*0\r\n")

type EmptyMultiBulkReply struct{}

func (e *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

// 空 ： 什么都没有
type NoReply struct {
}

var noBytes = []byte("")

func (n NoReply) ToBytes() []byte {
	return noBytes
}
