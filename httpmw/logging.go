package httpmw

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type logEntry struct {
	ResponseTime string        `json:"response_time"`
	Status       int           `json:"status"`
	Latency      time.Duration `json:"latency"`
	IP           string        `json:"ip"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Username     string        `json:"username,omitempty"`
	UserGroups   []string      `json:"user_groups,omitempty"`
	BodySize     int           `json:"body_size"`
}

// LogFormatter builds a logging entry in JSON format containing these fields:
// - Unix timestamp of the request time
// - Client's IP address
// - Accessed API endpoint/path
// - Request method (GET, POST, PUT, PATCH, DELETE)
// - Request's response status code
// - Latency of the request in milliseconds
// - Response body size in bytes
//
// This function will also look into the gin's context for a user instance.
// If a CognitoUser instance is found, the formatter will also include the following fields:
// - Username
// - User associated group(s)
func LogFormatter(p gin.LogFormatterParams) string {
	entry := &logEntry{
		ResponseTime: p.TimeStamp.Format(time.RFC3339),
		Status:       p.StatusCode,
		Latency:      p.Latency,
		IP:           p.ClientIP,
		Method:       p.Method,
		Path:         p.Path,
		BodySize:     p.BodySize,
	}

	var user *CognitoUser

	if val, ok := p.Keys["user"]; ok && val != nil {
		user, _ = val.(*CognitoUser)
	}

	if user != nil {
		entry.Username = user.Username
		entry.UserGroups = user.GetGroups()
	}

	b, _ := json.Marshal(entry)
	return fmt.Sprintln(string(b))
}
