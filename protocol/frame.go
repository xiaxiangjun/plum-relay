package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const FrameRequest byte = 1
const FrameResponse byte = 2

/*
0       8       16      24      32
+-------+-------+-------+-------+
| v-m-f |  head-length  |  ->
+-------+-------+-------+-------+
 -> all-length          |  ->
+-------+-------+-------+-------+
 -> stream-id           |
+-------+-------+-------+
v-m-f: version(2bit), method(2bit), flag(4bit)
head-length: head length(8bit)
all-length: all length(32bit), 包含头部的长度

> 头部扩展字段，用于存放额外的信息
> 扩展字段采用 key-value格式组合
key(8bit) + value(?bit)
> key: 为有限的数量，具体含意由通讯双方自行约定
> value: 只支持有限的数据类型，由 data-type(4bit) + data-length(12bit) + data(?bit)格式组成
> data-type定义：
> > 1: string, 2: int32, 3: int64,
*/

type Frame struct {
	streamID uint32
	method   byte
	flag     byte
	head     map[byte]*HeadValue
	body     []byte
}

func NewFrame(method byte, streamID uint32) *Frame {
	return &Frame{
		streamID: streamID,
		method:   method,
		head:     make(map[byte]*HeadValue),
	}
}

func ReadFrame(reader io.Reader) (*Frame, error) {
	// read head
	var preHead [11]byte
	_, err := io.ReadFull(reader, preHead[:])
	if nil != err {
		return nil, err
	}

	// 判断协议是否支持
	if 0x50 != preHead[0] && 0x60 != preHead[0] {
		return nil, fmt.Errorf("version or method error")
	}

	// 读取数据区
	frameLen := binary.BigEndian.Uint32(preHead[3:7])
	buff := make([]byte, frameLen)
	copy(buff[:11], preHead[:])
	_, err = io.ReadFull(reader, buff[11:])
	if nil != err {
		return nil, err
	}

	// 解析包
	headLen := binary.BigEndian.Uint16(buff[1:3])
	frame := &Frame{
		streamID: binary.BigEndian.Uint32(buff[7:11]),
		method:   (buff[0] & 0x30) >> 4,
		flag:     buff[0] & 0xf,
		body:     buff[headLen:],
	}

	frame.head, err = DecodeHeadValue(buff[11:headLen])
	if nil != err {
		return nil, err
	}

	return frame, nil
}

func (self *Frame) Method() byte {
	return self.method
}

func (self *Frame) StreamID() uint32 {
	return self.streamID
}

func (self *Frame) GetHead(key byte) (*HeadValue, bool) {
	v, o := self.head[key]
	return v, o
}

func (self *Frame) SetHead(key byte, value *HeadValue) {
	self.head[key] = value
}

func (self *Frame) GetBody() []byte {
	return self.body
}

func (self *Frame) SetBody(b []byte) {
	self.body = b
}

func (self *Frame) RangeHead(cb func(byte, *HeadValue)) {
	for k, v := range self.head {
		cb(k, v)
	}
}

func (self *Frame) GetFlag() byte {
	return self.flag
}

func (self *Frame) SetFlag(f byte) {
	self.flag = f & 0xf
}

func (self *Frame) Encode() []byte {
	writer := bytes.NewBuffer(nil)

	// 组合head
	data := bytes.NewBuffer(nil)
	for k, v := range self.head {
		data.Write(v.Encode(k))
	}

	// write header
	writer.WriteByte((byte(0x1) << 6) | ((self.method & 0x3) << 4) | (self.flag & 0xf))
	// head len
	headLen := 11 + data.Len()
	writer.WriteByte(byte((headLen >> 8) & 0xff))
	writer.WriteByte(byte(headLen & 0xff))
	// data len
	dataLen := headLen + len(self.body)
	writer.WriteByte(byte((dataLen >> 24) & 0xff))
	writer.WriteByte(byte((dataLen >> 16) & 0xff))
	writer.WriteByte(byte((dataLen >> 8) & 0xff))
	writer.WriteByte(byte(dataLen & 0xff))
	// stream id
	writer.WriteByte(byte((self.streamID >> 24) & 0xff))
	writer.WriteByte(byte((self.streamID >> 16) & 0xff))
	writer.WriteByte(byte((self.streamID >> 8) & 0xff))
	writer.WriteByte(byte(self.streamID & 0xff))
	// header
	if len(data.Bytes()) > 0 {
		writer.Write(data.Bytes())
	}

	// body
	if len(self.body) > 0 {
		writer.Write(self.body)
	}

	return writer.Bytes()
}
