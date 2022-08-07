package relay_server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"plum-relay/config"
	"plum-relay/protocol"
	"plum-relay/relay-server/apiserve"
	"plum-relay/utils"
)

type RelayServer struct {
	iniConfig *config.ServerConfigRoot
	apiRouter *utils.ApiRouter
	store     *config.MemoryStore
}

// 初始化配置文件
func (self *RelayServer) Init(cfg *config.ServerConfigRoot) {
	self.iniConfig = cfg
	self.initApiRouter()
	self.store.Init()
}

func (self *RelayServer) initApiRouter() {
	self.apiRouter = &utils.ApiRouter{}

	self.addRouter("register", apiserve.RegisterHandle)
}

func (self *RelayServer) addRouter(uri string, f interface{}) {
	err := self.apiRouter.AddFunc(uri, f)
	if nil != err {
		panic(err)
	}
}

// 启动监听
func (self *RelayServer) ListenAndServe() error {
	if nil == self.iniConfig {
		return fmt.Errorf("ini config is nil")
	}

	certPem := []byte(self.iniConfig.Server.Cert)
	keyPem := []byte([]byte(self.iniConfig.Server.Key))
	// 加载配置文件
	caCert, _, err := config.LoadCert(certPem, keyPem)
	if nil != err {
		return err
	}

	cert, err := tls.X509KeyPair(certPem, keyPem)
	if nil != err {
		return err
	}

	// 构建ca池
	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	// 加载配置文件
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		Certificates: []tls.Certificate{cert},
	}

	// 监听端口
	l, err := tls.Listen("tcp4", fmt.Sprintf("0.0.0.0:%d", self.iniConfig.Server.Port), tlsConfig)
	if nil != err {
		return err
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if nil != err {
			return err
		}

		go self.serveConn(conn)
	}
}

// 处理一个连接服务
func (self *RelayServer) serveConn(conn net.Conn) {
	stream := protocol.NewRelayStream(utils.NewConnWrap(conn, context.Background()), self)
	defer stream.Close()

	stream.Serve()
	// 调用关闭事件
	self.OnClose(stream)
}

// 处理关闭事件
func (self *RelayServer) OnClose(stream *protocol.RelayStream) {

}

// 接收到消息处理过程
func (self *RelayServer) OnRequrest(stream *protocol.RelayStream, req *protocol.Frame) (res *protocol.Frame, err error) {
	// 读取要处理的任务
	hv, ok := req.GetHead(protocol.HeaderTask)
	if false != ok {
		return nil, fmt.Errorf("not found")
	}

	// 调用处理过程
	buf, err := self.apiRouter.Call(&config.Parameter{
		Store:  self.store,
		Stream: stream,
	}, hv.String(), req.GetBody())
	if nil != err {
		return nil, err
	}

	// 回应客户端
	res = protocol.NewFrame(0, 0)
	res.SetBody(buf)
	return
}
