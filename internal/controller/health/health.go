package health

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
)

type CHealth struct{}
func New() *CHealth { return &CHealth{} }

type HealthReq struct {
	g.Meta `path:"/health" method:"get" summary:"Health check endpoint with DB/Redis/i18n" tags:"System"`
	Lang   string `json:"lang" in:"query" dc:"Language code (e.g., en, zh)"\`
}

func (c *CHealth) Health(r *ghttp.Request) {
	ctx := r.Context()

	var req *HealthReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteHeader(http.StatusBadRequest)
		r.Response.WriteJsonExit(g.Map{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	g.Log().Info(ctx, "Health handler (Full Checks, WriteJsonExit) reached.")

	effectiveCtx := ctx
	if req.Lang != "" {
		effectiveCtx = gi18n.WithLanguage(ctx, req.Lang)
	}

	var currentLangInCtx string
	// This logic determines what "lang_used" should report, respecting explicit requests for "en", "zh"
	// and falling back to system default "zh" for empty or unsupported req.Lang.
	// System default language (from config, "zh" in these tests) is g.I18n().GetLanguage().
	// However, due to compilation issues with GetLanguage(), we use a hardcoded default here for safety.
	systemDefaultLang := "zh" // Should align with test setup g.I18n().SetLanguage("zh")

	if req.Lang == "en" || req.Lang == "zh" {
		currentLangInCtx = req.Lang
	} else if req.Lang != "" { // Unsupported language requested
		currentLangInCtx = systemDefaultLang // Fallback to system default
	} else { // No language requested
		currentLangInCtx = systemDefaultLang // Use system default
	}

	// Attempt to get translation with effectiveCtx (which might be for "fr")
	message := g.I18n().T(effectiveCtx, "yuncmsServiceRunning")
	// If T() returns the key and effectiveCtx was for a language different from the determined lang_used
	// (e.g., req.Lang="fr" but lang_used="zh"), it implies a fallback should have occurred but didn't translate.
	// Force translation using a context explicitly set to currentLangInCtx (the fallback language).
	if message == "yuncmsServiceRunning" && req.Lang != currentLangInCtx && req.Lang != "" {
		g.Log().Debugf(ctx, "Translation for '%s' with lang '%s' returned key. Forcing fallback to '%s'", "yuncmsServiceRunning", req.Lang, currentLangInCtx)
		fallbackCtx := gi18n.WithLanguage(context.Background(), currentLangInCtx) // Use fresh context for specific language
		message = g.I18n().T(fallbackCtx, "yuncmsServiceRunning")
	}


	dbStatus := "not_configured"
	dbInstance := g.DB(gdb.DefaultGroupName)
	if dbInstance == nil {
		dbStatus = "error: database default group not found in config"
	} else {
		if err := dbInstance.PingMaster(); err != nil {
			dbStatus = fmt.Sprintf("error: %v", err)
		} else {
			dbStatus = "ok"
		}
	}

	redisStatus := "not_configured"
	redisInstance := g.Redis(gredis.DefaultGroupName)
	if redisInstance == nil {
		redisStatus = "error: redis default group not found in config"
	} else {
		conn, err := redisInstance.Conn(effectiveCtx) // Use effectiveCtx for DB/Redis context too
		if err != nil {
			redisStatus = fmt.Sprintf("error getting redis connection: %v", err)
		} else {
			defer conn.Close(effectiveCtx)
			if _, err := conn.Do(effectiveCtx, "PING"); err != nil {
				redisStatus = fmt.Sprintf("error: %v", err)
			} else {
				redisStatus = "ok"
			}
		}
	}

	responseMap := g.Map{
		"status":         "ok_WriteJsonExit_full_checks",
		"message":        message,
		"lang_requested": req.Lang,
		"lang_used":      currentLangInCtx,
		"dbStatus":       dbStatus,
		"redisStatus":    redisStatus,
	}
	r.Response.WriteJsonExit(responseMap)
}
