package server

import (
	"fmt"
	"strconv"
)

type Packer map[string]interface{}

func (self *Packer) GetString(key string) string {
	val, ok := (*self)[key]
	if false == ok {
		return ""
	}

	str, _ := val.(string)
	return str
}

func (self *Packer) GetInt(key string) int {
	val, ok := (*self)[key]
	if false == ok {
		return 0
	}

	n, _ := strconv.Atoi(fmt.Sprint(val))
	return n
}
