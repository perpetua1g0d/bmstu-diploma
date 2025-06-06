func (ts *TokenSource) runScheduler(ctx context.Context) {
	planner := time.NewTimer(0)
	defer planner.Stop()

	for {
		select {
		case <-ts.closeCh:
			return
		case <-ts.refreshCh:
		case <-planner.C:
		}

		delay := func() (delay time.Duration) {
			tokenResp, err := ts.issuer.IssueToken(ctx, ts.scope)
			if err != nil {
				log.Printf("failed to issue token to %s scope: %v", ts.scope, err)
				return ts.cfg.ErrTokenBackoff
			}

			accessToken := tokenResp.AccessToken
			ts.token.Store(&accessToken)

			expiry := tokenResp.ExpiresIn
			newDelay := calcDelay(time.Until(expiry))
			log.Printf("New token to %s scope has been issued, expiry: %s, until_next: %s", ts.scope, expiry, newDelay)
			return newDelay
		}()

		resetTimer(planner, delay)
	}
}