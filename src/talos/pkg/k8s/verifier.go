package k8s

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"

	// "github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5"
	// "github.com/go-jose/go-jose/v3/jwt"
)

type Verifier struct {
	publicKey *rsa.PublicKey
}

func NewVerifier(_ context.Context) (*Verifier, error) {
	publicKey, err := getPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s public key: %w", err)
	}

	return &Verifier{
		publicKey: publicKey,
	}, nil
}

func (v *Verifier) VerifyWithClient(k8sToken string) (string, jwt.Claims, error) {
	var claims privateClaims
	token, err := jwt.ParseWithClaims(k8sToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})
	if err != nil {
		return "", nil, fmt.Errorf("parsing jwt: %v", err)
	}

	// claims, ok := token.Claims.(privateClaims)
	// if !ok {
	// 	return "", claims, fmt.Errorf("token cannot be converted to known one, which means it is invalid")
	// } else
	if !token.Valid {
		return "", claims, fmt.Errorf("token cannot be converted to known one, which means it is invalid")
	}

	podName := claims.Kubernetes.Pod.Name
	namespace := claims.Kubernetes.Namespace

	// Namespace is validated here to provide readable error instead of default invalid client error.
	if podName == "" || namespace == "" {
		return "", claims, fmt.Errorf("invalid k8s token claims (pod: %s, namespace: %s)", podName, namespace)
	} else if !strings.HasPrefix(podName+"-", namespace) {
		return "", claims, fmt.Errorf("pod name and namespace must both start with service name (pod: %s, namespace: %s)", podName, namespace)
	}

	return claims.Kubernetes.Namespace, claims, nil
}
