package httpmw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCognitoUser(t *testing.T) {
	assert := assert.New(t)

	t.Run("test groups", func(t *testing.T) {
		groups := []string{"main", "secondary"}

		user := new(CognitoUser)
		user.SetGroups(groups)

		for _, group := range groups {
			assert.True(user.IsInGroup(group))
		}

		assert.Equal(user.GetGroups(), groups)
	})

	t.Run("test username", func(t *testing.T) {
		username := "john.doe"

		user := new(CognitoUser)
		user.SetUsername(username)

		assert.Equal(username, user.GetUsername())
	})
}
