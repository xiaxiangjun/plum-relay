package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func LoadCert(caCrt, caKey []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	caBlock, _ := pem.Decode(caCrt)
	cert, err := x509.ParseCertificate(caBlock.Bytes)
	if nil != err {
		return nil, nil, err
	}

	keyBlock, _ := pem.Decode(caKey)
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if nil != err {
		return nil, nil, err
	}

	return cert, key, nil
}
