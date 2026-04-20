package localizer

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	error_helpers "github.com/yca-software/go-common/error"
	"golang.org/x/text/language"
)

type Translator interface {
	Translate(lang, key string, data map[string]any) string
	TranslateError(lang string, err *error_helpers.Error, data map[string]any)
}

type translator struct {
	bundle             *i18n.Bundle
	supportedLanguages []string
	defaultLanguage    string
}

func NewTranslator(supportedLanguages []string, defaultLanguage string, localesPath string) Translator {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	normalized := make([]string, 0, len(supportedLanguages))
	for _, lang := range supportedLanguages {
		normalized = append(normalized, strings.ToLower(lang))
	}
	defaultLang := strings.ToLower(defaultLanguage)

	for _, lang := range normalized {
		filePath := fmt.Sprintf("%s/%s.json", localesPath, lang)
		if _, err := os.Stat(filePath); err == nil {
			bundle.MustLoadMessageFile(filePath)
		}
	}

	return &translator{
		bundle:             bundle,
		supportedLanguages: normalized,
		defaultLanguage:    defaultLang,
	}
}

func (t *translator) Translate(lang, key string, data map[string]any) string {
	lng := strings.ToLower(strings.TrimSpace(lang))
	if lng == "" || !slices.Contains(t.supportedLanguages, lng) {
		lng = t.defaultLanguage
	}
	if msg, ok := t.localize(lng, key, data); ok {
		return msg
	}
	if lng != t.defaultLanguage {
		if msg, ok := t.localize(t.defaultLanguage, key, data); ok {
			return msg
		}
	}
	return key
}

func (t *translator) localize(lang string, key string, data map[string]any) (string, bool) {
	localizer := i18n.NewLocalizer(t.bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return "", false
	}
	return msg, true
}

func (t *translator) TranslateError(lang string, err *error_helpers.Error, data map[string]any) {
	if err == nil {
		return
	}
	templateData := data
	if templateData == nil && err.Extra != nil {
		switch extra := err.Extra.(type) {
		case map[string]any:
			templateData = extra
		case map[string]string:
			templateData = make(map[string]any, len(extra))
			for k, v := range extra {
				templateData[k] = v
			}
		}
	}

	msg := t.Translate(lang, fmt.Sprintf("error.%d.%s", err.StatusCode, err.ErrorCode), templateData)
	err.Message = msg
}
