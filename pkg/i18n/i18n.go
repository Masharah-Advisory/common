package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	bundle     *i18n.Bundle
	localizers map[string]*i18n.Localizer
	initOnce   sync.Once
	mutex      sync.RWMutex
)

// Setup initializes the i18n system with a locales directory
func Setup(localesDir string) error {
	var err error
	initOnce.Do(func() {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
		localizers = make(map[string]*i18n.Localizer)

		// Load all locale files
		err = filepath.Walk(localesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".json" {
				_, loadErr := bundle.LoadMessageFile(path)
				if loadErr != nil {
					return fmt.Errorf("failed to load locale file %s: %w", path, loadErr)
				}
			}
			return nil
		})

		// Create localizers for supported languages
		localizers["en"] = i18n.NewLocalizer(bundle, "en")
		localizers["ar"] = i18n.NewLocalizer(bundle, "ar")
	})
	return err
}

// Middleware returns a Gin middleware that automatically detects language
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := detectLanguage(c)
		c.Set("lang", lang)
		c.Next()
	}
}

// T translates a message for the current request
func T(c *gin.Context, key string, data ...map[string]interface{}) string {
	lang := getLang(c)

	mutex.RLock()
	localizer, exists := localizers[lang]
	mutex.RUnlock()

	if !exists {
		localizer = localizers["en"] // fallback
	}

	var templateData map[string]interface{}
	if len(data) > 0 {
		templateData = data[0]
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})
	if err != nil {
		return key // fallback to key
	}
	return msg
}

// detectLanguage gets language from headers with fallback to "en"
func detectLanguage(c *gin.Context) string {
	// Check X-Language header first
	if lang := c.GetHeader("X-Language"); lang != "" {
		return normalizeLang(lang)
	}

	// Check Accept-Language header
	if accept := c.GetHeader("Accept-Language"); accept != "" {
		if lang := parseAcceptLanguage(accept); lang != "" {
			return normalizeLang(lang)
		}
	}

	// Default to English
	return "en"
}

// getLang gets language from context with fallback
func getLang(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "en"
}

// normalizeLang converts language codes to supported format
func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	// Handle language-region codes (e.g., en-US -> en)
	if i := strings.Index(lang, "-"); i != -1 {
		lang = lang[:i]
	}

	switch lang {
	case "ar", "arabic":
		return "ar"
	case "en", "english":
		return "en"
	default:
		return "en"
	}
}

// parseAcceptLanguage parses Accept-Language header
func parseAcceptLanguage(accept string) string {
	languages := strings.Split(accept, ",")
	if len(languages) == 0 {
		return "en"
	}

	// Get first language
	firstLang := strings.TrimSpace(languages[0])
	if i := strings.Index(firstLang, ";"); i != -1 {
		firstLang = firstLang[:i]
	}

	return firstLang
}
