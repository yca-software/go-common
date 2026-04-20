# Password

Argon2id hashing and constant-time comparison for passwords. Uses fixed parameters (memory, iterations, parallelism, key length) and encodes salt + hash in a single string. Safe for storing user passwords and verifying login. Used by go-api auth service.

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `Hash(password)`, `Compare(password, encodedHash)`, Argon2id params and encoding. |
| **main_test.go** | Tests for hash round-trip and comparison (including invalid/empty).   |

## Functions

- **Hash(password)** — Returns encoded string with algorithm and params (e.g. `$argon2id$v=...$m=...,t=...,p=...$salt$hash`). Errors only on salt generation failure.
- **Compare(password, encodedHash)** — Constant-time comparison; returns false on parse failure or mismatch (timing-safe).

## Usage

```go
import "github.com/yca-software/go-common/password"

encoded, _ := password.Hash("user-secret")
// store encoded in DB
ok := password.Compare("user-secret", encoded)
```

Do not log or expose `encoded`; treat it as a secret. Use only for user passwords, not tokens (use token package for tokens).
