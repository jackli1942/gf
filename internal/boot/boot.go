package boot

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n" // Corrected path
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
)

func init() {
    ctx := context.Background()
    glog.Info(ctx, "Boot sequence started (i18n initialization)...")

    // Configuration Path Setup
    configPath := ""
    // Try common paths for config directory relative to where 'go run main.go' or 'go test' might be executed
    if gfile.Exists("/app/manifest/config") { // Most specific for subtask environment
        configPath = "/app/manifest/config"
    } else if gfile.Exists("manifest/config") { // CWD is project root
        configPath = "manifest/config"
    } else if gfile.Exists("../manifest/config") { // CWD is internal/
        configPath = "../manifest/config"
    } else if gfile.Exists("../../manifest/config") { // CWD is internal/boot or internal/logic etc.
        configPath = "../../manifest/config"
    } else {
        glog.Fatal(ctx, "Boot Init: Configuration directory 'manifest/config' not found at expected paths.")
        return // Stop if config path can't be found
    }

    // Set the found config path for g.Cfg()
    if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
        // Check if a path is already set and valid, to avoid overriding if main.go's init ran first
        // and set a specific path that this init should respect.
        // Since GetPath() was problematic, we'll rely on the order of init() execution
        // or assume that if this init is running, it's responsible for this setup.
        // A simple SetPath should be fine for most cases.
        adapter.SetPath(configPath)
        glog.Debugf(ctx, "Boot Init: Configuration path set to: %s by boot sequence", configPath)
    } else {
        glog.Warning(ctx, "Boot Init: Default config adapter is not *gcfg.AdapterFile. Path not set by boot.")
    }

    // Initialize i18n
    i18nPathKey := "i18n.path"
    defaultI18nPath := "resource/i18n" // Relative to project root
    i18nPath := g.Cfg().MustGet(ctx, i18nPathKey, defaultI18nPath).String()

    defaultLangKey := "i18n.defaultLanguage"
    defaultLang := g.Cfg().MustGet(ctx, defaultLangKey, "zh").String()

    // Adjust i18nPath for execution environment
    effectiveI18nPath := i18nPath
    // Try paths relative to where config was found, or absolute /app path
    if !gfile.Exists(effectiveI18nPath) { // If path from config (e.g. "resource/i18n") is not found directly
        // Check relative to the configPath's parent (effectively project root if configPath is "manifest/config")
        projRoot := gfile.Dir(configPath) // if configPath is "manifest/config", this is "manifest"
        if projRoot == "manifest" { projRoot = gfile.Dir(projRoot) } // get parent of "manifest" -> project root

        if gfile.Exists(gfile.Join(projRoot, i18nPath)) {
             effectiveI18nPath = gfile.Join(projRoot, i18nPath)
        } else if gfile.Exists(gfile.Join("/app", i18nPath)) { // Common for containerized execution
            effectiveI18nPath = gfile.Join("/app", i18nPath)
        }
        // Add more specific checks if needed, e.g. for tests in subdirs:
        // else if gfile.Exists(gfile.Join(projRoot, "..", i18nPath)) { ... }
    }

    glog.Infof(ctx, "Initializing i18n: Effective Path='%s', DefaultLanguage='%s'", effectiveI18nPath, defaultLang)
    if !gfile.Exists(effectiveI18nPath){
        glog.Fatalf(ctx, "i18n resource path does not exist: %s. Configured path: %s", effectiveI18nPath, i18nPath)
    }

    gi18n.Instance().SetPath(effectiveI18nPath)
    gi18n.Instance().SetLanguage(defaultLang)

    testKey := "hello"
    translated := gi18n.Instance().T(ctx, testKey)
    if translated == testKey || translated == "" {
         glog.Warningf(ctx, "i18n test translation for key '%s' in default language '%s' failed. Got: '%s'. Check path ('%s') and i18n files.",
            testKey, defaultLang, translated, effectiveI18nPath)
    } else {
         glog.Infof(ctx, "i18n test translation for key '%s' in default lang '%s': '%s'", testKey, defaultLang, translated)
    }

    glog.Info(ctx, "Boot sequence (i18n initialization) finished.")
}
