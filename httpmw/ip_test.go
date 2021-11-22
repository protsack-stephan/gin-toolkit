package httpmw

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestCheckIPSuccess(t *testing.T) {
	assert := assert.New(t)
	ipRange := ipBand{net.ParseIP("192.168.10.1"), net.ParseIP("192.168.10.10")}

	assert.True(checkIP(ipRange, "192.168.10.2"))
}

func TestCheckIPFails(t *testing.T) {
	assert := assert.New(t)
	ipRange := ipBand{net.ParseIP("192.168.10.1"), net.ParseIP("192.168.10.10")}

	assert.False(checkIP(ipRange, "192.168.10.22"))
}
