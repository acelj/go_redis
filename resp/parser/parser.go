package parser

import (
	"29go_redis/interface/resp"
	"29go_redis/lib/logger"
	"29go_redis/resp/reply"
	"bufio"
	"errors"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

// redis 解析器

type Payload struct {
	Data resp.Reply // 客户端和服务端发送的消息数据结果都一样，所以这里可以用reply表示
	Err  error
}

type readState struct {
	// 解析单行
	readingMultiLine bool
	// 记录多少长度
	expectedArgsCount int
	// 消息类型
	msgType byte
	// 用户传过来的数据
	args [][]byte
	// 数据块的长度
	bulkLen int64
}

// 看解析有没完成
func (s *readState) finished() bool {
	// 解析的长度和期望的数量相同
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// 异步, 大写, 开放给外面
// 通过管道异步给你，不需要等待
// 把解析的结果通过管道给出去，不卡在这里
func ParserStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parser0(reader, ch)
	return ch
}

// 解析器的核心
// io.Reader 上层传给这里的就是io流， 这里的方法是小写，需要开放一个函数接口给上层
func parser0(reader io.Reader, ch chan<- *Payload) {
	// 防止里面出现panic 后，程序直接宕掉
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()

	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte

	for true {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr {
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			ch <- &Payload{
				Err: err,
			}
			// 这里就是协议不规范的错误，直接continue
			state = readState{}
			continue

		}
		// 是不是多行解析模式
		// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
		if !state.readingMultiLine {
			if msg[0] == '*' { // *3\r\n
				err := parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' { // $3\r\n 单个字符串 ($4\r\nPING\r\n、 $-1\r\n)
				err := parseBulkHeader(msg, &state)
				if err != nil {
					// 告诉上层协议错误
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 { // $-1\r\n
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{}
					continue
				}
			} else { // : + - 这种情况
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else {
			// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
			// 现在需要处理这个字符串 $3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
			err := readBody(msg, &state)
			if err != nil {
				// 告诉上层协议错误
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{}
				continue
			}
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkRelpy(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0]) // 单行的只传递第一行
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}
}

// 实现几个辅助的函数
// 以\n 结尾的才是一行指令，
// 返回值： 返回内容， bool: 是指io 错误， error： 错误本身
// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// 内容中可能用\r\n.
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	// 1. \r\n 切分
	// 2. 之前读到了$ 数字，严格读取字符个数
	var msg []byte
	var err error

	if state.bulkLen == 0 { // 1. \r\n 切分
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		// 判断\n 前面是不是\r
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			// 这里不是io 错误， 第二个参数给false
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { // 2. 之前读到了$ 数字，严格读取字符个数
		msg = make([]byte, state.bulkLen+2) //加上\r\n
		_, err := io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// 把解析器设置成多行模式
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64

	// *3\r\n : 只取出中间的数字
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true // 现在是多行状态
		state.expectedArgsCount = int(expectedLine)
		// 前面解析*3, 这里后面就有3个参数
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// 解析单行数据
// $4\r\nPING\r\n
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 32)

	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 {
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true // 现在是多行状态
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// 有时候客户端会给服务端发送   +OK  -err
// 是固定的， 所以可以解析完
// +OK\r\n   -err\r\n   :5\r\n
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	// 切掉后面\r\n
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply // 这个类型是双向的数据结构

	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:]) // 切掉 +
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}

		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// 解析body
/**
*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
-经过上面处理后： $3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
$4\r\nPING\r\n
-经过上面处理后： PING\r\n
就是上面这种，剩下后面的内容部分
*/
func readBody(msg []byte, state *readState) error {
	// 去掉后面\r\n
	line := msg[0 : len(msg)-2]

	var err error

	// $3
	if line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 32)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		// $0\r\n
		if state.bulkLen <= 0 {
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		// 非 $
		state.args = append(state.args, line)
	}

	return nil
}
