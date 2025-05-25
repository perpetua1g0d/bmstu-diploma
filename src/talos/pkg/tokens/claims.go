package tokens

import "time"

type Claims struct {
	Exp      time.Time `json:"exp"`
	Iat      time.Time `json:"iat"`
	Iss      string    `json:"iss"`
	Sub      string    `json:"sub"`
	Aud      string    `json:"aud"`
	Scope    string    `json:"scope"`
	Roles    []string  `json:"roles"`
	ClientID string    `json:"clientID"`
}
