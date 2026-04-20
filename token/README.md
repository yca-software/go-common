# Token

Secure random token generation and SHA256 hashing. Used for refresh tokens, email verification, password reset, and API keys: generate a token, show it once to the user or client, store only the hash in the DB and compare by hashing the presented value. Do not use for passwords (use the password package).

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `TokenLength` (32), `GenerateToken()` (crypto/rand, base64 URL-encoded), `HashToken(token)` (SHA256 hex). |
| **main_test.go** | Tests for uniqueness, length, and hash consistency.                  |

## Functions

- **GenerateToken()** — Returns a new cryptographically random token (32 bytes, base64 URL-encoded). Use for one-time secrets and tokens that are shown once.
- **HashToken(token)** — Returns SHA256 hash as hex string. Store the hash; on verification, hash the input and compare with stored hash (constant-time if needed at call site).

## Usage

```go
import "github.com/yca-software/go-common/token"

raw, _ := token.GenerateToken()
// send raw to user once (e.g. link, API key response)
storedHash := token.HashToken(raw)
// in DB store storedHash only
// on verify: token.HashToken(userInput) == storedHash
```

Never log or expose `raw` after it has been shown to the user. Use constant-time comparison when comparing hashes if the token is user-controlled.
