package password_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	password_helpers "github.com/yca-software/go-common/password"
)

type PasswordTestSuite struct {
	suite.Suite
}

func TestPasswordTestSuite(t *testing.T) {
	suite.Run(t, new(PasswordTestSuite))
}

func (s *PasswordTestSuite) TestHash() {
	password := "test-password-123"
	hash, err := password_helpers.Hash(password)
	s.Require().NoError(err)

	s.NotEmpty(hash)
	s.Contains(hash, "$argon2id$")
	s.NotEqual(password, hash)
}

func (s *PasswordTestSuite) TestHash_ProducesDifferentHashes() {
	password := "same-password"
	hash1, err := password_helpers.Hash(password)
	s.Require().NoError(err)
	hash2, err := password_helpers.Hash(password)
	s.Require().NoError(err)

	// Each hash should be different due to random salt
	s.NotEqual(hash1, hash2)
}

func (s *PasswordTestSuite) TestHash_EmptyPassword() {
	hash, err := password_helpers.Hash("")
	s.Require().NoError(err)

	// Empty password should still produce a hash
	s.NotEmpty(hash)
	s.Contains(hash, "$argon2id$")
}

func (s *PasswordTestSuite) TestCompare_CorrectPassword() {
	password := "my-secure-password"
	hash, err := password_helpers.Hash(password)
	s.Require().NoError(err)

	result := password_helpers.Compare(password, hash)
	s.True(result, "Correct password should match hash")
}

func (s *PasswordTestSuite) TestCompare_IncorrectPassword() {
	password := "correct-password"
	wrongPassword := "wrong-password"
	hash, err := password_helpers.Hash(password)
	s.Require().NoError(err)

	result := password_helpers.Compare(wrongPassword, hash)
	s.False(result, "Incorrect password should not match hash")
}

func (s *PasswordTestSuite) TestCompare_EmptyPassword() {
	password := "some-password"
	hash, err := password_helpers.Hash(password)
	s.Require().NoError(err)

	result := password_helpers.Compare("", hash)
	s.False(result, "Empty password should not match non-empty hash")
}

func (s *PasswordTestSuite) TestCompare_InvalidHashFormat() {
	password := "test-password"

	// Test various invalid hash formats
	invalidHashes := []string{
		"not-a-hash",
		"$argon2id$",
		"$argon2id$v=19$m=65536,t=3,p=4$",
		"$argon2$v=19$m=65536,t=3,p=4$salt$hash",   // wrong algorithm
		"$argon2id$v=18$m=65536,t=3,p=4$salt$hash", // wrong version
	}

	for _, invalidHash := range invalidHashes {
		result := password_helpers.Compare(password, invalidHash)
		s.False(result, "Invalid hash format should return false")
	}
}

func (s *PasswordTestSuite) TestCompare_EmptyHash() {
	result := password_helpers.Compare("any-password", "")
	s.False(result, "Empty hash should return false")
}

func (s *PasswordTestSuite) TestHashAndCompare_Integration() {
	testCases := []struct {
		name     string
		password string
	}{
		{"simple password", "password123"},
		{"complex password", "P@ssw0rd!@#$%^&*()"},
		{"unicode password", "пароль123"},
		{"long password", "this-is-a-very-long-password-that-exceeds-normal-length-requirements"},
		{"short password", "123"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			hash, err := password_helpers.Hash(tc.password)
			s.Require().NoError(err)
			s.NotEmpty(hash)

			// Correct password should match
			s.True(password_helpers.Compare(tc.password, hash))

			// Wrong password should not match
			s.False(password_helpers.Compare(tc.password+"wrong", hash))
		})
	}
}
