package logic

import (
	"context" // Ensure context is imported
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	// Import the hailaz gf-casbin-adapter v2
	// Note: The actual import path might vary if it's a fork or specific version not on main proxy.
	// Assuming "github.com/hailaz/gf-casbin-adapter/v2" is the correct, fetchable path.
	adapter "github.com/hailaz/gf-casbin-adapter/v2"
)

var (
	enforcerHailaz *casbin.Enforcer // Renamed to avoid conflict with any previous attempts
	onceHailaz     sync.Once        // Renamed
	initHailazErr  error            // Renamed
)

// CasbinEnforcer returns the Casbin enforcer instance.
// It initializes the enforcer on first call using hailaz/gf-casbin-adapter.
func CasbinEnforcer() (*casbin.Enforcer, error) {
	onceHailaz.Do(func() {
		ctx := context.Background() // Use a background context for initialization logging/config
		modelPath := g.Cfg().MustGet(ctx, "casbin.modelPath").String()
		if !gfile.Exists(modelPath) {
			initHailazErr = gerror.Newf("Casbin model file '%s' not found", modelPath)
			g.Log().Error(ctx, initHailazErr)
			return
		}

		// Initialize the hailaz gf-casbin-adapter.
		opt := adapter.Options{
			GDB: g.DB("default"),
			// TableName: "casbin_rule", // Default is "casbin_rule" for this adapter
		}
		a, adapterErr := adapter.NewAdapter(opt)
		if adapterErr != nil {
			initHailazErr = gerror.Wrap(adapterErr, "Failed to initialize hailaz/gf-casbin-adapter")
			g.Log().Error(ctx, initHailazErr)
			return
		}

		e, enforcerErr := casbin.NewEnforcer(modelPath, a)
		if enforcerErr != nil {
			initHailazErr = gerror.Wrap(enforcerErr, "Failed to create Casbin enforcer with hailaz/gf-casbin-adapter")
			g.Log().Error(ctx, initHailazErr)
			return
		}

		loadPolicyErr := e.LoadPolicy()
		if loadPolicyErr != nil {
			g.Log().Warningf(ctx, "Casbin (hailaz): Failed to load policy: %v. (Normal if no policies in DB yet)", loadPolicyErr)
		} else {
			g.Log().Info(ctx, "Casbin (hailaz): Policies loaded successfully from database.")
		}
		enforcerHailaz = e
		g.Log().Info(ctx, "Casbin enforcer with hailaz/gf-casbin-adapter initialized.")
	})

	if initHailazErr != nil {
		return nil, initHailazErr
	}
	if enforcerHailaz == nil && initHailazErr == nil { // Safeguard
		return nil, gerror.New("Casbin enforcer (hailaz) is nil after initialization and no error was recorded")
	}
	return enforcerHailaz, nil
}

// InitCasbinEnforcer is a helper to explicitly initialize Casbin on app start.
func InitCasbinEnforcer(ctx context.Context) {
	_, err := CasbinEnforcer()
	if err != nil {
		g.Log().Fatalf(ctx, "Fatal: Casbin Enforcer (hailaz) failed to initialize during app startup: %+v", err)
	}
	g.Log().Info(ctx, "Casbin Enforcer (hailaz) successfully initialized on app startup.")
}
