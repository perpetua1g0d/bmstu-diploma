default allow = false
allow { input.jwt.claims.access[_] == "postgres:orders:SELECT" }
