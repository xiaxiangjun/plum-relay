package relay_server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"plum-relay/config"
)

type RelayServer struct {
	iniConfig *config.ServerConfigRoot
}

// 初始化配置文件
func (self *RelayServer) Init(cfg *config.ServerConfigRoot) {
	self.iniConfig = cfg
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

func (self *RelayServer) serveConn(conn net.Conn) {
	defer conn.Close()

}
