package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/gorm-adapter/v3"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"

	// GORM MySQL driver, needed by gormadapter
	_ "gorm.io/driver/mysql"
	// GoFrame MySQL Driver
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

var (
	enforcer  *casbin.Enforcer
	once      sync.Once
	initError error // Store initialization error
)

func init() {
    ctx := context.Background()
    configPath := ""
    // Prioritize /app/manifest/config for subtask/container environments
    if gfile.Exists("/app/manifest/config") {
        configPath = "/app/manifest/config"
    } else if gfile.Exists("manifest/config") { // For running from project root
        configPath = "manifest/config"
    } else if gfile.Exists("../../manifest/config") { // For tests running from package dir (like internal/service)
        configPath = "../../manifest/config"
    }

    if configPath != "" {
        if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
            // Check if a path is already set. Since GetPath() was an issue,
            // we assume that if this init() is called, it's responsible for path setting,
            // or it's okay to ensure the path is set.
            // A more sophisticated approach might involve a global flag or checking g.Cfg().Get() for a known key.
            adapterFile.SetPath(configPath)
            glog.Debug(ctx, "Casbin Service Init: Configuration path set to:", configPath)
        } else {
            glog.Warning(ctx, "Casbin Service Init: Default config adapter is not *gcfg.AdapterFile.")
        }
    } else {
        glog.Warning(ctx, "Casbin Service Init: Config directory not found via common paths. Relying on GoFrame default search or main init.")
    }
}

// GetEnforcer initializes and returns the Casbin enforcer instance.
func GetEnforcer(ctx context.Context) (*casbin.Enforcer, error) {
	once.Do(func() {
		dbHost := g.Cfg().MustGet(ctx, "database.default.host").String()
		dbPort := g.Cfg().MustGet(ctx, "database.default.port").String()
		dbUser := g.Cfg().MustGet(ctx, "database.default.user").String()
		dbPass := g.Cfg().MustGet(ctx, "database.default.pass").String()
		dbName := g.Cfg().MustGet(ctx, "database.default.name").String()

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPass, dbHost, dbPort, dbName)
		glog.Debugf(ctx, "Casbin GORM Adapter DSN: %s", dsn)

		var adapter *gormadapter.Adapter
		adapter, initError = gormadapter.NewAdapter("mysql", dsn, true)
		if initError != nil {
			log.Printf("Failed to create Casbin GORM adapter: %v", initError)
			glog.Errorf(ctx, "Failed to create Casbin GORM adapter: %v", initError)
			return
		}

		modelPathCfgKey := "casbin.model"
		defaultModelPath := "manifest/config/casbin_model.conf"
		modelPath := g.Cfg().MustGet(ctx, modelPathCfgKey, defaultModelPath).String()

        resolvedModelPath := ""
        // Check paths in order of preference/likelihood
        if gfile.Exists(modelPath) { // Path as-is (could be absolute or already correct relative)
            resolvedModelPath = modelPath
        } else if gfile.Exists("/app/" + modelPath) { // Path relative to /app
             resolvedModelPath = "/app/" + modelPath
        } else if gfile.Exists("../../" + modelPath) { // Path relative to this file's dir, going up two levels
             resolvedModelPath = "../../" + modelPath
        }


		glog.Debugf(ctx, "Casbin Model Path configured as '%s', attempting to resolve. Final path used: '%s'", modelPath, resolvedModelPath)
        if resolvedModelPath == "" || !gfile.Exists(resolvedModelPath) {
            initError = fmt.Errorf(
				"casbin model file not found. Configured: '%s', Default: '%s'. Checked: '%s', '/app/%s', '../../%s'",
				modelPath, defaultModelPath, modelPath, modelPath, modelPath, // Show originally checked paths
			)
            log.Printf("%v", initError)
            glog.Errorf(ctx, "%v", initError)
            return
        }

		enforcer, initError = casbin.NewEnforcer(resolvedModelPath, adapter)
		if initError != nil {
			log.Printf("Failed to create Casbin enforcer: %v", initError)
            glog.Errorf(ctx, "Failed to create Casbin enforcer: %v", initError)
			return
		}

		loadErr := enforcer.LoadPolicy()
		if loadErr != nil {
			log.Printf("Warning: Failed to load policy from DB: %v. This might be normal if no policies are set yet.", loadErr)
            glog.Warningf(ctx, "Failed to load policy from DB: %v. This might be normal if no policies are set yet.", loadErr)
		}
        glog.Info(ctx, "Casbin enforcer initialized successfully.")
	})

    if initError != nil {
        return nil, initError
    }
    if enforcer == nil && initError == nil {
         return nil, fmt.Errorf("enforcer is nil after initialization attempt, but no specific error was captured")
    }
	return enforcer, nil
}
