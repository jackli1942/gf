package middleware

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"time"
)

// Logger logs the request information.
func Logger(r *ghttp.Request) {
	r.Middleware.Next()
	g.Log().Infof(r.Context(), "%s %s %d %s", r.Method, r.URL.Path, r.Response.Status, time.Since(r.EnterTime))
}
