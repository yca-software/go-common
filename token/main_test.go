package token_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	token_helpers "github.com/yca-software/go-common/token"
)

type TokenTestSuite struct {
	suite.Suite
}

func TestTokenTestSuite(t *testing.T) {
	suite.Run(t, new(TokenTestSuite))
}

func (s *TokenTestSuite) TestGenerateToken() {
	token, err := token_helpers.GenerateToken()
	s.NoError(err)
	s.NotEmpty(token)
	s.GreaterOrEqual(len(token), 32) // Base64 encoded 32 bytes
}

func (s *TokenTestSuite) TestGenerateToken_ProducesDifferentTokens() {
	token1, err1 := token_helpers.GenerateToken()
	s.NoError(err1)

	token2, err2 := token_helpers.GenerateToken()
	s.NoError(err2)

	s.NotEqual(token1, token2, "Each token should be unique")
}

func (s *TokenTestSuite) TestHashToken() {
	token := "test-token-123"
	hash := token_helpers.HashToken(token)

	s.NotEmpty(hash)
	s.Equal(64, len(hash)) // SHA256 produces 64 hex characters
	s.NotEqual(token, hash)
}

func (s *TokenTestSuite) TestHashToken_ConsistentHashing() {
	token := "same-token"
	hash1 := token_helpers.HashToken(token)
	hash2 := token_helpers.HashToken(token)

	s.Equal(hash1, hash2, "Same token should produce same hash")
}

func (s *TokenTestSuite) TestHashToken_DifferentTokens() {
	hash1 := token_helpers.HashToken("token1")
	hash2 := token_helpers.HashToken("token2")

	s.NotEqual(hash1, hash2, "Different tokens should produce different hashes")
}

func (s *TokenTestSuite) TestHashToken_EmptyString() {
	hash := token_helpers.HashToken("")
	s.NotEmpty(hash)
	s.Len(hash, 64) // SHA256 hex length
}
