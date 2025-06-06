func Test_Verify(t *testing.T) {
	// Arrange
	token := "ey..."
	jwk := JWK{
		N: `xItwc...`,
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
