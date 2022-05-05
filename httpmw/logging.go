package httpmw

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type logEntry struct {
	RequestTime string   `json:"request_time"`
	Ip          string   `json:"ip"`
	Path        string   `json:"path"`
	Status      int      `json:"status"`
	Username    string   `json:"username"`
	UserGroups  []string `json:"user_groups"`
}

// LogFormatter builds a logging entry in JSON format containing these fields:
// - Unix timestamp of the request time
// - Client's IP address
// - Accessed API endpoint/path
// - Request's response status code
//
// This function will also look into the gin's context for a user instance.
// If a CognitoUser instance is found, the formatter will also include the following fields:
// - Username
// - User associated group(s)
func LogFormatter(params gin.LogFormatterParams) string {
	var user *CognitoUser

	entry := &logEntry{
		RequestTime: params.TimeStamp.Format(time.RFC3339),
		Ip:          params.ClientIP,
		Path:        params.Path,
		Status:      params.StatusCode,
	}

	if val, ok := params.Keys["user"]; ok && val != nil {
		user, _ = val.(*CognitoUser)
	}

	if user != nil {
		entry.Username = user.Username
		entry.UserGroups = user.GetGroups()
	}

	b, _ := json.Marshal(entry)
	return fmt.Sprintln(string(b))
}
