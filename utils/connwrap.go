package utils

import (
	"context"
	"io"
	"net"
	"time"
)

type IConnWrap interface {
	Detach(ctx context.Context) *ConnWrap
}

type ConnWrap struct {
	conn    net.Conn
	reader  *ReadWrap
	context context.Context
	cancel  context.CancelFunc
}

func NewConnWrap(conn net.Conn, ctx context.Context) *ConnWrap {
	ctx, ccl := context.WithCancel(context.Background())
	cw := &ConnWrap{
		conn:    conn,
		reader:  NewReadWrap(conn),
		cancel:  ccl,
		context: ctx,
	}

	return cw
}

func (self *ConnWrap) Detach(ctx context.Context) *ConnWrap {
	ctx, ccl := context.WithCancel(context.Background())
	// 构建一个新的conn
	cw := &ConnWrap{
		conn:    self.conn,
		reader:  self.reader,
		cancel:  ccl,
		context: ctx,
	}

	// 调用关闭处理函数
	self.cancel()

	return cw
}

func (self *ConnWrap) Read(b []byte) (n int, err error) {
	return self.reader.Read(b, self.context)
}

func (self *ConnWrap) Write(b []byte) (n int, err error) {
	select {
	case _, _ = <-self.context.Done():
		return 0, io.EOF
	default:
		return self.conn.Write(b)
	}
}

func (self *ConnWrap) Close() error {
	select {
	case _, _ = <-self.context.Done():
		return io.EOF
	default:
		return self.conn.Close()
	}
}

func (self *ConnWrap) LocalAddr() net.Addr {
	select {
	case _, _ = <-self.context.Done():
		return nil
	default:
		return self.conn.LocalAddr()
	}
}

func (self *ConnWrap) RemoteAddr() net.Addr {
	select {
	case _, _ = <-self.context.Done():
		return nil
	default:
		return self.conn.RemoteAddr()
	}
}

func (self *ConnWrap) SetDeadline(t time.Time) error {
	select {
	case _, _ = <-self.context.Done():
		return io.EOF
	default:
		return self.conn.SetDeadline(t)
	}
}

func (self *ConnWrap) SetReadDeadline(t time.Time) error {
	select {
	case _, _ = <-self.context.Done():
		return io.EOF
	default:
		return self.conn.SetReadDeadline(t)
	}
}

func (self *ConnWrap) SetWriteDeadline(t time.Time) error {
	select {
	case _, _ = <-self.context.Done():
		return io.EOF
	default:
		return self.conn.SetWriteDeadline(t)
	}
}
