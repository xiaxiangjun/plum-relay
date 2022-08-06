package protocol

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
)

type HeadValueType interface {
	int32 | int64 | string
}

type HeadValue struct {
	value interface{}
}

// 创建一个headValue
func NewHeadValue[T HeadValueType](s T) *HeadValue {
	return &HeadValue{value: s}
}

/*
> key: 为有限的数量，具体含意由通讯双方自行约定
> value: 只支持有限的数据类型，由 data-type(4bit) + data-length(12bit) + data(?bit)格式组成
> data-type定义：
> > 1: string, 2: int32, 3: int64,
*/
// 解值格式
func DecodeHeadValue(data []byte) (map[byte]*HeadValue, error) {
	ret := make(map[byte]*HeadValue)
	for len(data) > 3 {
		key := data[0]
		vType := binary.BigEndian.Uint16(data[1:3])
		vLen := vType & 0xfff
		vType = vType >> 12
		if int(vLen)+3 > len(data) {
			return nil, fmt.Errorf("can't decode full")
		}

		switch vType {
		case 1:
			ret[key] = NewHeadValue(string(data[3 : 3+int(vLen)]))
		case 2:
			if 4 != vLen {
				return nil, fmt.Errorf("error data type (2) length %d", vLen)
			}

			ret[key] = NewHeadValue(int32(binary.BigEndian.Uint32(data[3 : 3+int(vLen)])))
		case 3:
			if 8 != vLen {
				return nil, fmt.Errorf("error data type (3) length %d", vLen)
			}

			ret[key] = NewHeadValue(int64(binary.BigEndian.Uint64(data[3 : 3+int(vLen)])))
		default:
			return nil, fmt.Errorf("error data type: %d", vType)
		}

		// 指针下移
		data = data[3+int(vLen):]
	}

	return ret, nil
}

func (self *HeadValue) String() string {
	switch reflect.TypeOf(self.value).Kind() {
	case reflect.String:
		return reflect.ValueOf(self.value).String()
	case reflect.Int64:
		fallthrough
	case reflect.Int32:
		return fmt.Sprint(reflect.ValueOf(self.value).Int())
	default:
		return ""
	}
}

func (self *HeadValue) Int() int64 {
	switch reflect.TypeOf(self.value).Kind() {
	case reflect.String:
		i, _ := strconv.ParseInt(reflect.ValueOf(self.value).String(), 10, 64)
		return i
	case reflect.Int64:
		fallthrough
	case reflect.Int32:
		return reflect.ValueOf(self.value).Int()
	default:
		return 0
	}
}

/*
> key: 为有限的数量，具体含意由通讯双方自行约定
> value: 只支持有限的数据类型，由 data-type(4bit) + data-length(12bit) + data(?bit)格式组成
> data-type定义：
> > 1: string, 2: int32, 3: int64,
*/
func (self *HeadValue) Encode(key byte) []byte {
	var vType uint16
	var vData []byte
	switch reflect.TypeOf(self.value).Kind() {
	case reflect.String:
		vType = 1
		vData = []byte(reflect.ValueOf(self.value).String())
	case reflect.Int64:
		vType = 3
		vData = make([]byte, 8)
		binary.BigEndian.PutUint64(vData, uint64(reflect.ValueOf(self.value).Int()))
	case reflect.Int32:
		vType = 2
		vData = make([]byte, 4)
		binary.BigEndian.PutUint32(vData, uint32(reflect.ValueOf(self.value).Int()))
	default:
		return nil
	}

	if len(vData) >= 0x1000 {
		return nil
	}

	vType = (vType << 12) | uint16(len(vData)&0xfff)
	// 构建返回值
	ret := make([]byte, 1+2+len(vData))
	ret[0] = key
	binary.BigEndian.PutUint16(ret[1:3], vType)
	copy(ret[3:], vData)
	return ret
}
