package main

import (
	"context"
	"yuncms/internal/cmd"
	_ "yuncms/internal/boot"
	"yuncms/internal/router"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
    "github.com/gogf/gf/v2/os/gproc" // Added for gproc.Pid()
    _ "github.com/gogf/gf/contrib/drivers/mysql/v2" // Ensure DB driver is imported
)

func main() {
    ctx := context.Background()
    // Configuration Path Setup (ensure this is robust as per previous steps)
    configPath := "manifest/config"
    // Try absolute path for subtask environment if relative doesn't exist
    if !gfile.Exists(configPath) {
         altPath := "/app/manifest/config"
         if gfile.Exists(altPath) {
             configPath = altPath
         } else {
              // Fallback to a path relative to where main.go might be if not at root
              // This case might be less common for typical project structures but added for robustness
              altPath = "./manifest/config"
              if gfile.Exists(altPath) {
                  configPath = altPath
              } else {
                  glog.Fatal(ctx, "Main: Config directory 'manifest/config' not found at expected paths (CWD, /app, ./).")
                  return // Exit if config is critical and not found
              }
         }
    }
    // Set the path for the default configuration adapter
    // Ensure this runs before any g.Cfg().MustGet calls in command Funcs or other inits that depend on it
    if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
        // It's usually safe to set the path here, as main.go's init phase (if any) or direct calls
        // are the earliest point for application-wide config.
        // The boot.go init also attempts this, this ensures it if boot.go didn't run or failed.
        adapter.SetPath(configPath)
        glog.Debug(ctx, "Main: Configuration path set to:", configPath)
    } else {
        glog.Warning(ctx, "Main: Default config adapter is not *gcfg.AdapterFile. Path may not be correctly set if default search paths fail.")
    }

    // Define the main server command
	mainCmd := &gcmd.Command{ // Changed to pointer to use AddCommand method correctly
		Name:  "main",
		Usage: "main (no args) to start server, or [subcommand]", // Updated Usage
		Brief: "start http server for yuncms or run subcommands",    // Updated Brief
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server() // Get default server instance
            router.BindCentralRouter(s) // Bind all application routes

			serverAddr := g.Cfg().MustGet(ctx, "server.address", ":8080").String()
            s.SetAddr(serverAddr) // Set server address
            g.Log().Infof(ctx, "yuncms server starting at http://127.0.0.1%s (PID: %d)", serverAddr, gproc.Pid())
			s.Run() // Start the server
			return nil
		},
	}
    // Add other commands like migrate
    mainCmd.AddCommand(&cmd.Migrate)

    // Execute the main command
    mainCmd.Run(ctx)
}
