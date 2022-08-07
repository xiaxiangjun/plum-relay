package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type ApiRouter struct {
	router map[string]*apiCall
}

type apiCall struct {
	fn  reflect.Value
	ctx reflect.Type
	req reflect.Type
	res reflect.Type
}

func (self *ApiRouter) Init() {
	self.router = make(map[string]*apiCall)
}

func (self *ApiRouter) CheckUri(uri string) bool {
	api, ok := self.router[uri]
	if ok == false || api == nil {
		return false
	}

	return true
}

// 查看请求参数
func (self *ApiRouter) LookReqeust(uri string) (string, error) {
	api, ok := self.router[uri]
	if false == ok {
		return "", errors.New("Not Found")
	}

	req := reflect.New(api.req)
	self.fillRequest(req.Interface())

	buf, _ := json.Marshal(req.Interface())

	var out bytes.Buffer
	json.Indent(&out, buf, "", "\t")

	return out.String(), nil
}

// 填充结构体
func (self *ApiRouter) fillRequest(ins interface{}) {
	val := reflect.ValueOf(ins)
	if val.Kind() != reflect.Ptr {
		return
	}

	if val.Elem().Kind() != reflect.Struct {
		return
	}

	for i := 0; i < val.Elem().NumField(); i++ {
		switch val.Elem().Field(i).Kind() {
		case reflect.Slice:
			// 数组
			// 创建新对像
			tp := reflect.New(val.Elem().Field(i).Type().Elem())
			self.fillRequest(tp.Interface())
			// 加入到数组
			arr := reflect.Append(val.Elem().Field(i), tp.Elem())
			val.Elem().Field(i).Set(arr)
		case reflect.Ptr:
			tp := reflect.New(val.Elem().Field(i).Type().Elem())
			self.fillRequest(tp.Interface())
			val.Elem().Field(i).Set(tp)
		}
	}
}

// 添加路由
// func (*ctx, req *struct, res *struct) error
func (self *ApiRouter) AddFunc(uri string, fn interface{}) error {
	api := &apiCall{}
	api.fn = reflect.ValueOf(fn)
	// 判断是否为函数
	if api.fn.Kind() != reflect.Func {
		return fmt.Errorf("%s: handler is not func", uri)
	}

	fnType := api.fn.Type()
	// 判断参数个数
	if fnType.NumIn() < 2 || fnType.NumIn() > 3 {
		return fmt.Errorf("%s:  parameter number error: %d", uri, fnType.NumIn())
	}

	// 判断context
	pi := 0
	if fnType.NumIn() == 3 {
		// 判断是否为指针
		if reflect.Ptr != fnType.In(pi).Kind() {
			return fmt.Errorf("%s: parameter 1 is not pointer", uri)
		}

		api.ctx = fnType.In(0).Elem()
		if reflect.Struct != api.ctx.Kind() {
			return fmt.Errorf("%s: parameter 1 is not struct pointer", uri)
		}

		pi++
	}

	////////////////////////////////////////////////////////////
	// 读取req参数
	if reflect.Ptr != fnType.In(pi).Kind() {
		return fmt.Errorf("%s: parameter %d is not pointer", uri, pi+1)
	}

	api.req = fnType.In(pi).Elem()
	if reflect.Struct != api.req.Kind() {
		return fmt.Errorf("%s: parameter %d is not struct pointer", uri, pi+1)
	}

	pi++
	////////////////////////////////////////////////////////////
	// 读取res
	if reflect.Ptr != fnType.In(pi).Kind() {
		return fmt.Errorf("%s: parameter %d is not pointer", uri, pi+1)
	}

	api.res = fnType.In(pi).Elem()
	if reflect.Struct != api.res.Kind() {
		return fmt.Errorf("%s: parameter %d is not struct pointer", uri, pi+1)
	}

	// 保存参数
	self.router[uri] = api
	return nil
}

func (self *ApiRouter) CallEx(ctx interface{}, uri string, req []byte,
	before func(uri string, ctx interface{}, req interface{}),
	after func(uri string, ctx interface{}, res interface{}),
) ([]byte, error) {
	// 查找函数
	api, ok := self.router[uri]
	if false == ok {
		return nil, fmt.Errorf("not found")
	}

	// 构建参数
	var args [3]reflect.Value
	var pi int = 0

	// ctx
	if nil != api.ctx {
		// 判断是否为指针
		if reflect.Ptr != reflect.TypeOf(ctx).Kind() {
			return nil, fmt.Errorf("ctx is not pointer")
		}

		if api.ctx.Name() != reflect.TypeOf(ctx).Elem().Name() {
			return nil, fmt.Errorf("ctx is not same: %s <> %s",
				api.ctx.Name(), reflect.TypeOf(ctx).Elem().Name())
		}

		args[0] = reflect.ValueOf(ctx)
		pi++
	}

	// request
	args[pi] = reflect.New(api.req)
	if req != nil {
		err := json.Unmarshal(req, args[pi].Interface())
		if nil != err {
			return nil, err
		}
	}

	if value, check := CheckMax(args[pi].Interface()); check == false {

		var resMap map[string]string
		resMap = map[string]string{}
		resMap["code"] = value + "DataLengthExceedsLimit."
		resMap["message"] = value + " has exceeded the length limit."
		resMap["requst_id"] = ""
		return json.Marshal(resMap)
	}

	// 调用前对request进行处理
	if nil != before {
		before(uri, ctx, args[pi].Interface())
	}

	pi++
	// response
	args[pi] = reflect.New(api.res)

	// 调用函数
	ret := api.fn.Call(args[:pi+1])

	// 判断返回参数
	if len(ret) > 0 {
		e, o := ret[0].Interface().(error)
		if o && nil != e {
			return nil, e
		}
	}

	buf, err := json.Marshal(args[pi].Interface())
	if nil != after {
		after(uri, ctx, args[pi].Interface())
	}
	return buf, err
}

// -2: 找不到接口
// -3: 输入参数格式不对
// -4: 调用接口错误
func (self *ApiRouter) Call(ctx interface{}, uri string, req []byte) ([]byte, error) {
	return self.CallEx(ctx, uri, req, nil, nil)
}

func CheckMax(value interface{}) (string, bool) {
	val := reflect.ValueOf(value)
	if reflect.Ptr == val.Type().Kind() {
		val = val.Elem()
	}

	// fmt.Println("226 >>>", val.Kind())
	if reflect.Struct != val.Kind() {
		return "", true
	}

	for i := 0; i < val.Type().NumField(); i++ {
		tp := val.Field(i).Kind()
		if reflect.Slice == tp {
			for m := 0; m < val.Field(i).Len(); m++ {
				f, o := CheckMax(val.Field(i).Index(m).Interface())
				if false == o {
					return f, o
				}
			}

			continue
		} else if reflect.Struct == tp || reflect.Ptr == tp {
			f, o := CheckMax(val.Field(i).Interface())
			if false == o {
				return f, o
			}

			continue
		} else if reflect.String != tp {
			continue
		}

		// fmt.Println(">>>", val.Type().Field(i).Tag.Get("json"))
		length := val.Type().Field(i).Tag.Get("max")
		if "" == length {
			continue
		}

		size, _ := strconv.Atoi(length)
		if val.Field(i).Len() > size {
			return val.Type().Field(i).Tag.Get("json"), false
		}

	}
	return "", true
}
