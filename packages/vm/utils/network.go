package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func DefaultNetworkInterface() (string, error) {
	if iface := os.Getenv("HOST_NET_IFACE"); iface != "" {
		return iface, nil
	}

	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", err
	}

	return parseDefaultNetworkInterface(string(out))
}

func parseDefaultNetworkInterface(route string) (string, error) {
	fields := strings.Fields(route)
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "dev" {
			return fields[i+1], nil
		}
	}

	return "", fmt.Errorf("default network interface not found")
}
