package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	assert.NotNil(t, new(Module))
}