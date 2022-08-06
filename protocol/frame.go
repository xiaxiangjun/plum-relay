package protocol

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
	head map[byte]*HeadValue
}
