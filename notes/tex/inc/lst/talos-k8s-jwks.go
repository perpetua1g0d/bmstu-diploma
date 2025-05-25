type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func getPublicKey() (*rsa.PublicKey, error) {
	k8sCertPath := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	caCert, err := os.ReadFile(k8sCertPath)
	if err != nil {
		return nil, fmt.Errorf("error reading CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	k8sTokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	token, err := os.ReadFile(k8sTokenPath)
	if err != nil {
		return nil, fmt.Errorf("error reading token: %w", err)
	}

	req, err := http.NewRequest("GET", "https://kubernetes.default.svc/openid/v1/jwks", nil)
	if err != nil {
		return nil, fmt.Errorf("creating k8s jwks request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+string(token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JWKS request failed: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("JWKS parse error: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, errors.New("no keys in JWKS")
	}

	key := jwks.Keys[0]
	return makeRSAPublicKey(key)
}

func makeRSAPublicKey(key JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent: %w", err)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}
