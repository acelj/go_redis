package reply

// 这里写的一些固定的异常回复

// 未知错误
type UnknowErrReply struct {
}

var unknowErrBytes = []byte("-Err unknown\r\n")

func (u UnknowErrReply) Error() string {
	return "Err unknown"
}

func (u UnknowErrReply) ToBytes() []byte {
	return unknowErrBytes
}

// ArgNumErr
type ArgNumErrReply struct {
	Cmd string
}

func (r *ArgNumErrReply) Error() string {
	return "-Err wrong number of arguments for " + r.Cmd + "comand\r\n"
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-Err wrong number of arguments for '" + r.Cmd + "' comand\r\n")
}

func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{
		Cmd: cmd,
	}
}

// 语法错误
type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-Err syntax error\r\n")
var theSyntaxErrReply = &SyntaxErrReply{}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

func (s *SyntaxErrReply) Error() string {
	return "Err syntax error"
}

func (s *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

// 错误数据类型
type WrongTypeErrReply struct{}

var wrongTypeErrReply = []byte("-WRONGTYPE Operation against a key holding the wrong king of value")

func (w *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong king of value"
}

func (w *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrReply
}

// 协议不规范， 就是前面的不规范
type ProtocolErrReply struct {
	Msg string
}

func (p *ProtocolErrReply) Error() string {
	return "ERR Protocol error: '" + p.Msg
}

func (p *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + p.Msg + "'\r\n")
}
