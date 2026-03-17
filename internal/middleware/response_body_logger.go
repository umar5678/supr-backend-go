package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func DevelopmentLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipLogging(c.Request.URL.Path) {
			c.Next()
			return
		}

		startTime := time.Now()

		logRequest(c)

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		duration := time.Since(startTime)
		logResponse(c, blw.body.String(), duration)
	}
}

func logRequest(c *gin.Context) {
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	query := c.Request.URL.RawQuery
	if query != "" {
		query = "?" + query
	}

	fmt.Printf("%s %s%s | %s | %s",
		c.Request.Method,
		c.Request.URL.Path,
		query,
		c.ClientIP(),
		maskAuth(c.GetHeader("Authorization")),
	)

	if len(bodyBytes) > 0 {
		var bodyData interface{}
		if err := json.Unmarshal(bodyBytes, &bodyData); err == nil {
			if isSimpleBody(bodyData) {
				fmt.Printf(" | %s", formatInline(bodyData))
			} else {
				fmt.Printf("\n   📦 %s", formatMultiline(bodyData, "   "))
			}
		}
	}
	fmt.Println()
}

func logResponse(c *gin.Context, responseBody string, duration time.Duration) {
	statusCode := c.Writer.Status()
	statusEmoji := getStatusEmoji(statusCode)

	var responseData interface{}
	if len(responseBody) > 0 {
		json.Unmarshal([]byte(responseBody), &responseData)
	}

	responseData = cleanResponseData(responseData)

	fmt.Printf("%s %d | %dms | %db",
		statusEmoji,
		statusCode,
		duration.Milliseconds(),
		len(responseBody),
	)

	if responseData != nil && hasDataToShow(responseData) {
		if isSimpleBody(responseData) {
			fmt.Printf(" | %s", formatInline(responseData))
		} else {
			fmt.Printf("\n   📦 %s", formatMultiline(responseData, "   "))
		}
	}
	fmt.Printf("\n\n")
}

func cleanResponseData(data interface{}) interface{} {
	if respMap, ok := data.(map[string]interface{}); ok {
		cleaned := make(map[string]interface{})
		for k, v := range respMap {
			if k == "meta" || k == "requestId" || k == "timestamp" || k == "version" {
				continue
			}
			cleaned[k] = v
		}

		if len(cleaned) <= 2 {
			hasOnlyBasic := true
			for key := range cleaned {
				if key != "success" && key != "message" {
					hasOnlyBasic = false
					break
				}
			}
			if hasOnlyBasic {
				return nil
			}
		}

		return cleaned
	}
	return data
}

func isSimpleBody(data interface{}) bool {
	m, ok := data.(map[string]interface{})
	if !ok {
		return true
	}

	if len(m) > 3 {
		return false
	}

	for _, v := range m {
		switch v.(type) {
		case map[string]interface{}, []interface{}:
			return false
		}
	}

	return true
}

func formatInline(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		var parts []string
		for key, val := range v {
			parts = append(parts, fmt.Sprintf("%s:%v", key, formatValue(val)))
		}
		return strings.Join(parts, ", ")
	case []interface{}:
		return fmt.Sprintf("[%d items]", len(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatMultiline(data interface{}, indent string) string {
	var result strings.Builder

	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			result.WriteString(fmt.Sprintf("\n%s   %s: ", indent, key))

			switch subVal := val.(type) {
			case map[string]interface{}:
				result.WriteString(formatMultiline(subVal, indent+"   "))
			case []interface{}:
				result.WriteString(fmt.Sprintf("[%d items]", len(subVal)))
				for i, item := range subVal {
					result.WriteString(fmt.Sprintf("\n%s      [%d] ", indent, i))
					if itemMap, ok := item.(map[string]interface{}); ok {
						result.WriteString(formatMultiline(itemMap, indent+"      "))
					} else {
						result.WriteString(fmt.Sprintf("%v", item))
					}
				}
			default:
				result.WriteString(fmt.Sprintf("%v", formatValue(val)))
			}
		}
	case []interface{}:
		result.WriteString(fmt.Sprintf("[%d items]", len(v)))
		for i, item := range v {
			result.WriteString(fmt.Sprintf("\n%s   [%d] ", indent, i))
			if itemMap, ok := item.(map[string]interface{}); ok {
				result.WriteString(formatMultiline(itemMap, indent+"   "))
			} else {
				result.WriteString(fmt.Sprintf("%v", item))
			}
		}
	default:
		result.WriteString(fmt.Sprintf("%v", v))
	}

	return result.String()
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		if len(v) > 40 {
			return fmt.Sprintf("\"%s...\"", v[:37])
		}
		return fmt.Sprintf("\"%s\"", v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.2f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func hasDataToShow(data interface{}) bool {
	if data == nil {
		return false
	}

	if m, ok := data.(map[string]interface{}); ok {
		if len(m) == 0 {
			return false
		}

		if dataField, exists := m["data"]; exists {
			if dataField == nil {
				return false
			}
			if dataMap, ok := dataField.(map[string]interface{}); ok && len(dataMap) == 0 {
				return false
			}
			if dataArr, ok := dataField.([]interface{}); ok && len(dataArr) == 0 {
				return false
			}
		}

		return true
	}

	if arr, ok := data.([]interface{}); ok {
		return len(arr) > 0
	}

	return true
}

func maskAuth(auth string) string {
	if auth == "" {
		return "NoAuth"
	}
	if strings.HasPrefix(auth, "Bearer ") {
		token := auth[7:]
		if len(token) > 8 {
			return "Auth:***" + token[len(token)-4:]
		}
	}
	return "Auth:***"
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

func getStatusEmoji(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return ""
	case statusCode >= 300 && statusCode < 400:
		return ""
	case statusCode >= 400 && statusCode < 500:
		return ""
	case statusCode >= 500:
		return ""
	default:
		return ""
	}
}
