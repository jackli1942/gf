package middleware

import "github.com/gogf/gf/v2/net/ghttp"

// Recovery is a middleware handler for recovering from panics.
func Recovery(r *ghttp.Request) {
	defer func() {
		if err := recover(); err != nil {
			// Log the panic, return a 500 error, etc.
			r.Response.WriteStatusExit(ghttp.StatusInternalServerError, "Internal Server Error")
		}
	}()
	r.Middleware.Next()
}
