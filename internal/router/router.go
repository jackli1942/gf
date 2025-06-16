package router

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
    "github.com/gogf/gf/v2/i18n/gi18n" // Corrected import path from previous subtasks
    "github.com/gogf/gf/v2/text/gstr"
    // "yuncms/internal/controller" // Only if a common/default controller is used here
)

// LanguageMiddleware sets the language for the current request context based on headers.
func LanguageMiddleware(r *ghttp.Request) {
    // Priority: Custom Header > Standard Header
    langSource := ""
    customLangHeader := r.Header.Get("X-Api-Language")
    standardLangHeader := r.Header.Get("Accept-Language")

    if customLangHeader != "" {
        langSource = customLangHeader
    } else if standardLangHeader != "" {
        langSource = standardLangHeader
    }

    if langSource != "" {
        // Extract the primary language part (e.g., "en" from "en-US,en;q=0.9")
        parts := gstr.Split(gstr.Split(langSource, ",")[0], ";")
        finalLang := gstr.Trim(parts[0])
        if finalLang != "" {
             // SetLanguage will use the default from gi18n.Instance() if finalLang is not supported
             gi18n.SetLanguage(r.Context(), finalLang)
             g.Log().Debugf(r.Context(), "Language resolved to '%s' (from header source: '%s') for current request", finalLang, langSource)
        } else {
             g.Log().Debugf(r.Context(), "No valid language found in header: '%s'. Using default.", langSource)
        }
    } else {
        g.Log().Debugf(r.Context(), "No language headers (X-Api-Language, Accept-Language) found. Using default language.")
    }
    r.Middleware.Next()
}

// BindCentralRouter configures and binds all routes for the server.
func BindCentralRouter(s *ghttp.Server) {
    s.Use(ghttp.MiddlewareHandlerResponse) // Standardizes response structure
    s.Use(LanguageMiddleware)              // Applies language detection globally

	s.Group("/api/v1", func(apiV1Group *ghttp.RouterGroup) { // Changed 'group' to 'apiV1Group'
		RegisterUserRoutes(apiV1Group) // Register user-specific routes
		// Future: RegisterOtherModuleRoutes(apiV1Group)
	})

    // Optional: A default route for / or /ping for health checks
    s.GET("/", func(r *ghttp.Request){
        r.Response.Write("yuncms api is running.")
    })
    s.GET("/ping", func(r *ghttp.Request){
        r.Response.Write("pong")
    })
}
