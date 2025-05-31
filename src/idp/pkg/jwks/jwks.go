package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/tokens"
)

type KeyPair struct {
	PrivateKey  *rsa.PrivateKey
	Certificate *x509.Certificate
	KeyID       string
}

func GenerateKeyPair() *KeyPair {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "idp-oidc"},
		NotBefore:             now,
		NotAfter:              now.Add(24 * time.Hour * 365),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	certDER, _ := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		privateKey.Public(),
		privateKey,
	)

	cert, _ := x509.ParseCertificate(certDER)

	return &KeyPair{
		PrivateKey:  privateKey,
		Certificate: cert,
		KeyID:       generateKeyID(),
	}
}

func (k *KeyPair) JWKS() jose.JSONWebKeySet {
	jwk := jose.JSONWebKey{
		Key:          k.PrivateKey.Public(),
		Certificates: []*x509.Certificate{k.Certificate},
		KeyID:        k.KeyID,
		Algorithm:    "RS256",
		Use:          "sig",
	}

	return jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
}

func generateKeyID() string {
	const defaultLength = 24

	buf := make([]byte, defaultLength)
	rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

func GenerateJWT(signer jose.Signer, claims tokens.Claims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	signature, err := signer.Sign(payload)
	if err != nil {
		return "", err
	}

	return signature.CompactSerialize()
}

func getX5t(cert *x509.Certificate) string {
	h := sha1.Sum(cert.Raw)
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func getX5tS256(cert *x509.Certificate) string {
	h := sha256.Sum256(cert.Raw)
	return base64.RawURLEncoding.EncodeToString(h[:])
}
