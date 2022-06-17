package main

import (
	"AnyLANServer/server"
	"flag"
)

func main() {
	listenPort := 0
	passwd := ""

	flag.IntVar(&listenPort, "port", 0, "set listen port")
	flag.StringVar(&passwd, "passwd", "", "set password")

	flag.Parse()
	// 判断配置参数是否正确
	if 0 == listenPort || "" == passwd {
		flag.Usage()
		return
	}

	svr := server.NewServer(passwd)

	svr.ListenAndServe(listenPort)
}
