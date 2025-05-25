func CertsHandler(keys *jwks.KeyPair) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwks := keys.JWKS()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
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
