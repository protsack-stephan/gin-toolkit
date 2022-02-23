package httpmw

// CognitoUser cogntito user entity.
type CognitoUser struct {
	Username string
	Groups   []string
	Lookup   map[string]struct{}
}

// GetUsername get private username property
func (cu *CognitoUser) GetUsername() string {
	return cu.Username
}

// SetUsername sets username for user
func (cu *CognitoUser) SetUsername(username string) {
	cu.Username = username
}

// GetGroups get user groups private property
func (cu *CognitoUser) GetGroups() []string {
	return cu.Groups
}

// Set cognito groups for user
func (cu *CognitoUser) SetGroups(groups []string) {
	cu.Lookup = make(map[string]struct{})

	for _, group := range groups {
		cu.Lookup[group] = struct{}{}
	}

	cu.Groups = groups
}

// Checks if user groups contains passed groups
func (cu *CognitoUser) IsInGroup(groups ...string) bool {
	for _, group := range groups {
		if _, ok := cu.Lookup[group]; ok {
			return true
		}
	}

	return false
}
