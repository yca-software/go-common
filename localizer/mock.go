package localizer

import (
	"github.com/stretchr/testify/mock"
	error_helpers "github.com/yca-software/go-common/error"
)

type MockTranslator struct {
	mock.Mock
}

func NewMockTranslator() *MockTranslator {
	return &MockTranslator{}
}

func (m *MockTranslator) Translate(lang string, key string, params map[string]any) string {
	args := m.Called(lang, key, params)
	return args.String(0)
}

func (m *MockTranslator) TranslateError(lang string, err *error_helpers.Error, params map[string]any) {
	args := m.Called(lang, err, params)
	if err != nil {
		err.Message = args.String(0)
	}
}
