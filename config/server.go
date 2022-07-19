package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/BurntSushi/toml"
	"math/big"
	"os"
	"time"
)

// TOML文件格式如下
// [server]
// port=8603
// cert=xxx
// key=xxx
//

type Config struct {
	Server ServerConfig `toml:"server"`
}

type ServerConfig struct {
	Port int    `toml:"port"`
	Cert string `toml:"cert"`
	Key  string `toml:"key"`
}

// 生成默认的配置文件
func CreateServerDefault(path string) error {
	// 生成密钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if nil != err {
		return err
	}

	// 证书模板
	certTpl := x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			CommonName: "plum-os",
			Country:    []string{"CN"}, // 国家
		},
		NotBefore:             time.Now(),                                                   // 开始时间
		NotAfter:              time.Now().AddDate(50, 0, 0),                                 // 过期时间
		BasicConstraintsValid: true,                                                         // 基本的有效性约束
		IsCA:                  true,                                                         // 是否根证书
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, // 数字签名, 密钥加密
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
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

	cfg := &Config{
		Server: ServerConfig{
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

func InitServerConfig(path string) error {
	return nil
}
