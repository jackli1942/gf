package i18nutil_test

import (
	"context"
	"testing"
	"yuncms/internal/utility/i18nutil"
	_ "yuncms/internal/boot" // Ensure boot sequence (and i18n init) runs

	"github.com/gogf/gf/v2/i18n/gi18n"
)

func TestTranslate(t *testing.T) {
	// Default language is 'zh', set by boot.go init via config.
	// i18n instance is configured by boot.go.

    // Test with default language (zh)
    ctxZh := gi18n.WithLanguage(context.Background(), "zh")
    expectedZhHello := "你好世界"
    actualZhHello := i18nutil.T(ctxZh, "hello")
    if actualZhHello != expectedZhHello {
        t.Errorf("Expected '%s' for 'hello' in 'zh', got '%s'", expectedZhHello, actualZhHello)
    }

    expectedZhWelcome := "欢迎使用 云CMS"
    actualZhWelcome := i18nutil.T(ctxZh, "welcome", "Appname", "云CMS")
    if actualZhWelcome != expectedZhWelcome {
        t.Errorf("Expected '%s' for 'welcome' in 'zh' with placeholder, got '%s'", expectedZhWelcome, actualZhWelcome)
    }

    expectedZhUserExists := "用户名 'test' 已存在"
    actualZhUserExists := i18nutil.T(ctxZh, "error_username_exists", "Username", "test")
    if actualZhUserExists != expectedZhUserExists {
         t.Errorf("Expected '%s' for 'error_username_exists' in 'zh' with placeholder, got '%s'", expectedZhUserExists, actualZhUserExists)
    }

    // Test with English
    ctxEn := gi18n.WithLanguage(context.Background(), "en")
    expectedEnHello := "Hello World"
    actualEnHello := i18nutil.T(ctxEn, "hello")
    if actualEnHello != expectedEnHello {
        t.Errorf("Expected '%s' for 'hello' in 'en', got '%s'", expectedEnHello, actualEnHello)
    }

    expectedEnWelcome := "Welcome to YunCMS"
    actualEnWelcome := i18nutil.T(ctxEn, "welcome", "Appname", "YunCMS")
    if actualEnWelcome != expectedEnWelcome {
        t.Errorf("Expected '%s' for 'welcome' in 'en' with placeholder, got '%s'", expectedEnWelcome, actualEnWelcome)
    }

    // Test non-existent key (should return key itself)
    // Use a context with a known language like 'en' or 'zh'
    actualNonExistent := i18nutil.T(ctxEn, "non_existent_key_12345")
    if actualNonExistent != "non_existent_key_12345" {
        t.Errorf("Expected '%s' for non-existent key, got '%s'", "non_existent_key_12345", actualNonExistent)
    }

    // Test fallback to default language if current language file or key is missing
    // Default language is 'zh' as per boot.go and config.yaml
    ctxFr := gi18n.WithLanguage(context.Background(), "fr") // French - no fr.yaml file exists
    actualFallbackHello := i18nutil.T(ctxFr, "hello")
    if actualFallbackHello != expectedZhHello {
        t.Errorf("Expected fallback to default ('%s') for 'hello' in 'fr', got '%s'", expectedZhHello, actualFallbackHello)
    }
}
