package httpmw

// CognitoUser cogntito user entity.
type CognitoUser struct {
	Username string
	Groups   map[string]struct{}
}

// SetUsername sets username for user
func (cu *CognitoUser) SetUsername(username string) {
	cu.Username = username
}

// Set cognito groups for user
func (cu *CognitoUser) SetGroups(groups []interface{}) {
	groupsMap := make(map[string]struct{})

	for _, group := range groups {
		groupsMap[group.(string)] = struct{}{}
	}

	cu.Groups = groupsMap
}

// Checks if user groups contains passed groups
func (cu *CognitoUser) IsInGroup(groups []string) bool {
	for _, group := range groups {
		if _, ok := cu.Groups[group]; ok {
			return true
		}
	}

	return false
}
