package main

import (
	_ "yuncms/internal/packed"
	_ "yuncms/internal/logic"
	_ "yuncms/internal/adapter"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2" // MySQL driver
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"   // Redis driver

	"github.com/gogf/gf/v2/os/gctx"
	"yuncms/internal/cmd"
)

func main() {
	// This is a basic main function.
	// It will be expanded to initialize routers, load configuration, etc.
	// For now, it just calls a placeholder command server.
	// The actual server startup logic will be added in subsequent steps.
	cmd.Main.Run(gctx.New())
}
