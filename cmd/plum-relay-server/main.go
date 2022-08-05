package main

import (
	"flag"
	"fmt"
	"log"
	"plum-relay/config"
	"plum-relay/utils"
)

func main() {
	var initConfig bool

	flag.BoolVar(&initConfig, "init", false, "write default config")

	flag.Parse()

	if initConfig {
		// 创建配置文件
		err := config.CreateServerDefault(utils.GetExePath("server.conf"))
		if nil != err {
			log.Println("write config fail:", err)
		} else {
			log.Println("write config success")
		}

		return
	}

	fmt.Println("exe: ", utils.GetExePath("test.conf"))
}

type Number interface {
	~int | ~float64
}

func test[A Number](a, b A) A {
	return a + b
}
