package utils

import (
	"net"

	"github.com/rs/zerolog"
)

func GetLocalIP(logger *zerolog.Logger) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error().Err(err).Msg("error getting local IP addresses")

		return ""
	}

	for _, addr := range addrs {
		// Check if the address is an IP address (not a loopback)
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String() // Return first non-loopback IPv4 address
			}
		}
	}

	return ""
}
