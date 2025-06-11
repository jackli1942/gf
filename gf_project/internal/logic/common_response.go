package logic

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	// "github.com/gogf/gf/v2/frame/g" // Not used directly in this file
)

// JsonRes is a standard JSON response structure.
type JsonRes struct {
	Code    int         `json:"code"`    // Error code, 0 for success (use gcode.CodeOK.Code())
	Message string      `json:"message"` // Message
	Data    interface{} `json:"data"`    // Data payload
	// TraceId string      `json:"traceId,omitempty"` // Optional: for tracing requests
}

// Respond is a helper function to send a JSON response.
func Respond(r *ghttp.Request, code int, message string, data ...interface{}) {
	var resData interface{}
	if len(data) > 0 {
		resData = data[0]
	}
	// traceId := r.Context().Value("TraceID") // Example if TraceID is in context via middleware
	r.Response.WriteJsonExit(JsonRes{ // Using WriteJsonExit to also terminate handler execution
		Code:    code,
		Message: message,
		Data:    resData,
		// TraceId: gconv.String(traceId),
	})
}

// Success sends a standard success response.
func Success(r *ghttp.Request, data ...interface{}) {
	var payload interface{}
	msg := "Operation successful" // Default success message
	if len(data) > 0 {
		payload = data[0]
	}
	// Allow providing a custom success message as the second argument if payload is also provided
	if len(data) > 1 {
		if m, ok := data[1].(string); ok {
			msg = m
		}
	}
	Respond(r, gcode.CodeOK.Code(), msg, payload)
}

// Error sends a standard error response.
// It automatically extracts code and message from gerror if err is a *gerror.Error.
func Error(r *ghttp.Request, err error, customMessage ...string) {
	code := gcode.CodeInternal // Default error code
	msg := "Operation failed"  // Default error message

	if err != nil {
		gErr := gerror.Cause(err)
		if gErr != nil {
			// Try to get code from gerror
			errCode := gerror.Code(gErr)
			if errCode != gcode.CodeNil { // CodeNil means no specific code was set in gerror
				code = errCode
			}
		}
		msg = err.Error() // Use the error's message by default
	}

	// Allow overriding the message with a custom one
	if len(customMessage) > 0 {
		msg = customMessage[0]
	}
	Respond(r, code.Code(), msg) // Use .Code() to get the integer value
}
