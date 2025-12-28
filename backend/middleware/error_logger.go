package middleware

import (
	"fmt"
	"movie-ticket-backend/services"

	"github.com/gin-gonic/gin"
)

// ErrorLogger is a middleware that captures all API errors (status >= 400)
// and records them into the Audit Log as SYSTEM_ERROR.
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Check if we have an error status
		status := c.Writer.Status()
		if status >= 400 {
			userID, _ := c.Get("userID")
			uidStr := ""
			if userID != nil {
				uidStr = userID.(string)
			}

			// Context for the log
			details := map[string]interface{}{
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"status_code": status,
				"ip":          c.ClientIP(),
				"user_agent":  c.Request.UserAgent(),
			}

			if c.Request.URL.RawQuery != "" {
				details["query"] = c.Request.URL.RawQuery
			}

			// Check if any errors were specifically attached to the context
			if len(c.Errors) > 0 {
				details["gin_errors"] = c.Errors.Errors()

				// Use the last error for the main error logging
				lastErr := c.Errors.Last().Err
				services.LogError("SYSTEM_ERROR", uidStr, lastErr, details)
			} else {
				// No specific error object, but status is failure
				// (e.g. c.JSON(400, ...) without c.Error(err))
				services.LogInfo("SYSTEM_ERROR", uidStr, details)
			}

			fmt.Printf(" [AUDIT] Logged SYSTEM_ERROR: %s %s -> Status %d\n", c.Request.Method, c.Request.URL.Path, status)
		}
	}
}
