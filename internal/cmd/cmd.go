package cmd

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	// "github.com/gogf/gf/v2/i18n/gi18n" // Removed as per compiler error
	// "github.com/gogf/gf/v2/net/ghttp" // Removed as per compiler error
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
	"yuncms/internal/router"
	"yuncms/internal/app/service" // Ensure updated import path
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
			        return fmt.Errorf("configuration file not found")
			    }
			}

			configAdapter, err := gcfg.NewAdapterFile(configPath)
			if err != nil {
				g.Log().Fatalf(ctx, "Failed to create config adapter from '%s': %v", configPath, err)
				return err
			}
			g.Cfg().SetAdapter(configAdapter)

			i18nPath := g.Cfg().MustGet(ctx, "application.i18nPath", "resource/i18n").String()
			defaultLang := g.Cfg().MustGet(ctx, "application.defaultLanguage", "zh").String()

			g.Log().Infof(ctx, "Initializing i18n with path: '%s' and default language: '%s'", i18nPath, defaultLang)
			if err := g.I18n().SetPath(i18nPath); err != nil {
				g.Log().Fatalf(ctx, "Failed to set i18n path: %v", err)
				return err
			}
			g.I18n().SetLanguage(defaultLang)

			g.Log().Info(ctx, "Initializing Casbin service via service.InitCasbin()...")
			service.InitCasbin()
			g.Log().Info(ctx, "Casbin service initialization via service.InitCasbin() attempted.")

			s := g.Server()

			frontendStaticPath := g.Cfg().MustGet(ctx, "server.frontendStaticPath", "web/admin/dist").String()
			adminRoutePrefix := g.Cfg().MustGet(ctx, "server.adminRoutePrefix", "/admin").String()

			if adminRoutePrefix != "" && adminRoutePrefix[0] != '/' {
				adminRoutePrefix = "/" + adminRoutePrefix
			}

			if gfile.Exists(frontendStaticPath) {
				s.AddStaticPath(adminRoutePrefix, frontendStaticPath)
				g.Log().Infof(ctx, "Serving static files from path '%s' under URL prefix '%s'", frontendStaticPath, adminRoutePrefix)

				spaIndexFile := frontendStaticPath + "/index.html"
				if adminRoutePrefix != "" {
					s.SetRewrite(adminRoutePrefix+"/*", spaIndexFile)
					g.Log().Infof(ctx, "SPA fallback configured for prefix '%s/*' to serve '%s'", adminRoutePrefix, spaIndexFile)
				} else {
					g.Log().Info(ctx, "Frontend at root ('/') detected. SPA fallback for root needs careful manual rule or middleware if not using distinct admin prefix for static serving.")
				}
			} else {
				g.Log().Warningf(ctx, "Frontend static path '%s' not found. Frontend will not be served. This is normal if frontend has not been built yet.", frontendStaticPath)
			}

			router.BindController(s)

			serverAddr := g.Cfg().MustGet(ctx, "server.address", ":8081").String()

			g.Log().Infof(ctx, "Starting Yuncms server on %s", serverAddr)
			s.SetAddr(serverAddr)
			s.Run()
			return nil
		},
	}
)
