package httpmodules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	assert.NotNil(t, new(Route))
}
