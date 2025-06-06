func verifyToken(rawToken string, certs *jose.JSONWebKeySet) (*tokenClaims, error) {
	token, err := jwt.ParseSigned(rawToken)
	if err != nil {
		log.Printf("failed to parse token: %v", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	var claims tokenClaims
	for _, header := range token.Headers {
		keys := certs.Key(header.KeyID)
		if len(keys) == 0 {
			continue
		}

		for _, key := range keys {
			if err := token.Claims(key.Public(), &claims); err == nil {
				return &claims, nil
			}
		}
	}

	log.Printf("no certificate found to parse token. certs: %v, tokenHeaders: %v", certs, token.Headers)
	return nil, fmt.Errorf("no certificate found to parse token")
}