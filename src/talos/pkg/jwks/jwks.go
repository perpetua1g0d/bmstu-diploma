package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type KeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func GenerateKeyPair() *KeyPair {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}
}

func (k *KeyPair) PublicPEM() string {
	pubASN1 := x509.MarshalPKCS1PublicKey(k.PublicKey)
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})
	return string(pubBytes)
}
