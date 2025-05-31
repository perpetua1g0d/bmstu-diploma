package k8s

import (
	"github.com/golang-jwt/jwt/v5"
)

type privateClaims struct {
	Exp        *jwt.NumericDate `json:"exp"`
	Iat        *jwt.NumericDate `json:"iat"`
	Nbf        *jwt.NumericDate `json:"nbf"`
	Iss        string           `json:"iss"`
	Sub        string           `json:"sub"`
	Aud        jwt.ClaimStrings `json:"aud"`
	Kubernetes kubernetesClaims `json:"kubernetes.io"`
}

func (p privateClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return p.Exp, nil
}

func (p privateClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return p.Iat, nil
}

func (p privateClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return p.Nbf, nil
}

func (p privateClaims) GetIssuer() (string, error) {
	return p.Iss, nil
}

func (p privateClaims) GetSubject() (string, error) {
	return p.Sub, nil
}

func (p privateClaims) GetAudience() (jwt.ClaimStrings, error) {
	return p.Aud, nil
}

type ref struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}
type kubernetesClaims struct {
	Namespace string `json:"namespace"`
	Pod       ref    `json:"pod"`
}
