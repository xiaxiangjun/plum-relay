package protocol

import (
	"bytes"
	"testing"
)

func Test_ReadFrame(t *testing.T) {
	buff := bytes.NewBuffer(nil)

	frame := NewFrame(FrameRequest, 1000)
	buff.Write(frame.Encode())

	frame = NewFrame(FrameResponse, 1000)
	frame.SetHead(1, NewHeadValue("hello"))
	frame.SetBody([]byte("hi"))
	buff.Write(frame.Encode())

	t.Logf("%+v", buff.Bytes())

	reader := bytes.NewReader(buff.Bytes())
	frame, err := ReadFrame(reader)
	if nil != err {
		t.Error(err)
		return
	}

	t.Logf("%+v", frame)

	frame, err = ReadFrame(reader)
	if nil != err {
		t.Error(err)
		return
	}

	t.Logf("%+v", frame)
}
