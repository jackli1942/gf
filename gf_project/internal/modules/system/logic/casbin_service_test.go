package logic

import (
	"context"
	"os"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/test/gtest"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	_ "github.com/hailaz/gf-casbin-adapter/v2"
)

// TestMain sets up the testing environment.
func TestMain(m *testing.M) {
	ctx := gctx.GetGlobal() // Use a global context for setup logging

	// Determine project root based on current file's location.
	// This assumes the test is run from within the 'internal/modules/system/logic' directory.
	// Adjust if 'go test ./...' is run from project root.
	// For 'go test' within package dir:
	// Pwd is logic dir. projectRoot is logic/../../..
	projectRoot := gfile.normalize(gfile.Pwd() + "/../../../../") // from logic to project root

	// If projectRoot is not found correctly, try to use an environment variable or skip config loading.
	// This is a common challenge in tests.
	// For now, assume manifest/config is relative to where  is invoked if projectRoot is tricky.
	// Best practice is often to have test-specific config files or use build tags for test configs.

	// Try to set config path for g.Cfg()
	// This should point to the directory containing config.yaml
	configDirPath := gfile.Join(projectRoot, "manifest", "config")

	// Check if default config file (config.yaml) exists at the calculated path.
	// If 'go test ./...' is run from root, g.Cfg() might find it if AddPath("manifest/config") was used.
	// However, explicit SetPath is more reliable for tests.
	if !gfile.Exists(gfile.Join(configDirPath, "config.yaml")) {
		// Fallback: if running 'go test' from package dir, default adapter might not find manifest/config
		// Try to set path relative to where command is run if absolute fails.
		// This is often tricky. A common pattern is to copy test configs into the test dir
		// or have a dedicated test config file.
		// For this exercise, we assume the primary config.yaml is the one to load.
		g.Log().Warningf(ctx, "TestMain: config.yaml not found at '%s'. Ensure tests are run from project root or path is correctly set.", configDirPath)
		// Attempt to load config from a path relative to where  might be run (project root)
		// This path would be manifest/config/config.yaml
		// If GFS_GCFG_PATH is set, it might also work.
		// Let's assume Init() in casbin_service will try to load config.
		// We need to ensure it finds casbin.modelPath.
	} else {
		g.Cfg().GetAdapter().(*gcfg.AdapterFile).SetPath(configDirPath)
		g.Log().Infof(ctx, "TestMain: Config path set to: %s", configDirPath)
	}

	// Ensure casbin_model.conf exists and its path is correctly interpreted by CasbinEnforcer.
	// CasbinEnforcer reads "casbin.modelPath" from config.
	modelPathFromConfig := g.Cfg().MustGet(ctx, "casbin.modelPath", "manifest/config/casbin_model.conf").String()
	resolvedModelPath := modelPathFromConfig
	if !gfile.IsAbs(modelPathFromConfig) {
		// If path is relative, it's relative to where binary runs OR config path.
		// If config path was set, Cfg adapter should make it relative to that.
		// If config path not set, it's relative to PWD.
		// Let's try to make it absolute from projectRoot if path not found directly.
		if gfile.Exists(gfile.Join(configDirPath, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(configDirPath, modelPathFromConfig)
		} else if gfile.Exists(gfile.Join(projectRoot, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(projectRoot, modelPathFromConfig)
		}
	}
	if !gfile.Exists(resolvedModelPath) {
		g.Log().Fatalf(ctx, "TestMain: Casbin model file '%s' (resolved from config value '%s') not found. Tests cannot proceed.", resolvedModelPath, modelPathFromConfig)
	}
	// Override casbin.modelPath in config for the test run to ensure absolute path is used by service.
	// This is a robust way to ensure the service finds the file during testing.
	g.Cfg().Set("casbin.modelPath", resolvedModelPath)
	g.Log().Infof(ctx, "TestMain: Overriding casbin.modelPath for test to: %s", resolvedModelPath)

	// Database gflite.db and its casbin_rule table must exist from previous steps.
	// A better test suite might initialize a temporary test DB.
	// For now, we rely on the existing gflite.db.

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestCasbinEnforcer_Initialization(t *testing.T) {
	ctx := gctx.New() // Fresh context for each test
	gtest.C(t, func(t *gtest.T) {
		// The CasbinEnforcer uses sync.Once. To test its initialization logic multiple times
		// or under different conditions, the package-level 'once' and 'enforcer' variables
		// would need to be reset. This is not possible for unexported variables from a test package.
		// Thus, this test effectively checks the first-time initialization or subsequent calls
		// which will return the already initialized instance.

		// If there was an error during TestMain's config setup affecting CasbinEnforcer,
		// that error would be caught here.
		e, err := CasbinEnforcer()

		t.AssertNil(err, "CasbinEnforcer() should not return an error on successful initialization.")
		t.AssertNE(e, nil, "CasbinEnforcer() should return a non-nil enforcer instance.")

		// If you want to test specific policies, you'd add them here and use e.Enforce(...)
		// Example:
		// _, err = e.AddPolicy("test_sub", "/test/obj", "read")
		// t.AssertNil(err, "AddPolicy should not fail")
		// allowed, err := e.Enforce("test_sub", "/test/obj", "read")
		// t.AssertNil(err, "Enforce should not fail")
		// t.Assert(allowed, "Enforce should allow access to defined policy")
		// _, err = e.RemovePolicy("test_sub", "/test/obj", "read") // Clean up
		// t.AssertNil(err)
	})
}

// TODO: Add tests for policy enforcement if Casbin rules are added.
// TODO: Consider testing with a temporary/in-memory SQLite DB for better isolation.
