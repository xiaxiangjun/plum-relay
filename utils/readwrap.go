package utils

import (
	"bufio"
	"context"
	"io"
	"sync"
)

type ReadWrap struct {
	locker sync.Mutex
	reader *bufio.Reader
}

func NewReadWrap(reader io.Reader) *ReadWrap {
	return &ReadWrap{
		reader: bufio.NewReader(reader),
	}
}

func (self *ReadWrap) Read(buf []byte, ctx context.Context) (int, error) {
	self.locker.Lock()
	defer self.locker.Unlock()

	// 读取数据
	out, err := self.reader.Peek(len(buf))
	if nil != err {
		return 0, err
	}

	select {
	case _, _ = <-ctx.Done():
		return 0, io.EOF
	default:
		self.reader.Discard(len(out))
		copy(buf[:len(out)], out)
		return len(out), nil
	}
}
