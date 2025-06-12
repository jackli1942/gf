package cmd

import (
	"context"
	"fmt" // Needed for fmt.Errorf

	"github.com/gogf/gf/v2/frame/g"
	// "github.com/gogf/gf/v2/i18n/gi18n" // Not directly used, g.I18n() is used
	// "github.com/gogf/gf/v2/net/ghttp" // Not directly used here
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
	"yuncms/internal/router"
	"yuncms/internal/service" // For Casbin Init
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start yuncms server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			configPath := "manifest/config/config.yaml"
			if !gfile.Exists(configPath) {
			    altConfigPath := "config.yaml"
			    if gfile.Exists(altConfigPath) {
			        configPath = altConfigPath
			    } else {
			        g.Log().Fatalf(ctx, "Configuration file '%s' (or '%s') not found.", "manifest/config/config.yaml", altConfigPath)
			        return fmt.Errorf("configuration file not found") // fmt needed here
			    }
			}

			configAdapter, err := gcfg.NewAdapterFile(configPath)
			if err != nil {
				g.Log().Fatalf(ctx, "Failed to create config adapter from '%s': %v", configPath, err)
				return err
			}
			g.Cfg().SetAdapter(configAdapter)

			i18nPath := g.Cfg().MustGet(ctx, "application.i18nPath", "resource/i18n").String()
			defaultLang := g.Cfg().MustGet(ctx, "application.defaultLanguage", "zh").String() // Default zh

			g.Log().Infof(ctx, "Initializing i18n with path: '%s' and default language: '%s'", i18nPath, defaultLang)
			if err := g.I18n().SetPath(i18nPath); err != nil {
				g.Log().Fatalf(ctx, "Failed to set i18n path: %v", err)
				return err
			}
			g.I18n().SetLanguage(defaultLang)


			// Initialize Casbin Enforcer
			g.Log().Info(ctx, "Initializing Casbin service...")
			service.InitCasbin()
			g.Log().Info(ctx, "Casbin service initialization attempted.")
	s := g.Server()

			// --- Frontend Static File Serving Setup ---
			frontendStaticPath := g.Cfg().MustGet(ctx, "server.frontendStaticPath", "web/admin/dist").String()
			adminRoutePrefix := g.Cfg().MustGet(ctx, "server.adminRoutePrefix", "/admin").String()

			if adminRoutePrefix != "" {
			    if adminRoutePrefix[0] != '/' {
			        adminRoutePrefix = "/" + adminRoutePrefix
			    }
			}

			s.AddStaticPath(adminRoutePrefix, frontendStaticPath)
			g.Log().Infof(ctx, "Serving static files from path '%s' under URL prefix '%s'", frontendStaticPath, adminRoutePrefix)

			if adminRoutePrefix != "" {
			    spaIndexFile := frontendStaticPath + "/index.html"
			    // Ensure this rewrite is specific enough. GoFrame's default behavior for static files
			    // might already serve index.html for directory requests.
			    // SetRewrite is powerful; /* might be too broad if not careful.
			    // A common pattern for SPA is to serve index.html for any path under prefix that is NOT a file.
			    // GoFrame's s.SetRewrite(prefix, path) serves `path` IF request path starts with `prefix`
			    // AND the original path is not found (404). This is suitable for SPA.
			    s.SetRewrite(adminRoutePrefix, spaIndexFile) // If /admin/anything is 404, serve /admin/index.html
			    // For more fine-grained control (e.g. not rewriting if /admin/assets/some.js exists but /admin/users does not),
			    // one might need a middleware. But AddStaticPath + SetRewrite is often sufficient.
			    // The prompt uses s.SetRewrite(adminRoutePrefix+"/*", spaIndexFile)
			    // Let's stick to the prompt's specific pattern for SetRewrite.
			    s.SetRewrite(adminRoutePrefix+"/*", spaIndexFile)
			    g.Log().Infof(ctx, "SPA fallback configured for prefix '%s/*' to serve '%s'", adminRoutePrefix, spaIndexFile)
			} else {
			    g.Log().Info(ctx, "Frontend at root ('/') detected. SPA fallback needs careful manual rule or middleware if not using distinct admin prefix.")
			}
			// --- End Frontend Static File Serving Setup ---

			router.BindController(s)

			serverAddr := g.Cfg().MustGet(ctx, "server.address", ":8081").String()

			g.Log().Infof(ctx, "Starting Yuncms server on %s", serverAddr)
			s.SetAddr(serverAddr)
			s.Run()
			return nil
		},
	}
)

// Removed the helper 'import "fmt"' as it's not needed in Go module mode if fmt is used above.
// It will be auto-managed by go mod tidy or go build.
// If fmt was truly unused, `go mod tidy` would remove it. The prompt had it for the error return.
// My version above uses fmt.Errorf, so the import is necessary.
// `goimports` tool would handle this. I will ensure `fmt` is imported.
// The provided code already had `import "fmt"`.
