package httpmw

import (
	"bytes"
	"net"
	"strings"
)

type ipBand struct {
	start net.IP
	end   net.IP
}

func checkIP(ipRange ipBand, ip string) bool {
	input := net.ParseIP(ip)

	return bytes.Compare(input, ipRange.start) >= 0 && bytes.Compare(input, ipRange.end) <= 0
}

func getIpRanges(ipRange string) []ipBand {
	var ipRanges []ipBand

	for _, ipRange := range strings.Split(ipRange, ",") {
		if len(ipRange) > 0 {
			ips := strings.Split(ipRange, "-")
			ipRanges = append(
				ipRanges,
				ipBand{
					net.ParseIP(ips[0]),
					net.ParseIP(ips[1]),
				},
			)
		}
	}

	return ipRanges
}
