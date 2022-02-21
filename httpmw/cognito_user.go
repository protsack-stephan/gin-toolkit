package httpmw

// CognitoUser cogntito user entity.
type CognitoUser struct {
	username string
	groups   []string
	lookup   map[string]struct{}
}

// GetUsername get private username property
func (cu *CognitoUser) GetUsername() string {
	return cu.username
}

// SetUsername sets username for user
func (cu *CognitoUser) SetUsername(username string) {
	cu.username = username
}

// GetGroups get user groups private property
func (cu *CognitoUser) GetGroups() []string {
	return cu.groups
}

// Set cognito groups for user
func (cu *CognitoUser) SetGroups(groups []string) {
	cu.lookup = make(map[string]struct{})

	for _, group := range groups {
		cu.lookup[group] = struct{}{}
	}

	cu.groups = groups
}

// Checks if user groups contains passed groups
func (cu *CognitoUser) IsInGroup(groups ...string) bool {
	for _, group := range groups {
		if _, ok := cu.lookup[group]; ok {
			return true
		}
	}

	return false
}
