package custommiddleware

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

// TraceID adds X-Request-ID: <trace-id> to response headers.
// It reads the traceID from the context so should be .Used after otelecho.Middleware
func TraceID() echo.MiddlewareFunc {
	// "inspired" by https://github.com/labstack/echo/blob/42f07ed880400b8bb80906dfec8138c572748ae8/middleware/request_id.go
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			traceID := trace.SpanContextFromContext(req.Context()).TraceID().String()
			res.Header().Set(echo.HeaderXRequestID, traceID)
			return next(c)
		}
	}
}
