package service_test

import (
	"context"
	"testing"
	"yuncms/internal/service"

	"github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2" // GoFrame MySQL Driver
    _ "gorm.io/driver/mysql" // GORM MySQL driver
)

func init() {
    ctx := context.Background()
    configPath := ""
    // Determine config path based on common execution CWDs for tests
    if gfile.Exists("/app/manifest/config") { // Standard for subtask environment if CWD is /app
        configPath = "/app/manifest/config"
    } else if gfile.Exists("manifest/config") { // If CWD is project root
         configPath = "manifest/config"
    } else if gfile.Exists("../../manifest/config") { // If CWD is internal/service
        configPath = "../../manifest/config"
    } else {
        glog.Warning(ctx, "Casbin Test Init: Config directory `manifest/config` not found via common paths. Relying on GoFrame default search or main init if it runs first.")
    }

    if configPath != "" {
        if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
            adapterFile.SetPath(configPath) // Set path for tests
            glog.Debug(ctx, "Casbin Test Init: Configuration path set to:", configPath)
        } else {
            glog.Warning(ctx, "Casbin Test Init: Default config adapter is not *gcfg.AdapterFile.")
        }
    }

    // The GetEnforcer service itself has logic to find the model file based on
    // the "casbin.model" key in the config.yaml. We just need to ensure config.yaml is loaded.
    // No need to g.Cfg().Set("casbin.model", ...) here.
    // We can log if the model file seems resolvable from common paths, for debugging.
    defaultModelPath := "manifest/config/casbin_model.conf"
    resolvedModelPath := ""
    if gfile.Exists(defaultModelPath) {
        resolvedModelPath = defaultModelPath
    } else if gfile.Exists("/app/" + defaultModelPath) {
         resolvedModelPath = "/app/" + defaultModelPath
    } else if gfile.Exists("../../" + defaultModelPath) {
         resolvedModelPath = "../../" + defaultModelPath
    }
    if resolvedModelPath != "" {
        glog.Debugf(ctx, "Casbin Test Init: casbin_model.conf seems resolvable by GetEnforcer to: %s (or similar path based on its own logic)", resolvedModelPath)
    } else {
        glog.Warningf(ctx, "Casbin Test Init: casbin_model.conf not found at common test paths by test init. Relying entirely on service's discovery logic for model file.")
    }
}

func TestCasbinEnforcer(t *testing.T) {
	ctx := context.Background()

	enf, err := service.GetEnforcer(ctx)
	if err != nil {
        t.Fatalf("GetEnforcer returned error on init: %v", err)
    }
	if enf == nil {
        t.Fatalf("Enforcer should not be nil after GetEnforcer without error")
    }

	db := g.DB()
	tables, errDb := db.Tables(ctx)
	if errDb != nil {
        t.Fatalf("Fetching tables errored: %v", errDb)
    }

	var foundCasbinRuleTable bool
	for _, tableName := range tables {
		if tableName == "casbin_rule" {
			foundCasbinRuleTable = true
			break
		}
	}
	if !foundCasbinRuleTable {
        t.Errorf("casbin_rule table should be created by adapter, but not found. Tables found: %v", tables)
    }

	sub := "alice"
	obj := "data1"
	act := "read"

	removed, err := enf.RemovePolicy(sub, obj, act)
    if err != nil {
        glog.Debugf(ctx, "Note: Error during pre-test RemovePolicy (expected if policy didn't exist): %v", err)
    }
    if removed {
         glog.Infof(ctx, "Pre-test policy removed for %s, %s, %s", sub, obj, act)
    }

	added, err := enf.AddPolicy(sub, obj, act)
	if err != nil {
        t.Fatalf("AddPolicy errored: %v", err)
    }
	if !added {
        has, _ := enf.HasPolicy(sub, obj, act)
        if !has {
            t.Errorf("AddPolicy returned false, and policy does not exist. It should have been added.")
        } else {
             glog.Info(ctx, "AddPolicy returned false, likely because policy already existed. This is acceptable.")
        }
    }

	allowed, err := enf.Enforce(sub, obj, act)
	if err != nil {
        t.Fatalf("Enforce (allow) errored: %v", err)
    }
	if !allowed {
        t.Errorf("Enforce should return true for added/existing policy (%s, %s, %s)", sub, obj, act)
    }

	allowed, err = enf.Enforce(sub, obj, "write")
	if err != nil {
        t.Fatalf("Enforce (deny) errored: %v", err)
    }
	if allowed {
        t.Errorf("Enforce should return false for non-existent policy (%s, %s, %s)", sub, obj, "write")
    }

	removed, err = enf.RemovePolicy(sub, obj, act)
	if err != nil {
        t.Fatalf("RemovePolicy errored: %v", err)
    }
	if !removed {
        hasPolicy, _ := enf.HasPolicy(sub, obj, act)
        if hasPolicy {
             t.Errorf("RemovePolicy returned false, but policy still exists.")
        } else {
             glog.Warningf(ctx, "RemovePolicy returned false, and policy (%s, %s, %s) was not found. This might be okay if AddPolicy initially found it existing.", sub, obj, act)
        }
    }

	allowed, err = enf.Enforce(sub, obj, act)
	if err != nil {
        t.Fatalf("Enforce (after remove) errored: %v", err)
    }
	if allowed {
        t.Errorf("Enforce should return false for removed policy (%s, %s, %s)", sub, obj, act)
    }
}
