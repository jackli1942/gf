package main

import (
	"context"
	"yuncms/internal/cmd" // Import your new migrate command package
	// _ "yuncms/internal/packed" // Import packed for side effects if used - Commented out as not used yet

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"

    // Import the MySQL driver
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func main() {
	// Set config path for main execution context
    // This ensures that no matter how main.go is run, it finds its config.
    // This is similar to what's done in the migrate.go init()
    configPath := "manifest/config" // Relative path from project root
	ctx := context.Background() // Create a context for main
    // For robust path detection, you might use GetMainPkgPath or runtime caller,
    // but for subtasks and typical project runs, this relative path is often sufficient.
    // If running from a different working directory, this might need adjustment.
    // The one in migrate.go's init() is more for when migrate.go is used as a library.
    if !gfile.Exists(configPath) {
         // Attempt an absolute path if running in subtask environment
         altPath := "/app/manifest/config"
         if gfile.Exists(altPath) {
             configPath = altPath
         }
    }
    // It's good practice to set this early.
    // Using (*gcfg.AdapterFile) assumes the default adapter is file-based.
    if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
        adapterFile.SetPath(configPath)
        glog.Debug(ctx, "Main: Configuration path set to:", configPath)
    } else {
        glog.Warning(ctx, "Main: Default config adapter is not *gcfg.AdapterFile, path not set.")
    }

	rootCmd := gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "yuncms server starting...")
			// Placeholder for actual server start, e.g., g.Server().Run()
            // For now, just a message. We'll implement server start in a later task.
			g.Log().Info(ctx, "yuncms server placeholder. Implement server start later.")
			return nil
		},
	}
	// Add the migrate command to the root command.
	rootCmd.AddCommand(&cmd.Migrate)
	// Execute the root command.
	rootCmd.Run(ctx)
}
