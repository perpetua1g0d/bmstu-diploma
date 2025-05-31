package k8s

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testClientID = "postgres-a"

func Test_Verify(t *testing.T) {
	// Arrange
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6Ilp2UFNpYlpISzk1YlhGRjhKSjJlY25MWHZRZ29aV25lMXB3UV9IYUh5TmsifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzc5NjUxMTU2LCJpYXQiOjE3NDgxMTUxNTYsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2YyIsImp0aSI6ImM0ZWUzNWFmLTFmZWYtNDNmYi1iNDYyLTJiNjk0OTM5MTdjMCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoicG9zdGdyZXMtYSIsIm5vZGUiOnsibmFtZSI6ImszZC1ibXN0dWNsdXN0ZXItc2VydmVyLTAiLCJ1aWQiOiI3YzEwM2NlOC05OTA2LTQ3NWMtOGM5Ni1jNGZiZjIyOWM1YTAifSwicG9kIjp7Im5hbWUiOiJwb3N0Z3Jlcy1hLTY3OTRmY2I1ZjctcWI5em0iLCJ1aWQiOiI5MTE3NDllNy01NTFiLTQ2M2UtYmNmZC03ZjEyNGFhZTgxNWUifSwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImRlZmF1bHQiLCJ1aWQiOiIyNDhkNDE3Mi0xZDdlLTRiMWEtYmRlNS05NzAyN2FiNDFlMmQifSwid2FybmFmdGVyIjoxNzQ4MTE4NzYzfSwibmJmIjoxNzQ4MTE1MTU2LCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6cG9zdGdyZXMtYTpkZWZhdWx0In0.spCWGajmAAENHgK5zG5NTX2dz82S0gunARu9-ncvurmV5XqEKOKSEypC9Rm3ap2WfwOei0zm4-0Hbi3fdYZUZkSDZO-HLSHxPZcQoGvVjiKDOAIRDGwNb8bNVqkPPyy1Q8K5cD1anJwPiXYFzIQ4ZKDu_Ikp5ajkA3KYpZUmPLFQ3a09k8ycTpvdnjSzCfIdEaUiO4jFlGJWEYKGo9XuY0VVjyyjwJAdDOj6Ry0aIJzJzQTjS-IUs_dL7XVQUSlxp4mZvhLhnrhiL6uU59tX1QVtliZ3MgkO3XN_F2G5kuoigFN0NQp8EpiCeC2-e9T-rchZ3MR8KbPp8lgs4SA1iA"

	jwk := JWK{
		N: `xItwcttR4qVTD4bBfsUDgpICFnoBk1H8qyN3jSVemH1wlPyn6CLn2aUmjQHW25f2LcraZr1_t7l0ogmaR46Gn7uyYGBEtIsNnjvvoAUVbmd8vIhPJI9flzDjJys4CEjefo1YFooD4YfqDei0GEYG2TYy42mO3TR6O3--47PbLIyZ2cbmHTwU-t_apqc3NUs0Sd6_gjDb0hrX0cFlOvBfL-J-3XEe4Zxew7-qDnjQGIKdSEgS-v-wCwr30iqK9yDfHO9cHUtRNirLb4dybeOh3_vBMMLCVpKtH6GonDEVyRv7qJCigEinpHB78Uq0PAb_l8SOHougk2qp8-Cp3nb7rw`,
		E: "AQAB",
	}

	wantAud := jwt.ClaimStrings{
		"https://kubernetes.default.svc.cluster.local",
		"k3s",
	}

	publicKey, err := makeRSAPublicKey(jwk)
	if err != nil {
		t.Fatalf("failed to create public rsa key: %v", err)
	}

	verifier := &Verifier{
		publicKey: publicKey,
	}

	// Act
	gotClientID, gotClaims, gotErr := verifier.VerifyWithClient(token)

	// Assert
	if gotErr != nil {
		t.Errorf("failed to verify token: %v", gotErr)
	} else if gotClientID != testClientID {
		t.Errorf("expected clientID: %s, got: %s", testClientID, gotClientID)
	}

	gotAud, gotAudErr := gotClaims.GetAudience()
	if gotAudErr != nil {
		t.Errorf("got unexpected aud err: %v", gotAudErr)
	}
	assert.EqualValues(t, wantAud, gotAud)
}
