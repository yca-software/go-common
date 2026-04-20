package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	http_helpers "github.com/yca-software/go-common/http"
)

type LanguageTestSuite struct {
	suite.Suite
	echo        *echo.Echo
	supported   []string
	defaultLang string
}

func TestLanguageTestSuite(t *testing.T) {
	suite.Run(t, new(LanguageTestSuite))
}

func (s *LanguageTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.supported = []string{"en", "es", "tr"}
	s.defaultLang = "en"
}

func (s *LanguageTestSuite) createContext(acceptLanguage string) echo.Context {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if acceptLanguage != "" {
		req.Header.Set("Accept-Language", acceptLanguage)
	}
	rec := httptest.NewRecorder()
	return s.echo.NewContext(req, rec)
}

func (s *LanguageTestSuite) TestGetLanguage_SupportedLanguage() {
	c := s.createContext("en")
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), "en", result)
}

func (s *LanguageTestSuite) TestGetLanguage_UnsupportedLanguage() {
	c := s.createContext("fr")
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), s.defaultLang, result)
}

func (s *LanguageTestSuite) TestGetLanguage_UpperCase() {
	c := s.createContext("EN")
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), "en", result)
}

func (s *LanguageTestSuite) TestGetLanguage_MixedCase() {
	c := s.createContext("Es")
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), "es", result)
}

func (s *LanguageTestSuite) TestGetLanguage_WithRegion() {
	c := s.createContext("en-US,en;q=0.9")
	// Should extract "en" from "en-US"
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), "en", result)
}

func (s *LanguageTestSuite) TestGetLanguage_EmptyHeader() {
	c := s.createContext("")
	// No Accept-Language header
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), s.defaultLang, result)
}

func (s *LanguageTestSuite) TestGetLanguage_ShortLanguageCode() {
	c := s.createContext("t") // Single character
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), s.defaultLang, result)
}

func (s *LanguageTestSuite) TestGetLanguage_SingleCharacter() {
	c := s.createContext("e")
	// Single character should be lowercased and checked
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), s.defaultLang, result)
}

func (s *LanguageTestSuite) TestGetLanguage_MultipleLanguages() {
	c := s.createContext("es,en;q=0.9,tr;q=0.8")
	// Should extract first language "es"
	result := http_helpers.GetLanguage(c, s.supported, s.defaultLang)
	assert.Equal(s.T(), "es", result)
}

func (s *LanguageTestSuite) TestGetLanguage_NilContext_ReturnsDefault() {
	result := http_helpers.GetLanguage(nil, s.supported, s.defaultLang)
	assert.Equal(s.T(), "en", result)
}

func (s *LanguageTestSuite) TestGetLanguage_NormalizesSupportedList() {
	// Uppercase in supported list should still match header "en"
	c := s.createContext("en")
	result := http_helpers.GetLanguage(c, []string{"EN", "ES", "TR"}, "en")
	assert.Equal(s.T(), "en", result)
}

func (s *LanguageTestSuite) TestParseLimitOffset_Defaults() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := s.echo.NewContext(req, httptest.NewRecorder())
	limit, offset := http_helpers.ParseLimitOffset(c, 20, 100)
	assert.Equal(s.T(), 20, limit)
	assert.Equal(s.T(), 0, offset)
}

func (s *LanguageTestSuite) TestParseLimitOffset_FromQuery() {
	req := httptest.NewRequest(http.MethodGet, "/?limit=50&offset=10", nil)
	c := s.echo.NewContext(req, httptest.NewRecorder())
	limit, offset := http_helpers.ParseLimitOffset(c, 20, 100)
	assert.Equal(s.T(), 50, limit)
	assert.Equal(s.T(), 10, offset)
}

func (s *LanguageTestSuite) TestParseLimitOffset_ClampsLimitToMax() {
	req := httptest.NewRequest(http.MethodGet, "/?limit=200", nil)
	c := s.echo.NewContext(req, httptest.NewRecorder())
	limit, _ := http_helpers.ParseLimitOffset(c, 20, 100)
	assert.Equal(s.T(), 20, limit)
}

func (s *LanguageTestSuite) TestParseLimitOffset_NilContext() {
	limit, offset := http_helpers.ParseLimitOffset(nil, 20, 100)
	assert.Equal(s.T(), 20, limit)
	assert.Equal(s.T(), 0, offset)
}
