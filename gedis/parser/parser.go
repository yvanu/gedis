package parser

import (
	"bufio"
	"bytes"
	"errors"
	"gedis/gedis/proto"
	"gedis/tool/logger"
	"io"
	"runtime/debug"
	"strconv"
)

type Payload struct {
	Err   error
	Reply proto.Reply
}

func ParseStream(reader io.Reader) <-chan *Payload {
	stream := make(chan *Payload)
	go parse(reader, stream)
	return stream
}

func parse(r io.Reader, stream chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err, string(debug.Stack()))
		}
	}()

	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			stream <- &Payload{Err: err}
			close(stream)
			return
		}

		length := len(line)
		if length <= 2 || line[length-2] != '\r' {
			continue
		}

		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		switch line[0] {
		case '*':
			// 数组  先读出来一行判断是什么类型，然后在对应的解析函数中接着读
			err := parseArrays(line, reader, stream)
			if err != nil {
				stream <- &Payload{Err: err}
				close(stream)
				return
			}
		case '+':
			stream <- &Payload{
				Reply: proto.NewSimpleStringReply(string(line[1:])),
			}
		case '-':
			stream <- &Payload{
				Reply: proto.NewSimpleErrReply(string(line[1:])),
			}
		case '$':
			err = parseBulkString(line, reader, stream)
			if err != nil {
				stream <- &Payload{Err: err}
				close(stream)
				return
			}
		default:
			args := bytes.Split(line, []byte{' '})
			stream <- &Payload{
				Reply: proto.NewMultiBulkReply(args),
			}
		}
	}
}

// $5\r\nvalue\r\n
func parseBulkString(header []byte, reader *bufio.Reader, out chan<- *Payload) error {
	num, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || num < -1 {
		out <- &Payload{Err: errors.New("invalid bulk string header" + string(header[1:]))}
		return nil
	} else if num == -1 {
		out <- &Payload{Reply: proto.NewNullBulkReply()}
		return nil
	}

	body := make([]byte, num+2)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	out <- &Payload{Reply: proto.NewBulkReply(body[:num])}
	return nil
}

func parseArrays(header []byte, reader *bufio.Reader, out chan<- *Payload) error {
	/*
		数组格式：

		*2\r\n
		$5\r\n
		hello\r\n
		$5\r\n
		world\r\n

	*/
	num, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || num < 0 {
		out <- &Payload{Err: errors.New("invalid array header" + string(header[1:]))}
		return nil
	}
	lines := make([][]byte, 0, num)
	for i := int64(0); i < num; i++ {
		var line []byte
		line, err = reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		length := len(line)
		if length < 4 || line[length-2] != '\r' || line[0] != '$' {
			out <- &Payload{Err: errors.New("invalid bulk string header" + string(line))}
			return nil
		}
		dataLen, err := strconv.ParseInt(string(line[1:length-2]), 10, 64)
		if err != nil || dataLen < -1 {
			out <- &Payload{Err: errors.New("invalid bulk string length" + string(line))}
			return nil
		} else if dataLen == -1 {
			lines = append(lines, nil)
		} else {
			// 2 是 \r\n
			body := make([]byte, dataLen+2)
			_, err = io.ReadFull(reader, body)
			if err != nil {
				return err
			}
			lines = append(lines, body[:dataLen])
		}
	}
	out <- &Payload{Reply: proto.NewMultiBulkReply(lines)}
	return nil
}

func parseError(out chan<- *Payload, err string) {

}
