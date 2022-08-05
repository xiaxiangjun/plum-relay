package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/BurntSushi/toml"
	"math/big"
	"os"
	"path/filepath"
	"plum-relay/utils"
	"time"
)

type ClientConfigRoot struct {
	Client ServerConfig `toml:"client"`
}

// 生成默认的配置文件
func CreateClientDefault(clientCfg, serverCfg string) error {
	// 判断文件是否存在
	if utils.FileIsExist(clientCfg) {
		return fmt.Errorf("config '%s' is exist", filepath.Base(clientCfg))
	}

	// 加载服务端配置文件
	svrConfig, err := LoadServerConfig(serverCfg)
	if nil != err {
		return err
	}

	svrCert, svrKey, err := LoadCert([]byte(svrConfig.Server.Cert), []byte(svrConfig.Server.Key))
	if nil != err {
		return err
	}

	// 生成密钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if nil != err {
		return err
	}

	// 证书模板
	certTpl := x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			CommonName: "plum-relay",
			Country:    []string{"CN"}, // 国家
		},
		NotBefore:             time.Now(),                                                   // 开始时间
		NotAfter:              time.Now().AddDate(50, 0, 0),                                 // 过期时间
		BasicConstraintsValid: true,                                                         // 基本的有效性约束
		IsCA:                  false,                                                        // 是否根证书
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, // 数字签名, 密钥加密
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:              []string{svrConfig.Server.IP},
	}

	// 生成证书
	cert, err := x509.CreateCertificate(rand.Reader, &certTpl, svrCert, &key.PublicKey, svrKey)
	if nil != err {
		return err
	}

	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	cfg := &ClientConfigRoot{
		Client: ServerConfig{
			IP:   svrConfig.Server.IP,
			Port: 8603,
			Cert: string(certPem),
			Key:  string(keyPem),
		},
	}

	fp, err := os.OpenFile(clientCfg, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if nil != err {
		return err
	}

	defer fp.Close()
	// 生成toml文件
	return toml.NewEncoder(fp).Encode(cfg)
}
