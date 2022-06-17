package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	password    string
	locker      sync.Mutex
	allRegister map[string]*Stream
	allContact  map[string]*Stream
}

func NewServer(passwd string) *Server {
	svr := &Server{
		password: passwd,
	}

	svr.init()

	return svr
}

func (self *Server) init() {
	self.allRegister = make(map[string]*Stream)
	self.allContact = make(map[string]*Stream)
}

func (self *Server) ListenAndServe(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if nil != err {
		log.Panic(err)
		return
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if nil != err {
			log.Println(err)
			break
		}

		// 创建一个stream对像
		stream := NewStream(conn, func(s *Stream) {
			self.locker.Lock()
			defer self.locker.Unlock()

			delete(self.allRegister, s.GetID())
			delete(self.allContact, s.GetID())
		})

		go self.streamServe(stream)
	}
}

func (self *Server) streamServe(stream *Stream) {
	// 读取一帧数据，判断是什么操作
	var pkt Packer
	err := stream.ReadPack(&pkt)
	if nil != err {
		log.Println(stream.RemoteAddr(), "read pack error: ", err)
		return
	}

	cmd := pkt.GetString("cmd")
	if len(cmd) < 4 {
		log.Println(stream.RemoteAddr(), "not support proctocol")
		stream.Close()
	}

	// 开始检验授权信息
	err = self.checkAuth(pkt)
	if nil != err {
		stream.WritePack(map[string]interface{}{
			"cmd":  "res:" + cmd[4:],
			"code": -5,
			"err":  "access denied",
		})
		return
	}

	if "req:register" == cmd {
		self.streamRegister(stream, pkt)
	} else if "req:contact" == cmd {
		self.streamContact(stream, pkt)
	} else if "req:connect" == cmd {
		self.streamConnect(stream, pkt)
	} else {
		log.Println(stream.RemoteAddr(), "not support proctocol")
		stream.Close()
	}
}

// 检测授权
func (self *Server) checkAuth(pkt Packer) error {
	cmd := pkt.GetString("cmd")
	sid := pkt.GetString("sid")
	timeout := pkt.GetInt("timeout")
	sign := pkt.GetString("sign")

	txt := fmt.Sprintf("%v:%v:%v", cmd, sid, timeout)
	hash := hmac.New(sha256.New, []byte(self.password))
	hash.Write([]byte(txt))
	sum := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	if "" == sign || sum != sign {
		return fmt.Errorf("sign is not same")
	}

	return nil
}

// 注册服务
func (self *Server) streamRegister(stream *Stream, pkt Packer) {
	sid := pkt.GetString("sid")

	// 保存stream
	self.locker.Lock()
	self.allRegister[sid] = stream
	stream.SetID(sid)
	self.locker.Unlock()

	// 回复消息
	stream.WritePack(map[string]interface{}{
		"cmd":  "res:register",
		"code": 0,
	})

	for {
		var pkt Packer
		err := stream.ReadPack(&pkt)
		if nil != err {
			break
		}

		cmd := pkt.GetString("cmd")
		if len(cmd) < 4 {
			break
		}

		switch cmd {
		case "req:hert":
			stream.WritePack(map[string]interface{}{
				"cmd":  "res:hert",
				"code": 0,
			})
		default:
			stream.WritePack(map[string]interface{}{
				"cmd":  "res:" + cmd[4:],
				"code": -2,
				"err":  "not support",
			})
		}
	}
}

// 建立联系
func (self *Server) streamContact(stream *Stream, pkt Packer) {
	sid := pkt.GetString("sid")

	// 保存stream
	self.locker.Lock()
	peer, ok := self.allContact[sid]
	if false == ok {
		self.allContact[sid] = stream
	} else {
		delete(self.allContact, sid)
	}

	self.locker.Unlock()
	// 判断是否保存操作
	if false == ok {
		return
	}

	// 交换两边的数据
	go io.Copy(peer, stream)
	go io.Copy(stream, peer)
}

// 建立连接
func (self *Server) streamConnect(stream *Stream, pkt Packer) {

}
