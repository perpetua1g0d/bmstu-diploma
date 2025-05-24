package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
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

func (k *KeyPair) JWKSPublicKey() map[string]interface{} {
	return map[string]interface{}{
		"kty": "RSA",
		"alg": "RS256",
		"n":   base64.RawURLEncoding.EncodeToString(k.PublicKey.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(k.PublicKey.E)).Bytes()),
		"kid": "talos-key-1",
	}
}

func GenerateJWT(keys *KeyPair, claims map[string]interface{}) string {
	// Реализация подписи JWT
	return "signed-jwt-token"
}
