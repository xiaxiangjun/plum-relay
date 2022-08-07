package protocol

import (
	"context"
	"fmt"
	"net"
	"plum-relay/utils"
	"sync"
	"sync/atomic"
	"time"
)

// stream事件回调
type StreamEvent interface {
	OnRequrest(stream *RelayStream, req *Frame) (res *Frame, err error)
}

type RelayStream struct {
	locker       sync.Mutex
	conn         net.Conn
	event        StreamEvent
	streamID     uint32
	waitResponse utils.SyncMap[uint32, chan<- *Frame]
}

// 创建一个新的stream
func NewRelayStream(conn net.Conn, event StreamEvent) *RelayStream {
	stream := &RelayStream{
		conn:     conn,
		event:    event,
		streamID: 1000,
	}

	// 初始化stream
	stream.init()

	return stream
}

func (self *RelayStream) init() {

}

func (self *RelayStream) Serve() error {
	for {
		// 设置读取超时
		self.conn.SetReadDeadline(time.Now().Add(time.Second * 15))
		// 读取一帧数据
		frame, err := ReadFrame(self.conn)
		if nil != err {
			return err
		}

		if FrameRequest == frame.Method() {
			// 请求消息
			go self.onRequest(frame)
		} else if FrameResponse == frame.Method() {
			// 回复消息
			go self.onResponse(frame)
		}
	}
}

// 处理一个请求
func (self *RelayStream) onRequest(req *Frame) {
	if nil == self.event {
		return
	}

	res, err := self.event.OnRequrest(self, req)
	if nil != err {
		res = NewFrame(FrameResponse, req.StreamID())
		res.SetFlag(FlagError)
		res.SetBody([]byte(err.Error()))
	}

	res.method = FrameResponse
	res.streamID = req.StreamID()
	// 回应对方
	self.postFrame(res)
}

func (self *RelayStream) onResponse(res *Frame) {
	self.waitResponse.CompareDelete(res.StreamID(), func(waiter chan<- *Frame) (del bool) {
		del = true
		// 防止waiter错误
		defer func() {
			recover()
		}()

		waiter <- res
		return
	})
}

func (self *RelayStream) postFrame(frame *Frame) {
	self.locker.Lock()
	defer self.locker.Unlock()

	buff := frame.Encode()
	for len(buff) > 0 {
		n, err := self.conn.Write(buff)
		if nil != err {
			return
		}

		buff = buff[n:]
	}
}

// 发送请求
func (self *RelayStream) Send(req *Frame, timeout int) (*Frame, error) {
	sid := atomic.AddUint32(&self.streamID, 1)
	req.streamID = sid
	req.method = FrameRequest

	// 创建回调等待
	waiter := make(chan *Frame)
	defer close(waiter)
	// 存储回调
	self.waitResponse.Store(sid, waiter)
	defer self.waitResponse.Delete(sid)

	// 将数据发送出去
	self.postFrame(req)

	// 等待服务端回应
	after := time.NewTimer(time.Second * time.Duration(timeout))

	select {
	case <-after.C:
		return nil, fmt.Errorf("time out")
	case res := <-waiter:
		after.Stop()
		return res, nil
	}
}

func (self *RelayStream) Close() {
	self.conn.Close()
}

func (self *RelayStream) Detach(ctx context.Context) (net.Conn, error) {
	c, ok := self.conn.(utils.IConnWrap)
	if false == ok {
		return nil, fmt.Errorf("error conn wrap")
	}

	return c.Detach(ctx), nil
}
