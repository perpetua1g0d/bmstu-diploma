type TokenSource struct {
	cfg *config.Config

	scope  string
	token  atomic.Pointer[string]
	issuer *Issuer

	refreshCh chan struct{}
	closeCh   chan struct{}
}

type TokenSet struct {
	sync.RWMutex

	set map[string]*TokenSource
}

func NewTokenSet(ctx context.Context, cfg *config.Config, scopes []string) (*TokenSet, error) {
	set := &TokenSet{
		set: make(map[string]*TokenSource),
	}
	for _, scope := range scopes {
		ts, err := NewTokenSource(ctx, cfg, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to create tokensource for %s scope: %w", scope, err)
		}

		set.set[scope] = ts
	}

	return set, nil
}

func (t *TokenSet) Token(scope string) (string, error) {
	ts, ok := t.set[scope]
	if !ok {
		return "", fmt.Errorf("no tokensource for provided scope: %s", scope)
	}

	token := ts.Token()
	if token == "" {
		return "", fmt.Errorf("token is empty for provided scope: %s, check logs", scope)
	}

	return token, nil
}

func NewTokenSource(ctx context.Context, cfg *config.Config, scope string) (*TokenSource, error) {
	issuer := NewIssuer(cfg)

	ts := &TokenSource{
		cfg:       cfg,
		scope:     scope,
		issuer:    issuer,
		token:     atomic.Pointer[string]{},
		refreshCh: make(chan struct{}),
		closeCh:   make(chan struct{}),
	}

	go ts.runScheduler(context.WithoutCancel(ctx))

	return ts, nil
}

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

		if !ts.cfg.SignEnabled {
			resetTimer(planner, 1*time.Hour)
			continue
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

func calcDelay(ttl time.Duration) time.Duration {
	return time.Duration(rand.Float32() * float32(ttl))
}

// resetTimer stops, drains and resets the timer.
func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}

	t.Reset(d)
}
