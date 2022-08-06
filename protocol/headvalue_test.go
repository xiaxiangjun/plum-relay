package protocol

import (
	"testing"
)

func Test_HeadValue(t *testing.T) {
	head := make(map[byte]*HeadValue)

	head[3] = NewHeadValue(int32(10))
	head[5] = NewHeadValue(int64(100))
	head[7] = NewHeadValue("this is test")

	data := make([]byte, 0, 1024)
	for k, v := range head {
		data = append(data, v.Encode(k)...)
	}

	t.Logf("%+v ", data)

	head, err := DecodeHeadValue(data)
	if nil != err {
		t.Error(err)
		return
	}

	t.Logf("%+v ", head)
}
