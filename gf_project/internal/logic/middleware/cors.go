package middleware

import "github.com/gogf/gf/v2/net/ghttp"

// CORS allows Cross-Origin Resource Sharing.
func CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}
