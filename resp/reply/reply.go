package reply

import (
	"29go_redis/interface/resp"
	"bytes"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")
	CRLF               = "\r\n"
)

// 在redis 中bulk 就表示字符串， 块
type BulkReply struct {
	Arg []byte // 回复 ： liudehua   "$8\r\nliudehua\r\n"
}

// 转化成resp 格式字符串
func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(b.Arg)) + CRLF + string(b.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// 多字符串回复
type MultiBulkReply struct {
	Args [][]byte
}

func (m *MultiBulkReply) ToBytes() []byte {
	argLen := len(m.Args)
	//logger.Info("===%v", argLen)
	// 用 buffer 做字符串拼装
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range m.Args {
		if arg == nil {
			buf.WriteString(string(nullBulkReplyBytes) + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
			//fmt.Printf("====%v", arg)
		}
	}
	return buf.Bytes()
}

func MakeMultiBulkRelpy(arg [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: arg}
}

// 回复一个状态
type StatusReply struct {
	Status string
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{Status: status}
}

func (s *StatusReply) ToBytes() []byte {
	return []byte("+" + s.Status + CRLF)
}

// 通用数字回复
type IntReply struct {
	Code int64
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

func (i *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 10) + CRLF)
}

// 这个接口相当于实现了两个接口
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// 一般化的错误回复, 自定义错误回复
type StandardErrReply struct {
	Status string
}

func (s *StandardErrReply) Error() string {
	return s.Status
}

func (s *StandardErrReply) ToBytes() []byte {
	return []byte("-" + s.Status + CRLF)
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{Status: status}
}

// 判断是否是正常回复
func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
