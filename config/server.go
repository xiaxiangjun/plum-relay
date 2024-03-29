package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"plum-relay/utils"
	"time"
)

// TOML文件格式如下
// [server]
// port=8603
// cert=xxx
// key=xxx
//

type ServerConfigRoot struct {
	Server ServerConfig `toml:"server"`
}

type ServerConfig struct {
	IP   string `toml:"ip"`
	Port int    `toml:"port"`
	Cert string `toml:"cert"`
	Key  string `toml:"key"`
}

// 生成默认的配置文件
func CreateServerDefault(path string, server string) error {
	// 判断文件是否存在
	if utils.FileIsExist(path) {
		return fmt.Errorf("config '%s' is exist", filepath.Base(path))
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
		IsCA:                  true,                                                         // 是否根证书
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, // 数字签名, 密钥加密
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:              []string{server},
	}

	// 生成证书
	cert, err := x509.CreateCertificate(rand.Reader, &certTpl, &certTpl, &key.PublicKey, key)
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

	cfg := &ServerConfigRoot{
		Server: ServerConfig{
			IP:   server,
			Port: 8603,
			Cert: string(certPem),
			Key:  string(keyPem),
		},
	}

	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if nil != err {
		return err
	}

	defer fp.Close()
	// 生成toml文件
	return toml.NewEncoder(fp).Encode(cfg)
}

// 加载server配置文件
func LoadServerConfig(path string) (*ServerConfigRoot, error) {
	// 读取配置文件
	txt, err := ioutil.ReadFile(path)
	if nil != err {
		return nil, err
	}

	cfg := &ServerConfigRoot{}
	err = toml.Unmarshal(txt, cfg)
	if nil != err {
		return nil, err
	}

	return cfg, nil
}
