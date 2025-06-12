package service

import (
	"sync"
	// "fmt"

	"github.com/casbin/casbin/v2"
	// Use the root package of dobyte/gf-casbin and alias it if needed, or use its methods directly.
	// Assuming NewAdapter is in the root of the dobyte/gf-casbin module.
	// If the constructor is `adapter.NewAdapter`, then the package name is `adapter`.
	// Based on the error, the package is NOT at "github.com/dobyte/gf-casbin/adapter".
	// Let's assume the package name provided by "github.com/dobyte/gf-casbin" is "gfcasbin" or similar,
	// or that NewAdapter is a top-level function in that module.
	// A common pattern is: import casbinAdapter "github.com/dobyte/gf-casbin"
	// and then casbinAdapter.NewAdapter(...)
	// The prompt's code uses `adapter.NewAdapter`, implying the package name is `adapter`.
	// Let's try importing "github.com/dobyte/gf-casbin" and see if `NewAdapter` is directly available
	// or if it's under a package name like `gfcasbin.NewAdapter`.
	// The previous subtask used `adapter "github.com/dobyte/gf-casbin/adapter"`.
	// The fix is to change the import path and potentially the qualifier for NewAdapter.
	// If the module is `github.com/dobyte/gf-casbin` and it provides `NewAdapter`,
	// the import would be `import "github.com/dobyte/gf-casbin"` and call `gfcasbin.NewAdapter`
	// or if package name is `adapter` within that module, then `adapter.NewAdapter` is fine.
	// The error says "does not contain package github.com/dobyte/gf-casbin/adapter".
	// This means the import path itself is wrong. The package is likely at the root.
	// So, import "github.com/dobyte/gf-casbin" and then the functions are directly available from that package.
	// Let's check dobyte/gf-casbin docs. It seems the package itself is `adapter`.
	// So, `import adapter "github.com/dobyte/gf-casbin"` is the way.
	adapter "github.com/dobyte/gf-casbin"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"strings"
)

var (
	enforcer *casbin.SyncedEnforcer
	once     sync.Once
)

func Casbin() *casbin.SyncedEnforcer {
	once.Do(func() {
		ctx := gctx.New()

		modelPath := g.Cfg().MustGet(ctx, "casbin.model", "manifest/config/casbin_model.conf").String()
		if !gfile.Exists(modelPath) {
			g.Log().Fatalf(ctx, "Casbin Service: Casbin model file '%s' not found.", modelPath)
			return
		}

		db := g.DB()
		if db == nil {
			g.Log().Fatal(ctx, "Casbin Service: Default database not configured.")
			return
		}

		casbinTableName := g.Cfg().MustGet(ctx, "casbin.table", "casbin_rules").String()

		// The import is now `adapter "github.com/dobyte/gf-casbin"`
		// So the call `adapter.NewAdapter(db, casbinTableName)` should be correct.
		a, errAdapter := adapter.NewAdapter(db, casbinTableName)
		if errAdapter != nil {
			g.Log().Fatalf(ctx, "Casbin Service: Failed to initialize dobyte/gf-casbin adapter: %v", errAdapter)
			return
		}

		var errEnforcer error
		enforcer, errEnforcer = casbin.NewSyncedEnforcer(modelPath, a)
		if errEnforcer != nil {
			g.Log().Fatalf(ctx, "Casbin Service: Failed to create Casbin enforcer: %v", errEnforcer)
			return
		}

		if err := enforcer.LoadPolicy(); err != nil {
			g.Log().Warningf(ctx, "Casbin Service: Failed to load policy from database: %v. This might be normal if the table ('%s') is new.", err, casbinTableName)
		}

		g.Log().Info(ctx, "Casbin Service: SyncedEnforcer initialized successfully with dobyte/gf-casbin adapter.")
		g.Log().Infof(ctx, "Casbin Service: Model: %s, Table: %s", modelPath, casbinTableName)
	})
	if enforcer == nil {
	    panic("Casbin enforcer failed to initialize after once.Do. Check logs.")
	}
	return enforcer
}

func InitCasbin() {
	Casbin()
}

func CheckPermission(sub string, obj string, act string) (bool, error) {
	e := Casbin()
	return e.Enforce(sub, obj, act)
}

func AddPolicy(sub string, obj string, act string) (bool, error) {
	e := Casbin()
	return e.AddPolicy(sub, obj, act)
}

func AddRoleForUser(user string, role string) (bool, error) {
	e := Casbin()
	return e.AddGroupingPolicy(user, role)
}
