package main

import (
	"flag"
	"fmt"
	"log"
	"plum-relay/config"
	"plum-relay/utils"
)

func main() {
	var initConfig string
	var ipConfig string

	flag.StringVar(&initConfig, "init", "", "write server/client config")
	flag.StringVar(&ipConfig, "ip", "", "server ip")

	flag.Parse()

	// 初始化配置文件
	if "" != initConfig {
		switch initConfig {
		case "server": // 创建配置文件
			// 判断是否配置IP信息
			if "" == ipConfig {
				log.Println("error: ip is not set")
				return
			}

			err := config.CreateServerDefault(utils.GetExePath("server.conf"), initConfig)
			if nil != err {
				log.Println("write config fail:", err)
			} else {
				log.Println("write config success")
			}
		case "client": // 创建配置文件
			err := config.CreateClientDefault(utils.GetExePath("client.conf"), utils.GetExePath("server.conf"))
			if nil != err {
				log.Println("write config fail:", err)
			} else {
				log.Println("write config success")
			}
		}

		return
	}

	fmt.Println("exe: ", utils.GetExePath("test.conf"))
}
