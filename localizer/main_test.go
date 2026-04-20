package localizer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	error_helpers "github.com/yca-software/go-common/error"
	localizer "github.com/yca-software/go-common/localizer"
)

type TranslateTestSuite struct {
	suite.Suite
	localesDir string
	translator localizer.Translator
}

func TestTranslateTestSuite(t *testing.T) {
	suite.Run(t, new(TranslateTestSuite))
}

func (s *TranslateTestSuite) SetupSuite() {
	// Get the absolute path to testdata/locales directory
	// This ensures the path works regardless of where tests are run from
	wd, err := os.Getwd()
	require.NoError(s.T(), err)
	s.localesDir = filepath.Join(wd, "testdata", "locales")

	// Create translator with test locales
	s.translator = localizer.NewTranslator([]string{"en", "es", "tr"}, "en", s.localesDir)
	require.NotNil(s.T(), s.translator)
}

func (s *TranslateTestSuite) TestNewTranslator() {
	translator := localizer.NewTranslator([]string{"en", "es", "tr"}, "en", s.localesDir)
	assert.NotNil(s.T(), translator)
}

func (s *TranslateTestSuite) TestTranslate_SupportedLanguage() {
	result := s.translator.Translate("en", "welcome.message", map[string]any{"Name": "John"})
	assert.Equal(s.T(), "Hello John", result)
}

func (s *TranslateTestSuite) TestTranslate_UnsupportedLanguage() {
	// Request unsupported language, should fallback to default
	result := s.translator.Translate("fr", "welcome.message", map[string]any{"Name": "John"})
	assert.Equal(s.T(), "Hello John", result)
}

func (s *TranslateTestSuite) TestTranslate_MissingKey() {
	// Missing key should return the key itself
	result := s.translator.Translate("en", "missing.key", nil)
	assert.Equal(s.T(), "missing.key", result)
}

func (s *TranslateTestSuite) TestTranslate_CaseInsensitive() {
	// Uppercase language should be lowercased and work
	result := s.translator.Translate("EN", "welcome.message", map[string]any{"Name": "John"})
	assert.Equal(s.T(), "Hello John", result)
}

func (s *TranslateTestSuite) TestTranslateError() {
	apiErr := error_helpers.NewBadRequestError(nil, "0001", nil)
	s.translator.TranslateError("en", apiErr, nil)

	assert.Equal(s.T(), "Invalid request", apiErr.Message)
}

func (s *TranslateTestSuite) TestTranslateError_WithExtraData() {
	extra := map[string]any{"Resource": "User"}
	apiErr := error_helpers.NewNotFoundError(nil, "0001", extra)
	s.translator.TranslateError("en", apiErr, nil)

	assert.Equal(s.T(), "User not found", apiErr.Message)
}

func (s *TranslateTestSuite) TestTranslateError_ExtraAsMapStringString() {
	extra := map[string]string{"Field": "email"}
	apiErr := error_helpers.NewUnprocessableEntityError(nil, "0001", extra)
	s.translator.TranslateError("en", apiErr, nil)

	assert.Equal(s.T(), "Validation failed: email", apiErr.Message)
}

func (s *TranslateTestSuite) TestTranslateError_MissingTranslation() {
	apiErr := error_helpers.NewBadRequestError(nil, "0002", nil)
	s.translator.TranslateError("en", apiErr, nil)

	// Missing translation should result in the key being used
	assert.Contains(s.T(), apiErr.Message, "error.400.0002")
}

func (s *TranslateTestSuite) TestTranslateError_NilError() {
	// Should not panic when err is nil
	s.translator.TranslateError("en", nil, nil)
}

func (s *TranslateTestSuite) TestNewTranslator_NormalizesLanguageTags() {
	// Uppercase in supported/default should match lowercase requests (list is normalized in constructor)
	translator := localizer.NewTranslator([]string{"EN", "ES"}, "EN", s.localesDir)
	require.NotNil(s.T(), translator)

	result := translator.Translate("en", "welcome.message", map[string]any{"Name": "John"})
	assert.Equal(s.T(), "Hello John", result)
}

func (s *TranslateTestSuite) TestTranslate_EmptyLangFallsBackToDefault() {
	result := s.translator.Translate("", "welcome.message", map[string]any{"Name": "John"})
	assert.Equal(s.T(), "Hello John", result)
}

func (s *TranslateTestSuite) TestTranslate_FallbackToDefaultWhenMissingInPreferredLanguage() {
	// welcome.message exists in en.json only; es.json uses a different shape and does not define it.
	result := s.translator.Translate("es", "welcome.message", map[string]any{"Name": "John"})
	assert.Equal(s.T(), "Hello John", result)
}
