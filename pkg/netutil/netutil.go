package netutil

import (
	"fmt"
	"net"
)

// CIDR return subnet of given string.
func CIDR(s string) (*net.IPNet, error) {
	_, subnet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("parce cidr: %w", err)
	}
	return subnet, nil
}

// GetLocalIp returns local ip address.
func GetLocalIp() (net.IP, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("get interface addr: %w", err)
	}

	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return nil, nil
}
